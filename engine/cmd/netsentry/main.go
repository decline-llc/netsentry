package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/internal/api"
	"github.com/decline-llc/netsentry/internal/config"
	"github.com/decline-llc/netsentry/internal/pipeline"
	"github.com/decline-llc/netsentry/internal/receiver"
	"github.com/decline-llc/netsentry/internal/rule"
	nssignal "github.com/decline-llc/netsentry/internal/signal"
	"github.com/decline-llc/netsentry/internal/stats"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config.yaml")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}

	var logger *zap.Logger
	if cfg.Logging.Format == "json" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync() //nolint:errcheck

	logger.Info("NetSentry starting", zap.String("version", "0.1.0"))

	if err := os.MkdirAll(cfg.Engine.DBDir, 0o750); err != nil {
		logger.Fatal("create db dir", zap.Error(err))
	}

	ruleEngine := rule.NewEngine()
	if cfg.Engine.RulesSeedFile != "" {
		rules, err := rule.LoadFromFile(cfg.Engine.RulesSeedFile)
		if err != nil {
			logger.Warn("could not load seed rules", zap.Error(err))
		} else {
			if err := ruleEngine.Reload(rules); err != nil {
				logger.Warn("could not build rule state", zap.Error(err))
			} else {
				logger.Info("rules loaded", zap.Int("count", ruleEngine.RuleCount()))
			}
		}
	}

	metrics := stats.New()
	ctx, cancel := nssignal.WaitForShutdown()
	defer cancel()

	store, err := alert.Open(ctx, alert.Options{
		Path:              cfg.Engine.DBPath,
		Dir:               cfg.Engine.DBDir,
		DailyShard:        cfg.Engine.DBShardDaily,
		JournalMode:       cfg.Engine.DBJournalMode,
		BusyTimeoutMS:     cfg.Engine.DBBusyTimeout,
		AggregationWindow: time.Duration(cfg.Engine.AlertAggregationWindow) * time.Second,
		RetentionDays:     cfg.Engine.AlertRetentionDays,
	})
	if err != nil {
		logger.Fatal("open alert store", zap.Error(err))
	}
	defer store.Close() //nolint:errcheck

	recv := receiver.New(receiver.Config{
		Path:       cfg.Engine.UDSSocketPath,
		SocketMode: receiver.ParseSocketMode(cfg.Capture.UDSSocketMode),
		BufferSize: cfg.Engine.ChannelBufferSize,
		Stats:      metrics,
	}, logger)
	if err := recv.Start(ctx); err != nil {
		logger.Fatal("start uds receiver", zap.Error(err))
	}
	worker := pipeline.NewWorker(ruleEngine, store, logger, metrics)
	go worker.Run(ctx, recv.Packets())
	startHTTPServer(ctx, cfg.Engine.APIPort, cfg.Engine.RulesSeedFile, store, recv, ruleEngine, metrics, logger)

	logger.Info("engine ready — waiting for shutdown signal (SIGINT/SIGTERM)",
		zap.String("uds", cfg.Engine.UDSSocketPath),
		zap.Int("api_port", cfg.Engine.APIPort))
	<-ctx.Done()
	logger.Info("shutdown signal received, exiting")
}

func startHTTPServer(ctx context.Context, port int, rulesSeedFile string, store *alert.Store, recv *receiver.Receiver, ruleEngine *rule.Engine, metrics *stats.Stats, logger *zap.Logger) {
	if port == 0 {
		port = 8080
	}
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           api.NewServerWithOptions(store, recv, ruleEngine, metrics, api.Options{RulesSeedFile: rulesSeedFile}).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	go func() {
		logger.Info("http api started", zap.Int("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http api failed", zap.Error(err))
		}
	}()
}
