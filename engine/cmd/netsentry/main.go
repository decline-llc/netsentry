package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/decline-llc/netsentry/internal/alert"
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
	startHTTPServer(ctx, cfg.Engine.APIPort, store, recv, ruleEngine, metrics, logger)

	logger.Info("engine ready — waiting for shutdown signal (SIGINT/SIGTERM)",
		zap.String("uds", cfg.Engine.UDSSocketPath),
		zap.Int("api_port", cfg.Engine.APIPort))
	<-ctx.Done()
	logger.Info("shutdown signal received, exiting")
}

func startHTTPServer(ctx context.Context, port int, store *alert.Store, recv *receiver.Receiver, ruleEngine *rule.Engine, metrics *stats.Stats, logger *zap.Logger) {
	if port == 0 {
		port = 8080
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		count, err := store.Count(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "ok",
			"alerts": count,
		})
	})
	mux.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
		count, err := store.Count(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		body := stats.RenderPrometheus(metrics.Snapshot(), map[string]float64{
			"netsentry_alerts_current":     float64(count),
			"netsentry_packet_queue_depth": float64(recv.QueueDepth()),
			"netsentry_rules_loaded":       float64(ruleEngine.RuleCount()),
		})
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	})
	mux.HandleFunc("/api/alerts", func(w http.ResponseWriter, r *http.Request) {
		alerts, err := store.List(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"alerts": alerts,
			"total":  len(alerts),
		})
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
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
		logger.Info("http api started", zap.Int("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http api failed", zap.Error(err))
		}
	}()
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
