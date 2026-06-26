package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/decline-llc/netsentry/internal/config"
	"github.com/decline-llc/netsentry/internal/receiver"
	"github.com/decline-llc/netsentry/internal/rule"
	nssignal "github.com/decline-llc/netsentry/internal/signal"
	"github.com/decline-llc/netsentry/pkg/model"
)

type alertStore struct {
	mu     sync.Mutex
	alerts []*model.Alert
}

func (s *alertStore) Add(alerts ...*model.Alert) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, alert := range alerts {
		now := time.Now().UTC()
		if alert.Timestamp.IsZero() {
			alert.Timestamp = now
		}
		alert.FirstSeen = alert.Timestamp
		alert.LastSeen = alert.Timestamp
		alert.WindowStart = alert.Timestamp
		alert.AggregatedCount = 1
		alert.ID = fmt.Sprintf("%s-%d", alert.RuleID, len(s.alerts)+1)
		alert.EventID = alert.ID
		s.alerts = append(s.alerts, alert)
	}
}

func (s *alertStore) List() []*model.Alert {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*model.Alert, len(s.alerts))
	copy(out, s.alerts)
	return out
}

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

	ctx, cancel := nssignal.WaitForShutdown()
	defer cancel()

	store := &alertStore{}
	recv := receiver.New(receiver.Config{
		Path:       cfg.Engine.UDSSocketPath,
		SocketMode: receiver.ParseSocketMode(cfg.Capture.UDSSocketMode),
		BufferSize: cfg.Engine.ChannelBufferSize,
	}, logger)
	if err := recv.Start(ctx); err != nil {
		logger.Fatal("start uds receiver", zap.Error(err))
	}
	startPacketConsumer(ctx, recv.Packets(), ruleEngine, store, logger)
	startHTTPServer(ctx, cfg.Engine.APIPort, store, logger)

	logger.Info("engine ready — waiting for shutdown signal (SIGINT/SIGTERM)",
		zap.String("uds", cfg.Engine.UDSSocketPath),
		zap.Int("api_port", cfg.Engine.APIPort))
	<-ctx.Done()
	logger.Info("shutdown signal received, exiting")
}

func startPacketConsumer(ctx context.Context, packets <-chan *model.PacketInfo, engine *rule.Engine, store *alertStore, logger *zap.Logger) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case pkt, ok := <-packets:
				if !ok {
					return
				}
				if pkt == nil {
					continue
				}
				alerts := engine.Match(pkt)
				for _, alert := range alerts {
					alert.Timestamp = pkt.Timestamp().UTC()
				}
				store.Add(alerts...)
				if len(alerts) > 0 {
					logger.Info("packet matched rules",
						zap.String("src_ip", pkt.SrcIP),
						zap.String("dst_ip", pkt.DstIP),
						zap.Int("alerts", len(alerts)))
				}
			}
		}
	}()
}

func startHTTPServer(ctx context.Context, port int, store *alertStore, logger *zap.Logger) {
	if port == 0 {
		port = 8080
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "ok",
			"alerts": len(store.List()),
		})
	})
	mux.HandleFunc("/api/alerts", func(w http.ResponseWriter, r *http.Request) {
		alerts := store.List()
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
