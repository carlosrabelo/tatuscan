//go:build windows || linux || darwin

package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/kardianos/service"
)

const (
	DefaultInterval    = 60 * time.Second
	DefaultServerURL   = "http://127.0.0.1:8040"
	EnvServerURL       = "TATUSCAN_URL"
	EnvCollectInterval = "TATUSCAN_INTERVAL"
	EnvAPIToken        = "TATUSCAN_API_TOKEN"
	AgentVersion       = "0.0.1"
)

// GetServerURL retrieves the base server URL from the environment variable.
// When unset (and no .env), defaults to DefaultServerURL.
func GetServerURL() string {
	Log.Debug("Getting ServerURL from environment variable")
	base := strings.TrimSpace(os.Getenv(EnvServerURL))
	if base == "" {
		base = DefaultServerURL
		Log.Debugf("%s not set; using default %s", EnvServerURL, DefaultServerURL)
	}
	base = strings.TrimRight(base, "/")
	url := base + "/api/machines"
	Log.Debugf("Final ServerURL: %s", url)
	return url
}

// ResolveInterval resolves the collection interval from flag, env, or default.
func ResolveInterval(flagVal string) time.Duration {
	if flagVal != "" {
		if d, err := time.ParseDuration(flagVal); err == nil {
			return d
		}
		Log.Fatalf("Invalid value for -interval: %s", flagVal)
	}
	if env := strings.TrimSpace(os.Getenv(EnvCollectInterval)); env != "" {
		if d, err := time.ParseDuration(env); err == nil {
			return d
		}
		Log.Fatalf("Invalid value for %s: %s", EnvCollectInterval, env)
	}
	return DefaultInterval
}

// SendData sends collected machine data to the server.
func SendData(info MachineInfo, serverURL string) error {
	Log.Info("Sending data to server")
	data, err := json.Marshal(info)
	if err != nil {
		Log.Errorf("Failed to serialize data: %v", err)
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPost, serverURL, bytes.NewBuffer(data))
	if err != nil {
		Log.Errorf("Failed to create HTTP request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("TatuScan/%s (%s)", AgentVersion, runtime.GOOS))
	if token := strings.TrimSpace(os.Getenv(EnvAPIToken)); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		Log.Errorf("Failed to send data: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		err := fmt.Errorf("server returned status: %d", resp.StatusCode)
		Log.Error(err)
		return err
	}

	Log.Info("Data sent successfully")
	return nil
}

// RunAgent runs the main collection loop until the context is cancelled.
func RunAgent(ctx context.Context, serverURL string, interval time.Duration) {
	Log.Info("Starting agent loop")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	doCycle := func() {
		Log.Debug("Starting collection and send cycle")
		info, err := CollectData()
		if err != nil {
			Log.Errorf("Failed to collect data: %v", err)
			return
		}
		if err := SendData(info, serverURL); err != nil {
			Log.Errorf("Failed to send data: %v", err)
			return
		}
		Log.Debug("Cycle completed")
	}

	doCycle()

	for {
		select {
		case <-ctx.Done():
			Log.Info("Stopping agent")
			return
		case <-ticker.C:
			doCycle()
		}
	}
}

// Program implements the service.Interface for running as a system service.
type Program struct {
	ServerURL string
	Interval  time.Duration
	cancel    context.CancelFunc
}

func (p *Program) Start(s service.Service) error {
	Log.Debugf("Starting TatuScan agent as service on OS: %s", runtime.GOOS)
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	go RunAgent(ctx, p.ServerURL, p.Interval)
	return nil
}

func (p *Program) Stop(s service.Service) error {
	Log.Debug("Stopping TatuScan agent")
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

// RunProgram handles interactive, daemon, and service execution modes.
func RunProgram(s service.Service, serverURL string, interval time.Duration, daemonMode bool) {
	if service.Interactive() {
		if daemonMode {
			ctx, cancel := context.WithCancel(context.Background())
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
			go func() { <-sigs; cancel() }()
			RunAgent(ctx, serverURL, interval)
		} else {
			Log.Info("Running single collection")
			info, err := CollectData()
			if err != nil {
				Log.Fatalf("Failed to collect data: %v", err)
			}
			if err := SendData(info, serverURL); err != nil {
				Log.Fatalf("Failed to send data: %v", err)
			}
		}
	} else {
		Log.Debug("Running as a service")
		if err := s.Run(); err != nil {
			Log.Fatalf("Failed to run service: %v", err)
		}
	}
}
