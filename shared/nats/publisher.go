package nats

import (
	"NATS_TIRE_SERVICE/shared/constants"
	"NATS_TIRE_SERVICE/shared/types"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
)

// PublishMatchFound публикует событие о найденном матче
func (c *Client) PublishMatchFound(match types.Match) error {
	return c.publishEvent(constants.TopicMatchFound, match, types.EventTypeMatchFound)
}

// PublishOddsUpdated публикует событие об обновленных коэффициентах
func (c *Client) PublishOddsUpdated(odds types.Odds) error {
	return c.publishEvent(constants.TopicOddsUpdated, odds, types.EventTypeOddsUpdated)
}

// PublishForkFound публикует событие о найденной вилке
func (c *Client) PublishForkFound(fork types.Fork) error {
	return c.publishEvent(constants.TopicForkFound, fork, types.EventTypeForkFound)
}

// PublishError публикует событие об ошибке
func (c *Client) PublishError(service, operation, errMsg string, critical bool) error {
	event := c.events.CreateErrorEvent(operation, errMsg, critical, "")
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal error event: %w", err)
	}

	return c.publish(constants.TopicErrors, data)
}

// PublishHealthCheck публикует событие health check
func (c *Client) PublishHealthCheck(service, status, message string) error {
	event := c.events.CreateHealthCheckEvent(status, message, nil)
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal health check event: %w", err)
	}

	return c.publish(constants.TopicHealth, data)
}

// publishEvent универсальный метод для публикации событий
func (c *Client) publishEvent(topic string, payload interface{}, eventType types.EventType) error {
	var event interface{}

	// Создаем соответствующее событие
	switch eventType {
	case types.EventTypeMatchFound:
		match, ok := payload.(types.Match)
		if !ok {
			return fmt.Errorf("invalid payload type for match event")
		}
		event = c.events.CreateMatchFoundEvent(match, "")

	case types.EventTypeOddsUpdated:
		odds, ok := payload.(types.Odds)
		if !ok {
			return fmt.Errorf("invalid payload type for odds event")
		}
		event = c.events.CreateOddsUpdatedEvent(odds, "")

	case types.EventTypeForkFound:
		fork, ok := payload.(types.Fork)
		if !ok {
			return fmt.Errorf("invalid payload type for fork event")
		}
		event = c.events.CreateForkFoundEvent(fork, "")

	default:
		return fmt.Errorf("unsupported event type: %s", eventType)
	}

	// Сериализуем событие
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Публикуем
	return c.publish(topic, data)
}

// publish базовый метод публикации
func (c *Client) publish(subject string, data []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isClosed {
		return fmt.Errorf("client is closed")
	}

	// Используем публикацию с подтверждением
	ack, err := c.jetStream.Publish(subject, data)
	if err != nil {
		c.logger.Error("failed to publish message",
			zap.Error(err),
			zap.String("subject", subject))
		return err
	}

	// Логируем успешную публикацию в debug режиме
	c.logger.Debug("message published successfully",
		zap.String("subject", subject),
		zap.String("stream", ack.Stream),
		zap.Uint64("sequence", ack.Sequence))

	return nil
}
