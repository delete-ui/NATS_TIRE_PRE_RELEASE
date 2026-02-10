package server

import (
	"NATS_TIRE_SERVICE/internal/config"
	"NATS_TIRE_SERVICE/internal/domain"
	"NATS_TIRE_SERVICE/internal/nats"
	"NATS_TIRE_SERVICE/pkg/utils"
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type HTTPServer struct {
	server     *http.Server
	logger     *utils.Logger
	port       int
	startTime  time.Time
	natsServer *nats.Server
	cfg        *config.Config
}

func NewHTTPServer(port int, logger *utils.Logger, natsServer *nats.Server, cfg *config.Config) *HTTPServer {

	mux := http.NewServeMux()

	server := &HTTPServer{
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  20 * time.Second,
		},
		logger:    logger,
		port:      port,
		startTime: time.Now(),
	}

	mux.HandleFunc("/health", server.healthHandler)
	mux.HandleFunc("/ready", server.readyHandler)
	mux.HandleFunc("/live", server.liveHandler)
	mux.HandleFunc("/metrics", server.metricsHandler)
	mux.HandleFunc("/", server.rootHandler)

	server.natsServer = natsServer
	server.cfg = cfg
	return server
}

func (s *HTTPServer) Start() error {
	s.logger.Infof("http server starting on port %d", s.port)

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Errorf("http server error %w", err)
	}

	return nil
}

func (s *HTTPServer) Stop() error {
	s.logger.Infof("stopping http server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Errorf("http server shutdown error %w", err)
	}

	s.logger.Infof("http server stopped")
	return nil
}

func (s *HTTPServer) sendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Failed to encode JSON response",
			zap.Error(err),
			zap.Any("data", data),
		)
	}
}

func (s *HTTPServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := domain.HealthResponse{
		Service:   s.cfg.AppName,
		Version:   s.cfg.Version,
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Uptime:    time.Since(s.startTime).String(),
	}

	s.sendJSONResponse(w, http.StatusOK, response)
	s.logger.Debug("Health check requested",
		zap.String("client", r.RemoteAddr),
		zap.String("user-agent", r.UserAgent()),
	)
}

// TODO: доработать проверку готовности сервиса к работе
func (s *HTTPServer) readyHandler(w http.ResponseWriter, r *http.Request) {

	response := domain.HealthResponse{
		Service:   s.cfg.AppName,
		Version:   s.cfg.Version,
		Status:    "ready",
		Timestamp: time.Now().UTC(),
		Uptime:    time.Since(s.startTime).String(),
	}

	s.sendJSONResponse(w, http.StatusOK, response)
}

// TODO: доработать проверку isAlive сервиса
func (s *HTTPServer) liveHandler(w http.ResponseWriter, r *http.Request) {

	response := map[string]interface{}{
		"service": s.cfg.AppName,
		"version": s.cfg.Version,
		"status":  "alive",
		"time":    time.Now().UTC().Format(time.RFC3339Nano),
	}

	s.sendJSONResponse(w, http.StatusOK, response)
}

func (s *HTTPServer) metricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := map[string]interface{}{
		"service":        s.cfg.AppName,
		"version":        s.cfg.Version,
		"uptime_seconds": time.Since(s.startTime).Seconds(),
		"start_time":     s.startTime.Format(time.RFC3339),
		"current_time":   time.Now().UTC().Format(time.RFC3339),
	}

	s.sendJSONResponse(w, http.StatusOK, metrics)
}

func (s *HTTPServer) rootHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	info := map[string]interface{}{
		"service":       s.cfg.AppName,
		"version":       s.cfg.Version,
		"endpoints":     []string{"/health", "/ready", "/live", "/metrics"},
		"documentation": "Health check endpoints for NATS service",
	}

	s.sendJSONResponse(w, http.StatusOK, info)
}
