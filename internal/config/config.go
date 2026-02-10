package config

import (
	"NATS_TIRE_SERVICE/internal/domain"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"os"
	"strings"
)

type Config struct {
	domain.MainSettings
	domain.NATSTireServer
	domain.JetStreamSettings
	domain.NATSSecurity
	domain.LoggerSettings
	domain.HTTPServer
}

func LoadConfigurations(envFile string) (*Config, error) {

	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			fmt.Printf("Notice: .env file not found at %s, using environment variables\n", envFile)
		}
	} else {
		if err := godotenv.Load(); err != nil {
			fmt.Printf("Notice: .env file not found, using environment variables\n")
		}
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) createDirs(cfg Config) error {

	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	if cfg.LogFile != "" {
		logDir := getDirFromPath(cfg.LogFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	return nil

}

func getDirFromPath(filePath string) string {
	if idx := strings.LastIndex(filePath, "/"); idx != -1 {
		return filePath[:idx]
	}
	if idx := strings.LastIndex(filePath, "\\"); idx != -1 {
		return filePath[:idx]
	}

	return "."
}

// TODO: расширить валидирование конфигураций
func validateConfig(cfg *Config) error {

	if cfg.NATSPort < 1 || cfg.NATSPort > 65535 {
		return fmt.Errorf("invalid NATS port: %d", cfg.NATSPort)
	}

	if cfg.NATSHTTPPort < 1 || cfg.NATSHTTPPort > 65535 {
		return fmt.Errorf("invalid NATS http port: %d", cfg.NATSHTTPPort)
	}
	if cfg.HTTPPort < 1 || cfg.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", cfg.HTTPPort)
	}

	ports := map[int]bool{
		cfg.NATSHTTPPort: true,
		cfg.NATSHTTPPort: true,
		cfg.HTTPPort:     true,
	}

	if len(ports) != 2 {
		return fmt.Errorf("ports must be unique: NATS=%d, NATS_HTTP=%d, HTTP=%d",
			cfg.NATSPort, cfg.NATSHTTPPort, cfg.HTTPPort)
	}

	if cfg.NATSUsername != "" && cfg.NATSPassword == "" {
		return fmt.Errorf("password is required when username is set")
	}

	if cfg.LogFormat != "json" && cfg.LogFormat != "console" {
		return fmt.Errorf("invalid log format: %s", cfg.LogFormat)
	}

	return nil
}

func (c *Config) GetNATSURL() string {
	return fmt.Sprintf("nats://localhost:%d", c.NATSPort)
}

func (c *Config) GetMonitoringURL() string {
	return fmt.Sprintf("http://localhost:%d", c.NATSHTTPPort)
}

func (c *Config) GetHealthURL() string {
	return fmt.Sprintf("http://localhost:%d", c.HTTPPort)
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

func (c *Config) NATSHasAuth() bool {
	return c.NATSUsername != "" && c.NATSPassword != ""
}
