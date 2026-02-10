package main

import (
	"NATS_TIRE_SERVICE/internal/config"
	"NATS_TIRE_SERVICE/internal/nats"
	"NATS_TIRE_SERVICE/internal/server"
	"NATS_TIRE_SERVICE/pkg/utils"
	"context"
	"fmt"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// TODO: доработки в файлах: /server/http.go, /config/config.go, /nats/server.go, /shared/types/models.go
func main() {
	cfg, err := config.LoadConfigurations(".env")
	if err != nil {
		fmt.Println(err)
	}

	logger, err := utils.NewLogger(cfg)
	if err != nil {
		fmt.Errorf("Failed to create logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			if !cfg.IsDevelopment() {
				log.Printf("Failed to sync logger: %s", err)
			}
		}
	}()

	logger.Info("Starting NATS Service",
		zap.String("app", cfg.AppName),
		zap.String("version", cfg.Version),
		zap.String("env", cfg.Env),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	natsServer, err := nats.NewServer(cfg, logger)
	if err != nil {
		log.Fatalf("failed to create NATS server: %v", err)
	}

	var wg sync.WaitGroup

	httpServer := server.NewHTTPServer(cfg.HTTPPort, logger, natsServer, cfg)
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting HTTP server",
			zap.Int("port", cfg.HTTPPort),
			zap.String("url", cfg.GetHealthURL()),
		)
		if err := httpServer.Start(); err != nil {
			logger.Errorf("HTTP server error", zap.Error(err))
		}
	}()

	time.Sleep(2 * time.Second)

	httpReady := false
	for i := 0; i < 10; i++ {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", cfg.HTTPPort))
		if err == nil {
			resp.Body.Close()
			httpReady = true
			logger.Info("HTTP server is ready")
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if !httpReady {
		logger.Warn("HTTP server may not be responding, but continuing...")
	}

	logger.Infof("Initializing NATS server",
		zap.Int("port", cfg.NATSPort),
		zap.Int("http_port", cfg.NATSHTTPPort),
		zap.Bool("jetstream", cfg.JetStreamEnabled),
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting NATS server")

		if err := natsServer.Start(); err != nil {
			logger.Error("NATS server error", zap.Error(err))
		}
	}()

	time.Sleep(2 * time.Second)

	if !natsServer.IsRunning() {
		logger.Warn("NATS server may not be running, but continuing...")
	} else {
		logger.Info(" NATS server confirmed running")
	}

	logger.Info("✅ Service started successfully!",
		zap.String("nats_url", cfg.GetNATSURL()),
		zap.String("monitoring_url", cfg.GetMonitoringURL()),
		zap.String("health_url", cfg.GetHealthURL()),
	)
	logger.Info(" Press Ctrl+C to stop")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // Docker/Kubernetes stop
		syscall.SIGQUIT, // Graceful shutdown
	)

	select {
	case sig := <-signalChan:
		logger.Info("Received signal",
			zap.String("signal", sig.String()),
		)

	case <-ctx.Done():
		logger.Info("Context cancelled")
	}

	logger.Info("Shutdown initiated")

	logger.Info("Stopping NATS server...")
	if err := natsServer.Stop(); err != nil {
		logger.Error("Error stopping NATS server", zap.Error(err))
	}

	logger.Info("Stopping HTTP server...")
	if err := httpServer.Stop(); err != nil {
		logger.Error("Error stopping HTTP server", zap.Error(err))
	}

	wg.Wait()

	logger.Info("Service stopped gracefully")
}
