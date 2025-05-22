package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/carlosrabelo/tatuscan/tools/tools/internal/apiclient"
	"github.com/carlosrabelo/tatuscan/tools/tools/internal/deleteolder"
	"github.com/carlosrabelo/tatuscan/tools/tools/internal/i18n"
)

func main() {
	if err := run(); err != nil {
		slog.Error("delete-older failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	apiBase := flag.String("api-base", "", "API base URL (default: TATUSCAN_URL or localhost)")
	logLevel := flag.String("log-level", "INFO", "log level")
	dryRun := flag.Bool("dry-run", false, "print deletes without calling the API")
	flag.Parse()
	_ = apiclient.LoadDotEnv(".env")
	if err := apiclient.SetLogLevel(*logLevel); err != nil {
		return err
	}
	cat := i18n.FromEnv()
	c := apiclient.New(apiclient.ResolveAPIBase(*apiBase), apiclient.ResolveToken())
	res, err := deleteolder.Process(context.Background(), c, *dryRun, os.Stdout, cat)
	if err != nil {
		return err
	}
	fmt.Println(cat.T("delete.hosts", res.DuplicateHosts))
	if *dryRun {
		fmt.Println(cat.T("delete.dry", res.Marked))
	} else {
		fmt.Println(cat.T("delete.removed", res.Removed))
	}
	for _, line := range res.Errors {
		fmt.Printf(" - %s\n", line)
	}
	return nil
}
