// pkg/nats/factory.go
package nats

import (
	"NATS_TIRE_SERVICE/shared/config"
	"NATS_TIRE_SERVICE/shared/types"
	"fmt"
)

// Factory создает клиент и возвращает интерфейсы Publisher и Consumer
func Factory(cfg *config.Config) (types.Publisher, types.Consumer, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create NATS client: %w", err)
	}

	return client, client, nil
}

// NewPublisher создает только Publisher
func NewPublisher(cfg *config.Config) (types.Publisher, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewConsumer создает только Consumer
func NewConsumer(cfg *config.Config) (types.Consumer, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}
