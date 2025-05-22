package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/carlosrabelo/tatuscan/tools/tools/internal/apiclient"
	"github.com/carlosrabelo/tatuscan/tools/tools/internal/i18n"
	"github.com/carlosrabelo/tatuscan/tools/tools/internal/updateactivation"
)

// Set by .make/build.sh via -ldflags "-X main.version=…".
var version = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("update-activation failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	csvPath := flag.String("csv", "inventario.csv", "path to the report CSV")
	apiBase := flag.String("api-base", "", "API base URL (default: TATUSCAN_URL or localhost)")
	logLevel := flag.String("log-level", "INFO", "log level")
	flag.Parse()
	_ = apiclient.LoadDotEnv(".env")
	if err := apiclient.SetLogLevel(*logLevel); err != nil {
		return err
	}
	slog.Debug("build version", "version", version)
	cat := i18n.FromEnv()
	index, err := updateactivation.LoadReport(*csvPath)
	if err != nil {
		return err
	}
	fmt.Println(cat.T("update.loaded", len(index)))
	if len(index) == 0 {
		fmt.Println(cat.T("update.empty"))
		return nil
	}
	c := apiclient.New(apiclient.ResolveAPIBase(*apiBase), apiclient.ResolveToken())
	res, err := updateactivation.Process(context.Background(), c, index, cat)
	if err != nil {
		return err
	}
	fmt.Println(cat.T("update.hosts", res.TotalHosts))
	fmt.Println(cat.T("update.numbered", res.HostsWithNum))
	fmt.Println(cat.T("update.matches", res.Matches))
	fmt.Println(cat.T("update.updated", res.Updated))
	for _, line := range res.Errors {
		fmt.Printf(" - %s\n", line)
	}
	return nil
}
