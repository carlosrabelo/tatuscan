package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/carlosrabelo/tatuscan/tools/tools/internal/addmanual"
	"github.com/carlosrabelo/tatuscan/tools/tools/internal/apiclient"
	"github.com/carlosrabelo/tatuscan/tools/tools/internal/i18n"
)

// Set by .make/build.sh via -ldflags "-X main.version=…".
var version = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("add-manual-inventory failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	e := addmanual.Entry{OS: "Chrome OS", IP: "0.0.0.0"}
	apiBase := flag.String("api-base", "", "API base URL (default: TATUSCAN_URL or localhost)")
	logLevel := flag.String("log-level", "INFO", "log level")
	dryRun := flag.Bool("dry-run", false, "print the payload without sending")
	flag.StringVar(&e.Hostname, "hostname", "", "hostname to register")
	flag.StringVar(&e.OS, "os", e.OS, "operating system name")
	flag.StringVar(&e.OSVersion, "os-version", "", "operating system version")
	flag.StringVar(&e.MachineID, "machine-id", "", "fixed machine_id (otherwise hashed from hostname)")
	flag.StringVar(&e.IP, "ip", e.IP, "reported IP")
	flag.Float64Var(&e.CPUPercent, "cpu-percent", 0, "CPU load")
	flag.StringVar(&e.Activation, "activation", "", "optional computer_activation date")
	flag.StringVar(&e.Salt, "salt", "", "extra salt for generated machine_id")
	flag.Parse()
	_ = apiclient.LoadDotEnv(".env")
	if e.Hostname == "" {
		flag.Usage()
		return fmt.Errorf("%s", i18n.FromEnv().T("add.hostname"))
	}
	if err := apiclient.SetLogLevel(*logLevel); err != nil {
		return err
	}
	slog.Debug("build version", "version", version)
	c := apiclient.New(apiclient.ResolveAPIBase(*apiBase), apiclient.ResolveToken())
	return addmanual.Send(context.Background(), c, e, *dryRun)
}
