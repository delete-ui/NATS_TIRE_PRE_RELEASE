package nats

import (
	"NATS_TIRE_SERVICE/internal/config"
	"NATS_TIRE_SERVICE/pkg/utils"
	"context"
	"fmt"
	"github.com/nats-io/nats-server/v2/server"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Server struct {
	config         *config.Config
	logger         *utils.Logger
	natsServer     *server.Server
	serverOpts     *server.Options
	running        bool
	startTime      time.Time
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

func NewServer(cfg *config.Config, logger *utils.Logger) (*Server, error) {

	ctx, cancel := context.WithCancel(context.Background())

	opts, err := prepareOptions(cfg)
	if err != nil {
		return nil, fmt.Errorf("error preparing NATS options: %w", err)
	}

	return &Server{
		config:         cfg,
		logger:         logger,
		serverOpts:     opts,
		running:        false,
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
	}, nil

}

func prepareOptions(cfg *config.Config) (*server.Options, error) {

	opts := &server.Options{
		Host:          "0.0.0.0",
		Port:          cfg.NATSPort,
		HTTPPort:      cfg.NATSHTTPPort,
		ServerName:    fmt.Sprintf("%s-%s", cfg.AppName, cfg.Env),
		JetStream:     cfg.JetStreamEnabled,
		StoreDir:      cfg.DataDir,
		MaxPayload:    server.MAX_PAYLOAD_SIZE,
		MaxConn:       -1,
		PingInterval:  2 * time.Minute,
		MaxPingsOut:   2,
		WriteDeadline: 10 * time.Second,
		HTTPHost:      "0.0.0.0",
	}

	if cfg.JetStreamEnabled {
		opts.JetStreamMaxMemory = parseMemorySize(cfg.MaxMemoryStore)
		opts.JetStreamMaxStore = parseMemorySize(cfg.MaxFileStore)
	}

	if cfg.NATSHasAuth() {
		opts.Username = cfg.NATSUsername
		opts.Password = cfg.NATSPassword
	}

	if err := prepareDataDirs(cfg); err != nil {
		return nil, err
	}

	return opts, nil
}

func parseMemorySize(sizeStr string) int64 {

	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	var multiplier int64 = 1
	if strings.HasSuffix(sizeStr, "KB") {
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "KB")
	} else if strings.HasSuffix(sizeStr, "MB") {
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
	} else if strings.HasSuffix(sizeStr, "GB") {
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "GB")
	} else if strings.HasSuffix(sizeStr, "TB") {
		multiplier = 1024 * 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "TB")
	}

	var size int64
	fmt.Sscanf(sizeStr, "%d", &size)

	return size * multiplier

}

func prepareDataDirs(cfg *config.Config) error {
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	if cfg.JetStreamEnabled {
		jetstreamDir := filepath.Join(cfg.DataDir, "jetstream")
		if err := os.MkdirAll(jetstreamDir, 0755); err != nil {
			return fmt.Errorf("failed to create jetstream directory: %w", err)
		}
	}

	return nil
}

func (s *Server) Start() error {
	s.logger.Infof("starting NATS server...")
	s.startTime = time.Now()

	s.natsServer = server.New(s.serverOpts)
	if s.natsServer == nil {
		return fmt.Errorf("failed to create NATS server")
	}

	s.configureNATSLogger()

	s.logger.Infof("NATS server configuration",
		zap.String("server_name", s.serverOpts.ServerName),
		zap.Int("port", s.serverOpts.Port),
		zap.Int("http_port", s.serverOpts.HTTPPort),
		zap.Bool("jetstream", s.serverOpts.JetStream),
		zap.Bool("auth_required", s.serverOpts.Username != ""),
		zap.String("data_dir", s.serverOpts.StoreDir),
	)

	go func() {
		s.natsServer.ConfigureLogger()
		s.natsServer.Start()
	}()

	if err := s.waitForStart(30 * time.Second); err != nil {
		return fmt.Errorf("failed to start NATS server: %w", err)
	}

	s.running = true
	s.logServerInfo()

	s.logger.Infof("NATS server started")
	return nil
}

func (s *Server) configureNATSLogger() {
	if s.natsServer == nil {
		return
	}

	natsLogger := &natsLoggerAdapter{logger: s.logger}
	s.natsServer.SetLogger(natsLogger, s.config.IsDevelopment(), s.config.IsDevelopment())
}

type natsLoggerAdapter struct {
	logger *utils.Logger
}

func (n *natsLoggerAdapter) Noticef(format string, v ...interface{}) {
	n.logger.Info(fmt.Sprintf(format, v...))
}

func (n *natsLoggerAdapter) Errorf(format string, v ...interface{}) {
	n.logger.Error(fmt.Sprintf(format, v...))
}

func (n *natsLoggerAdapter) Fatalf(format string, v ...interface{}) {
	n.logger.Fatal(fmt.Sprintf(format, v...))
}

func (n *natsLoggerAdapter) Warnf(format string, v ...interface{}) {
	n.logger.Warn(fmt.Sprintf(format, v...))
}

func (n *natsLoggerAdapter) Debugf(format string, v ...interface{}) {
	n.logger.Debug(fmt.Sprintf(format, v...))
}

func (n *natsLoggerAdapter) Tracef(format string, v ...interface{}) {
	n.logger.Debug(fmt.Sprintf(format, v...))
}

func (s *Server) waitForStart(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	monitoringURL := fmt.Sprintf("http://localhost:%d", s.serverOpts.HTTPPort)

	for time.Now().Before(deadline) {
		resp, err := http.Get(monitoringURL + "/varz")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				s.logger.Info("NATS server monitoring is ready",
					zap.String("url", monitoringURL),
				)
				return nil
			}
		}

		if s.natsServer != nil && s.natsServer.ReadyForConnections(100*time.Millisecond) {
			s.logger.Info("NATS server is ready for connections")
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("server failed to start within %v", timeout)
}

func (s *Server) logServerInfo() {
	if s.natsServer == nil {
		return
	}

	s.logger.Infof("📡 NATS Server Information",
		zap.String("server_name", s.serverOpts.ServerName),
		zap.Int("port", s.serverOpts.Port),
		zap.Int("http_port", s.serverOpts.HTTPPort),
		zap.String("client_url", s.config.GetNATSURL()),
		zap.String("monitoring_url", s.config.GetMonitoringURL()),
		zap.Bool("jetstream", s.config.JetStreamEnabled),
		zap.Bool("auth_required", s.serverOpts.Username != ""),
	)
}

func (s *Server) Stop() error {
	s.logger.Infof("stopping NATS server...")

	s.shutdownCancel()

	if s.natsServer == nil {
		s.logger.Warnf("NATS server was not initialize")
		return nil
	}

	s.natsServer.Shutdown()

	timeout := 10 * time.Second
	deadline := time.Now().Add(timeout)

	for s.IsRunning() && time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
	}

	if s.IsRunning() {
		s.logger.Warnf("NATS server is still running after timeout")
		return fmt.Errorf("NATS server shotdown timeout")
	}

	s.running = false
	s.logger.Infof("NATS server stopped gracefully",
		zap.Duration("uptime", time.Since(s.startTime)),
	)

	return nil
}

func (s *Server) IsRunning() bool {
	return s.running && s.natsServer != nil && s.natsServer.Running()
}

func (s *Server) GetInfo() map[string]interface{} {
	if s.natsServer == nil {
		return map[string]interface{}{
			"running": "false",
		}
	}

	info := map[string]interface{}{
		"running":        s.IsRunning(),
		"uptime":         time.Since(s.startTime).String(),
		"server_name":    s.serverOpts.ServerName,
		"port":           s.serverOpts.Port,
		"http_port":      s.serverOpts.HTTPPort,
		"client_url":     s.config.GetNATSURL(),
		"monitoring_url": s.config.GetMonitoringURL(),
		"jetstream":      s.config.JetStreamEnabled,
		"auth_required":  s.serverOpts.Username != "",
		"data_dir":       s.serverOpts.StoreDir,
	}

	if s.natsServer.Running() {
		info["ready"] = s.natsServer.ReadyForConnections(0)
		info["clients"] = s.getClientCount()
	}

	return info
}

// TODO: реализовать получения количество подключений через мониторинг
func (s *Server) getClientCount() int {
	return 1
}

func (s *Server) GetStats() map[string]interface{} {
	if s.natsServer == nil || !s.natsServer.Running() {
		return nil
	}

	return map[string]interface{}{
		"server_name":     s.serverOpts.ServerName,
		"port":            s.serverOpts.Port,
		"http_port":       s.serverOpts.HTTPPort,
		"jetstream":       s.config.JetStreamEnabled,
		"auth":            s.serverOpts.Username != "",
		"uptime":          time.Since(s.startTime).String(),
		"running":         s.IsRunning(),
		"ready":           s.natsServer.ReadyForConnections(0),
		"max_connections": s.serverOpts.MaxConn,
		"max_payload":     s.serverOpts.MaxPayload,
		"ping_interval":   s.serverOpts.PingInterval.String(),
		"write_deadline":  s.serverOpts.WriteDeadline.String(),
	}
}

func (s *Server) ReloadConfig() error {
	s.logger.Infof("reloading NATS server configurations...")
	s.logger.Warnf("configurations reload requires server restart")

	if err := s.Stop(); err != nil {
		return fmt.Errorf("failed to stop NATS server while reloading configuirations: %w", err)
	}

	time.Sleep(1 * time.Second)

	opts, err := prepareOptions(s.config)
	if err != nil {
		return fmt.Errorf("failed to prepare new NATS options for reloading options: %w", err)
	}

	s.serverOpts = opts

	if err := s.Start(); err != nil {
		return fmt.Errorf("failed to start NATS server after reload configurations: %w", err)
	}

	return nil
}

func (s *Server) ReadyForConnections(timeout time.Duration) bool {
	if s.natsServer == nil {
		return false
	}
	return s.natsServer.ReadyForConnections(timeout)
}

func (s *Server) GetNatsClientURL() string {
	return s.config.GetNATSURL()
}

func (s *Server) GetMonitoringURL() string {
	return s.config.GetMonitoringURL()
}

func (s *Server) GetOptions() *server.Options {
	return s.serverOpts
}

//

//

//

//

//

//

//

//

//

//

//

//

//

//

//
