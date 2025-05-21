package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/carlosrabelo/tatuscan/web/web/internal/apiclient"
	"github.com/carlosrabelo/tatuscan/web/web/internal/config"
	"github.com/carlosrabelo/tatuscan/web/web/internal/httpui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	logger := config.NewLogger(cfg.LogLevel)

	ui := httpui.New(apiclient.New(cfg.APIURL), logger, httpui.Options{
		OfflineAfter: cfg.OfflineAfter,
		DefaultLang:  cfg.DefaultLang,
	})
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           ui.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go listen(srv, logger, cfg)
	awaitShutdown(srv, logger)
}

func listen(srv *http.Server, logger *slog.Logger, cfg config.Config) {
	logger.Info("tatuscan-web listening", "addr", srv.Addr, "api", cfg.APIURL, "offline_after", cfg.OfflineAfter.String())
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("listen", "err", err)
		os.Exit(1)
	}
}

func awaitShutdown(srv *http.Server, logger *slog.Logger) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	logger.Info("tatuscan-web stopped")
}
