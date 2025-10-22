//go:build windows || linux || darwin

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/carlosrabelo/tatuscan/internal"
	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
)

const (
	defaultInterval    = 60 * time.Second
	envServerURL       = "TATUSCAN_URL"
	envCollectInterval = "TATUSCAN_INTERVAL"
	agentVersion       = "0.0.1"
)

var log *logrus.Logger // Logger global

// getServerURL retrieves the base server URL from environment variable
func getServerURL() string {
	log.Debug("Getting ServerURL from environment variable")
	base := os.Getenv(envServerURL)
	if base == "" {
		log.Fatalf("Environment variable %s not defined; is mandatory", envServerURL)
	}
	base = strings.TrimRight(base, "/")
	url := base + "/api/machines"
	log.Debugf("Final ServerURL: %s", url)
	return url
}

// sendData sends collected data to the server
func sendData(info internal.MachineInfo, serverURL string) error {
	log.Info("Sending data to server")
	data, err := json.Marshal(info)
	if err != nil {
		log.Errorf("Error to serialize data: %v", err)
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPost, serverURL, bytes.NewBuffer(data))
	if err != nil {
		log.Errorf("Error to create HTTP request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("TatuScan/%s (%s)", agentVersion, runtime.GOOS))

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Error to send data: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Accept 200 (OK) and 201 (Created) as valid responses
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		err := fmt.Errorf("server returned status: %d", resp.StatusCode)
		log.Error(err)
		return err
	}

	log.Info("Data sent successfully")
	return nil
}

// runAgent runs the main agent loop with context and ticker for immediate shutdown
func runAgent(ctx context.Context, serverURL string, interval time.Duration) {
	log.Info("Starting agent in repetitive mode (daemon or service)")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Execute one cycle immediately when starting
	doCycle := func() {
		log.Debug("Starting collection and send cycle")
		info, err := internal.CollectData()
		if err != nil {
			log.Errorf("Error to collect data: %v", err)
			return
		}
		if err := sendData(info, serverURL); err != nil {
			log.Errorf("Error to send data: %v", err)
			return
		}
		log.Debug("Cycle completed")
	}

	doCycle()

	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping agent by cancellation signal")
			return
		case <-ticker.C:
			doCycle()
		}
	}
}

// program implements the service interface
type program struct {
	serverURL string
	interval  time.Duration
	cancel    context.CancelFunc
}

func (p *program) Start(s service.Service) error {
	log.Debugf("Starting TatuScan agent as service on OS: %s", runtime.GOOS)
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	go runAgent(ctx, p.serverURL, p.interval)
	return nil
}

func (p *program) Stop(s service.Service) error {
	log.Debug("Stopping TatuScan agent")
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

func main() {
	// Configure the global logger
	log = logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		ForceColors:   false,
	})
	log.SetOutput(os.Stdout)

	// Configure logger for internal package
	internal.SetLogger(log)

	// Configure the flags
	logLevel := flag.String("l", "", "Set log level (debug, info, warn, error, fatal)")
	daemonMode := flag.Bool("d", false, "Run in daemon mode (repeat collection in cycles)")
	intervalFlag := flag.String("interval", "", "Collection interval (ex.: 60s, 2m). Env: TATUSCAN_INTERVAL")
	flag.Parse()

	// Set log level based on flag
	if *logLevel != "" {
		switch strings.ToLower(*logLevel) {
		case "debug":
			log.SetLevel(logrus.DebugLevel)
			log.Debug("Log level set as Debug")
		case "info":
			log.SetLevel(logrus.InfoLevel)
			log.Info("Log level set as Info")
		case "warn":
			log.SetLevel(logrus.WarnLevel)
			log.Warn("Log level set as Warn")
		case "error":
			log.SetLevel(logrus.ErrorLevel)
			log.Error("Log level set as Error")
		case "fatal":
			log.SetLevel(logrus.FatalLevel)
			log.Info("Log level set as Fatal")
		default:
			log.Fatalf("Invalid log level: %s. Use debug, info, warn, error or fatal", *logLevel)
		}
	} else {
		// Default level without -l: WarnLevel (shows only Warn, Error, Fatal)
		log.SetLevel(logrus.WarnLevel)
	}

	// Ensure single instance of the agent
	log.Debug("Checking single instance")
	internal.EnsureSingleInstance()

	// Get server URL (mandatory)
	log.Debug("Getting ServerURL")
	serverURL := getServerURL()

	// Determine collection interval (flag > env > default)
	interval := defaultInterval
	if *intervalFlag != "" {
		if d, err := time.ParseDuration(*intervalFlag); err == nil {
			interval = d
		} else {
			log.Fatalf("Invalid value for -interval: %v", err)
		}
	} else if env := strings.TrimSpace(os.Getenv(envCollectInterval)); env != "" {
		if d, err := time.ParseDuration(env); err == nil {
			interval = d
		} else {
			log.Fatalf("Invalid value for %s: %v", envCollectInterval, err)
		}
	}

	// Service configuration
	log.Debug("Configuring service")
	svcConfig := &service.Config{
		Name:        "TatuScanAgent",
		DisplayName: "TatuScan Agent",
		Description: "TatuScan monitoring agent",
	}

	// Create program for the service
	log.Debug("Creating service program")
	prg := &program{serverURL: serverURL, interval: interval}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatalf("Error to create service: %v", err)
	}

	// Manage service commands (ex.: install, start, stop)
	if flag.NArg() > 0 {
		for _, arg := range flag.Args() {
			log.Debugf("Managing service command: %s", arg)
			err = service.Control(s, arg)
			if err != nil {
				log.Fatalf("Error to control service: %v", err)
			}
		}
		return
	}

	// Execute the program
	if service.Interactive() {
		// Interactive mode: behavior depends on -d flag
		if *daemonMode {
			log.Info("Running in daemon mode (repetition activated via -d)")
			ctx, cancel := context.WithCancel(context.Background())
			// Capture signals for clean shutdown
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-sigs
				cancel()
			}()
			runAgent(ctx, serverURL, interval)
		} else {
			// Default behavior: execute single collection
			log.Info("Running single collection")
			info, err := internal.CollectData()
			if err != nil {
				log.Errorf("Error to collect data: %v", err)
				os.Exit(1)
			}
			if err := sendData(info, serverURL); err != nil {
				log.Errorf("Error to send data: %v", err)
				os.Exit(1)
			}
			log.Info("Collection completed successfully")
			os.Exit(0)
		}
	} else {
		// Service mode: run in cycles (Windows or Linux with systemd)
		log.Debug("Running as service (repetition automatically activated)")
		err = s.Run()
		if err != nil {
			log.Fatalf("Error to execute service: %v", err)
		}
	}
}
