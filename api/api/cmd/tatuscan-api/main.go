package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/carlosrabelo/tatuscan/api/api/internal/config"
	"github.com/carlosrabelo/tatuscan/api/api/internal/httpapi"
	"github.com/carlosrabelo/tatuscan/api/api/internal/service"
	"github.com/carlosrabelo/tatuscan/api/api/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	logger := config.NewLogger(cfg.LogLevel)

	st, err := store.Open(cfg.DBPath)
	if err != nil {
		logger.Error("database", "err", err)
		os.Exit(1)
	}
	defer st.Close()

	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		logger.Error("timezone", "err", err)
		os.Exit(1)
	}
	st.SetLocation(loc)

	// Transitional: Flask/SQLAlchemy datetime → RFC3339Nano (naive = app TZ).
	// Skip with TATUSCAN_SKIP_LEGACY_DATETIME_MIGRATE=1 when transition is done.
	if os.Getenv("TATUSCAN_SKIP_LEGACY_DATETIME_MIGRATE") == "" {
		migCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		res, err := st.NormalizeLegacyDatetimes(migCtx, loc, logger)
		cancel()
		if err != nil {
			logger.Error("legacy datetime migration", "err", err)
			os.Exit(1)
		}
		if res.Updated > 0 || res.Skipped > 0 {
			logger.Info("legacy datetime migration completed",
				"rows_updated", res.Updated, "rows_skipped", res.Skipped)
		}
	}

	svc, err := service.New(st, cfg.Timezone)
	if err != nil {
		logger.Error("service", "err", err)
		os.Exit(1)
	}

	api := httpapi.New(svc, st, logger, httpapi.Options{
		APIToken:     cfg.APIToken,
		OfflineAfter: cfg.OfflineAfter,
		DefaultLang:  cfg.DefaultLang,
	})
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go listen(srv, logger, cfg)
	awaitShutdown(srv, logger)
}

func listen(srv *http.Server, logger *slog.Logger, cfg config.Config) {
	logger.Info("tatuscan-api listening", "addr", srv.Addr, "db", cfg.DBPath, "tz", cfg.Timezone,
		"auth", cfg.APIToken != "", "offline_after", cfg.OfflineAfter.String())
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
	logger.Info("tatuscan-api stopped")
}
