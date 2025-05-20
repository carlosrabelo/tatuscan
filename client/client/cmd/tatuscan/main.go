//go:build windows || linux || darwin

package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/carlosrabelo/tatuscan/client/internal"
	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
)

func main() {
	log := setupLogger()
	logLevel := flag.String("l", "", "Set log level (debug, info, warn, error, fatal)")
	daemonMode := flag.Bool("d", false, "Run in daemon mode (repeat collection in cycles)")
	intervalFlag := flag.String("interval", "", "Collection interval (ex.: 60s, 2m). Env: TATUSCAN_INTERVAL")
	flag.Parse()
	_ = internal.LoadDotEnv(".env")
	level := *logLevel
	if level == "" {
		level = os.Getenv("TATUSCAN_LOG_LEVEL")
	}
	applyLogLevel(log, level)

	internal.EnsureSingleInstance()
	serverURL := internal.GetServerURL()
	interval := internal.ResolveInterval(*intervalFlag)
	s := mustService(log, serverURL, interval)

	if flag.NArg() > 0 {
		controlService(log, s, flag.Args())
		return
	}
	internal.RunProgram(s, serverURL, interval, *daemonMode)
}

func setupLogger() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{DisableColors: true, ForceColors: false})
	log.SetOutput(os.Stdout)
	internal.SetLogger(log)
	return log
}

func applyLogLevel(log *logrus.Logger, level string) {
	if level == "" {
		log.SetLevel(logrus.WarnLevel)
		return
	}
	parsed, err := logrus.ParseLevel(strings.ToLower(level))
	if err != nil {
		log.Fatalf("Invalid log level: %s. Use debug, info, warn, error or fatal", level)
	}
	log.SetLevel(parsed)
}

func mustService(log *logrus.Logger, serverURL string, interval time.Duration) service.Service {
	svcConfig := &service.Config{
		Name:        "TatuScanAgent",
		DisplayName: "TatuScan Agent",
		Description: "TatuScan monitoring agent",
	}
	prg := &internal.Program{ServerURL: serverURL, Interval: interval}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatalf("Error to create service: %v", err)
	}
	return s
}

func controlService(log *logrus.Logger, s service.Service, args []string) {
	for _, arg := range args {
		if err := service.Control(s, arg); err != nil {
			log.Fatalf("Error to control service: %v", err)
		}
	}
}
