package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
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
	"github.com/decline-llc/netsentry/pkg/model"
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
		RecoveryLogPath:   cfg.Engine.AlertRecoveryLogPath,
	})
	if err != nil {
		logger.Fatal("open alert store", zap.Error(err))
	}
	defer store.Close() //nolint:errcheck

	recv := receiver.New(receiver.Config{
		Path:       cfg.Engine.UDSSocketPath,
		SocketMode: receiver.ParseSocketMode(cfg.Engine.UDSSocketMode),
		BufferSize: cfg.Engine.ChannelBufferSize,
		Stats:      metrics,
	}, logger)
	if err := recv.Start(ctx); err != nil {
		logger.Fatal("start uds receiver", zap.Error(err))
	}
	var suppressionRules []alert.Suppression
	if cfg.Engine.SuppressionsFile != "" {
		loaded, err := alert.LoadSuppressionsFromFile(cfg.Engine.SuppressionsFile)
		if err != nil {
			logger.Warn("could not load suppressions", zap.Error(err), zap.String("path", cfg.Engine.SuppressionsFile))
		} else {
			suppressionRules = loaded
			logger.Info("suppressions loaded", zap.Int("count", len(suppressionRules)))
		}
	}
	suppressions, err := alert.NewSuppressionManagerWithFile(suppressionRules, cfg.Engine.SuppressionsFile)
	if err != nil {
		logger.Fatal("create suppression manager", zap.Error(err))
	}
	worker := pipeline.NewWorker(ruleEngine, store, logger, metrics)
	worker.SetSuppressor(suppressions)
	if cfg.Engine.RedactSensitiveFields {
		worker.SetRedactor(alert.RedactSensitivePayloads)
	}
	workerDone := startPipelineWorkers(ctx, worker, recv.Packets(), cfg.Engine.WorkerCount)
	startHTTPServer(ctx, cfg.Engine, store, recv, ruleEngine, metrics, suppressions, logger)
	startPprofServer(ctx, cfg.Engine, logger)

	logger.Info("engine ready — waiting for shutdown signal (SIGINT/SIGTERM)",
		zap.String("uds", cfg.Engine.UDSSocketPath),
		zap.Int("api_port", cfg.Engine.APIPort))
	<-ctx.Done()
	logger.Info("shutdown signal received, stopping receiver")
	recv.Stop()
	recv.Wait()
	<-workerDone
	logger.Info("shutdown complete")
}

func startPipelineWorker(ctx context.Context, worker *pipeline.Worker, packets <-chan *model.PacketInfo) <-chan struct{} {
	return startPipelineWorkers(ctx, worker, packets, 1)
}

func startPipelineWorkers(ctx context.Context, worker *pipeline.Worker, packets <-chan *model.PacketInfo, count int) <-chan struct{} {
	if count < 1 {
		count = 1
	}
	done := make(chan struct{})
	workersDone := make(chan struct{}, count)
	for range count {
		go func() {
			defer func() { workersDone <- struct{}{} }()
			worker.Run(ctx, packets)
		}()
	}
	go func() {
		defer close(done)
		for range count {
			<-workersDone
		}
	}()
	return done
}

func startHTTPServer(ctx context.Context, engineCfg config.EngineConfig, store *alert.Store, recv *receiver.Receiver, ruleEngine *rule.Engine, metrics *stats.Stats, suppressions *alert.SuppressionManager, logger *zap.Logger) {
	srv := newHTTPServer(engineCfg, store, recv, ruleEngine, metrics, suppressions, logger)
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	go func() {
		logger.Info("http api started", zap.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http api failed", zap.Error(err))
		}
	}()
}

func newHTTPServer(engineCfg config.EngineConfig, store *alert.Store, recv *receiver.Receiver, ruleEngine *rule.Engine, metrics *stats.Stats, suppressions *alert.SuppressionManager, logger *zap.Logger) *http.Server {
	port := engineCfg.APIPort
	if port == 0 {
		port = 8080
	}
	host := engineCfg.APIListenHost
	if host == "" {
		host = "127.0.0.1"
	}
	return &http.Server{
		Addr: net.JoinHostPort(host, strconv.Itoa(port)),
		Handler: api.NewServerWithOptions(store, recv, ruleEngine, metrics, api.Options{
			RulesSeedFile:        engineCfg.RulesSeedFile,
			AuthEnabled:          engineCfg.APIAuthEnabled,
			AuthToken:            engineCfg.APIAuthToken,
			HealthFreshnessLimit: time.Duration(engineCfg.HealthFreshnessLimitSeconds) * time.Second,
			Suppressions:         suppressions,
			AuditLogger:          logger,
		}).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
}

func startPprofServer(ctx context.Context, engineCfg config.EngineConfig, logger *zap.Logger) {
	if !engineCfg.PprofEnabled {
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	srv := &http.Server{
		Addr:              engineCfg.PprofAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	go func() {
		logger.Info("pprof server started", zap.String("addr", engineCfg.PprofAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("pprof server failed", zap.Error(err))
		}
	}()
}
