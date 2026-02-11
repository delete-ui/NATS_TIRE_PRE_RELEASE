package nats

import (
	"NATS_TIRE_SERVICE/shared/config"
	"NATS_TIRE_SERVICE/shared/constants"
	"NATS_TIRE_SERVICE/shared/types"
	"NATS_TIRE_SERVICE/shared/utils"
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"sync"
)

type Client struct {
	//Connection to NATS
	conn      *nats.Conn
	jetStream nats.JetStreamContext

	//Conf
	config *config.Config

	//Utils
	logger *utils.Logger
	events *utils.EventHelper

	//Subs
	subscriptions []*nats.Subscription
	handlers      map[string]types.EventHandler

	//Sync
	mu       sync.RWMutex
	isClosed bool

	//Graceful shutdown
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
}

func NewClient(cfg *config.Config) (*Client, error) {

	ctx, cancel := context.WithCancel(context.Background())

	logger, err := utils.NewLogger(cfg.ServiceName, cfg.Environment, cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	events := utils.NewEventHelper(cfg.ServiceName, constants.ProtocolVersion)

	client := &Client{
		config:     cfg,
		logger:     logger,
		events:     events,
		handlers:   make(map[string]types.EventHandler),
		ctx:        ctx,
		cancelFunc: cancel,
	}

	if err := client.connect(); err != nil {
		client.Close()
		return nil, err
	}

	// Настраиваем JetStream
	if err := client.setupJetStream(); err != nil {
		client.Close()
		return nil, err
	}

	logger.Info("NATS client initialized successfully",
		zap.String("url", cfg.NATSURL),
		zap.String("stream", cfg.StreamName),
		zap.String("service", cfg.ServiceName))

	return client, nil

}

func (c *Client) connect() error {

	opts := []nats.Option{
		nats.Timeout(c.config.ConnectTimeout),
		nats.ReconnectWait(c.config.ReconnectWait),
		nats.MaxReconnects(c.config.MaxReconnects),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			c.logger.Warn("disconnected from NATS", zap.Error(err))
		}),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			c.logger.Info("reconnected to NATS",
				zap.String("url", conn.ConnectedUrl()))
		}),
		nats.ClosedHandler(func(conn *nats.Conn) {
			c.logger.Info("connection to NATS closed")
		}),
		nats.ErrorHandler(func(conn *nats.Conn, sub *nats.Subscription, err error) {
			c.logger.Error("NATS error", zap.Error(err),
				zap.String("subject", sub.Subject))
		}),
	}

	if c.config.NATSUser != "" && c.config.NATSPassword != "" {
		opts = append(opts, nats.UserInfo(c.config.NATSUser, c.config.NATSPassword))
	}

	if c.config.TLSCertFile != "" && c.config.TLSKeyFile != "" {
		opts = append(opts, nats.ClientCert(c.config.TLSCertFile, c.config.TLSKeyFile))
		if c.config.TLSCAFile != "" {
			opts = append(opts, nats.RootCAs(c.config.TLSCAFile))
		}
	}

	nc, err := nats.Connect(c.config.NATSURL, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	c.conn = nc
	return nil
}

func (c *Client) setupJetStream() error {
	js, err := c.conn.JetStream()
	if err != nil {
		return fmt.Errorf("failed to get JetStream context: %w", err)
	}
	c.jetStream = js

	stream, err := js.StreamInfo(c.config.StreamName)
	if err != nil {
		c.logger.Info("creating new stream",
			zap.String("stream", c.config.StreamName))

		streamConfig := &nats.StreamConfig{
			Name:     c.config.StreamName,
			Subjects: c.config.StreamSubjects,
			MaxAge:   c.config.MaxMessageAge,
			MaxMsgs:  c.config.MaxMessages,
			Storage:  nats.FileStorage,
			Replicas: c.config.Replicas,
		}

		_, err = js.AddStream(streamConfig)
		if err != nil {
			return fmt.Errorf("failed to create stream: %w", err)
		}
	} else {
		c.logger.Debug("stream already exists",
			zap.String("stream", stream.Config.Name),
			zap.Int("messages", int(stream.State.Msgs)))
	}

	return nil
}
