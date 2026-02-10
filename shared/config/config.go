package config

import (
	"NATS_TIRE_SERVICE/shared/constants"
	"os"
	"strconv"
	"time"
)

type Config struct {
	//NATS
	NATSURL      string
	NATSUser     string
	NATSPassword string

	//JetStream
	StreamName     string
	StreamSubjects []string
	MaxMessageAge  time.Duration
	MaxMessages    int64
	Replicas       int

	//Consumers
	ConsumerName  string
	ConsumerGroup string
	AckWait       time.Duration
	MaxDeliver    int
	MaxAckPending int
	PullBatchSize int

	//Timeouts
	ConnectTimeout time.Duration
	RequestTimeout time.Duration
	ReconnectWait  time.Duration
	MaxReconnects  int

	//Logs
	ServiceName string
	LogLevel    string
	Environment string

	//TLS
	TLSCertFile string
	TLSKeyFile  string
	TLSCAFile   string
}

func DefaultConfig() *Config {
	return &Config{
		NATSURL:        "nats://localhost:4222",
		StreamName:     constants.StreamEvents,
		StreamSubjects: []string{"events.>"},
		MaxMessageAge:  constants.MaxMessageAge,
		MaxMessages:    100000,
		Replicas:       constants.Replicas,

		ConsumerName:  "default-consumer",
		AckWait:       constants.DefaultAckWait,
		MaxDeliver:    constants.MaxDeliver,
		MaxAckPending: constants.MaxAckPending,
		PullBatchSize: constants.PullBatchSize,

		ConnectTimeout: constants.ConnectTimeout,
		RequestTimeout: constants.RequestTimeout,
		ReconnectWait:  constants.ReconnectWait,
		MaxReconnects:  constants.MaxReconnects,

		ServiceName: "unknown-service",
		LogLevel:    "info",
		Environment: "development",
	}
}

func LoadFromEnv() *Config {
	cfg := DefaultConfig()

	if url := os.Getenv("NATS_URL"); url != "" {
		cfg.NATSURL = url
	}
	if user := os.Getenv("NATS_USER"); user != "" {
		cfg.NATSUser = user
	}
	if password := os.Getenv("NATS_PASSWORD"); password != "" {
		cfg.NATSPassword = password
	}
	if name := os.Getenv("SERVICE_NAME"); name != "" {
		cfg.ServiceName = name
	}
	if consumer := os.Getenv("CONSUMER_NAME"); consumer != "" {
		cfg.ConsumerName = consumer
	}
	if group := os.Getenv("CONSUMER_GROUP"); group != "" {
		cfg.ConsumerGroup = group
	}
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		cfg.Environment = env
	}
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.LogLevel = level
	}

	if batchSize := os.Getenv("PULL_BATCH_SIZE"); batchSize != "" {
		if val, err := strconv.Atoi(batchSize); err == nil {
			cfg.PullBatchSize = val
		}
	}

	return cfg
}

func (c *Config) GetStreamConfig() map[string]interface{} {
	return map[string]interface{}{
		"name":     c.StreamName,
		"subjects": c.StreamSubjects,
		"max_age":  c.MaxMessageAge,
		"max_msgs": c.MaxMessages,
		"replicas": c.Replicas,
	}
}

func (c *Config) GetConsumerConfig() map[string]interface{} {
	config := map[string]interface{}{
		"ack_wait":        c.AckWait,
		"max_deliver":     c.MaxDeliver,
		"max_ack_pending": c.MaxAckPending,
	}

	if c.ConsumerGroup != "" {
		config["deliver_group"] = c.ConsumerGroup
	}

	return config
}
