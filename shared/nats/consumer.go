// pkg/nats/consumer.go
package nats

import (
	"NATS_TIRE_SERVICE/shared/constants"
	"NATS_TIRE_SERVICE/shared/types"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"time"

	"github.com/nats-io/nats.go"
)

// SubscribeToMatches подписывается на события матчей
func (c *Client) SubscribeToMatches(handler types.EventHandler) error {
	return c.subscribe(constants.TopicMatchFound, handler, c.handleMatchMessage)
}

// SubscribeToOdds подписывается на события коэффициентов
func (c *Client) SubscribeToOdds(handler types.EventHandler) error {
	return c.subscribe(constants.TopicOddsUpdated, handler, c.handleOddsMessage)
}

// SubscribeToForks подписывается на события вилок
func (c *Client) SubscribeToForks(handler types.EventHandler) error {
	return c.subscribe(constants.TopicForkFound, handler, c.handleForkMessage)
}

// SubscribeToErrors подписывается на события ошибок
func (c *Client) SubscribeToErrors(handler types.EventHandler) error {
	return c.subscribe(constants.TopicErrors, handler, c.handleErrorMessage)
}

// subscribe универсальный метод подписки - ИСПРАВЛЕННАЯ ВЕРСИЯ
func (c *Client) subscribe(subject string, handler types.EventHandler, msgHandler func(*nats.Msg, types.EventHandler) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosed {
		return fmt.Errorf("client is closed")
	}

	// Правильный способ создания потребителя
	// Сначала проверяем/создаем durable consumer
	consumerConfig := &nats.ConsumerConfig{
		Durable:       c.config.ConsumerName,
		DeliverGroup:  c.config.ConsumerGroup,
		AckWait:       c.config.AckWait,
		MaxDeliver:    c.config.MaxDeliver,
		MaxAckPending: c.config.MaxAckPending,
		FilterSubject: subject,
		AckPolicy:     nats.AckExplicitPolicy,
		DeliverPolicy: nats.DeliverNewPolicy,
		ReplayPolicy:  nats.ReplayInstantPolicy,
	}

	// Пытаемся добавить или обновить потребителя
	_, err := c.jetStream.AddConsumer(c.config.StreamName, consumerConfig)
	if err != nil {
		// Если потребитель уже существует - используем его
		if err != nats.ErrConsumerNameAlreadyInUse {
			c.logger.Warn("consumer already exists, using existing",
				zap.String("consumer", c.config.ConsumerName))
		} else {
			return fmt.Errorf("failed to add consumer: %w", err)
		}
	}

	// Создаем pull подписку с правильными опциями
	sub, err := c.jetStream.PullSubscribe(
		subject,
		c.config.ConsumerName,
		[]nats.SubOpt{
			nats.Bind(c.config.StreamName, c.config.ConsumerName),
			nats.ManualAck(),
		}...,
	)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	// Сохраняем хендлер
	c.handlers[subject] = handler
	c.subscriptions = append(c.subscriptions, sub)

	// Запускаем обработку сообщений в отдельной горутине
	c.wg.Add(1)
	go c.processSubscription(sub, subject, msgHandler)

	c.logger.Info("subscribed to subject",
		zap.String("subject", subject),
		zap.String("consumer", c.config.ConsumerName),
		zap.String("stream", c.config.StreamName))

	return nil
}

// Альтернативный упрощенный метод - если не нужно создавать consumer отдельно
func (c *Client) subscribeSimple(subject string, handler types.EventHandler, msgHandler func(*nats.Msg, types.EventHandler) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosed {
		return fmt.Errorf("client is closed")
	}

	// Самый простой способ - создаем подписку с автоматическим созданием consumer
	sub, err := c.jetStream.PullSubscribe(
		subject,
		c.config.ConsumerName,
		nats.ManualAck(),
		nats.BindStream(c.config.StreamName),
	)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	c.handlers[subject] = handler
	c.subscriptions = append(c.subscriptions, sub)

	c.wg.Add(1)
	go c.processSubscription(sub, subject, msgHandler)

	return nil
}

// processSubscription обрабатывает сообщения из подписки
func (c *Client) processSubscription(sub *nats.Subscription, subject string, msgHandler func(*nats.Msg, types.EventHandler) error) {
	defer c.wg.Done()

	// Создаем тикер для пауз между итерациями
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Debug("stopping subscription processing",
				zap.String("subject", subject))
			return

		case <-ticker.C:
			// Получаем пакет сообщений с таймаутом
			msgs, err := sub.Fetch(c.config.PullBatchSize, nats.MaxWait(5*time.Second))
			if err != nil {
				if err == nats.ErrTimeout {
					continue
				}
				// Если нет сообщений или другие ошибки - логируем и продолжаем
				if err.Error() != "nats: no messages" {
					c.logger.Debug("fetch messages result",
						zap.Error(err),
						zap.String("subject", subject))
				}
				continue
			}

			// Обрабатываем каждое сообщение
			for _, msg := range msgs {
				handler := c.getHandler(subject)
				if handler == nil {
					c.logger.Error("no handler found for subject",
						zap.String("subject", subject))
					msg.Ack() // Не блокируем очередь
					continue
				}

				if err := msgHandler(msg, handler); err != nil {
					c.logger.Error("failed to process message",
						zap.Error(err),
						zap.String("subject", subject))

					// Можно NAK сообщение для повторной обработки
					if nakErr := msg.Nak(); nakErr != nil {
						c.logger.Error("failed to nak message", zap.Error(nakErr))
					}
				}
			}
		}
	}
}

// getHandler безопасно получает обработчик
func (c *Client) getHandler(subject string) types.EventHandler {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.handlers[subject]
}

// Обработчики сообщений
func (c *Client) handleMatchMessage(msg *nats.Msg, handler types.EventHandler) error {
	var event types.MatchFoundEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal match event: %w", err)
	}

	if err := handler.HandleMatchFound(event); err != nil {
		return fmt.Errorf("handler failed for match event: %w", err)
	}

	// Подтверждаем обработку
	if err := msg.Ack(); err != nil {
		return fmt.Errorf("failed to ack message: %w", err)
	}

	c.logger.Debug("match event processed",
		zap.String("match_id", event.Payload.ID),
		zap.String("event_id", event.EventHeader.EventID))

	return nil
}

func (c *Client) handleOddsMessage(msg *nats.Msg, handler types.EventHandler) error {
	var event types.OddsUpdatedEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal odds event: %w", err)
	}

	if err := handler.HandleOddsUpdated(event); err != nil {
		return fmt.Errorf("handler failed for odds event: %w", err)
	}

	if err := msg.Ack(); err != nil {
		return fmt.Errorf("failed to ack message: %w", err)
	}

	return nil
}

func (c *Client) handleForkMessage(msg *nats.Msg, handler types.EventHandler) error {
	var event types.ForkFoundEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal fork event: %w", err)
	}

	if err := handler.HandleForkFound(event); err != nil {
		return fmt.Errorf("handler failed for fork event: %w", err)
	}

	if err := msg.Ack(); err != nil {
		return fmt.Errorf("failed to ack message: %w", err)
	}

	return nil
}

func (c *Client) handleErrorMessage(msg *nats.Msg, handler types.EventHandler) error {
	var event types.ErrorEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal error event: %w", err)
	}

	if err := handler.HandleError(event); err != nil {
		return fmt.Errorf("handler failed for error event: %w", err)
	}

	if err := msg.Ack(); err != nil {
		return fmt.Errorf("failed to ack message: %w", err)
	}

	return nil
}

// Unsubscribe отписывается от всех подписок
func (c *Client) Unsubscribe() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var lastErr error
	for _, sub := range c.subscriptions {
		if err := sub.Unsubscribe(); err != nil {
			lastErr = err
			c.logger.Error("failed to unsubscribe",
				zap.Error(err),
				zap.String("subject", sub.Subject))
		}
	}

	c.subscriptions = nil
	c.handlers = make(map[string]types.EventHandler)

	if lastErr != nil {
		return fmt.Errorf("failed to unsubscribe some subscriptions: %w", lastErr)
	}
	return nil
}

// Добавим метод для переподключения потребителя
func (c *Client) RecreateConsumer(subject string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Удаляем старого потребителя если существует
	if err := c.jetStream.DeleteConsumer(c.config.StreamName, c.config.ConsumerName); err != nil {
		c.logger.Warn("failed to delete consumer",
			zap.Error(err),
			zap.String("consumer", c.config.ConsumerName))
	}

	// Создаем нового потребителя
	consumerConfig := &nats.ConsumerConfig{
		Durable:       c.config.ConsumerName,
		DeliverGroup:  c.config.ConsumerGroup,
		AckWait:       c.config.AckWait,
		MaxDeliver:    c.config.MaxDeliver,
		MaxAckPending: c.config.MaxAckPending,
		FilterSubject: subject,
		AckPolicy:     nats.AckExplicitPolicy,
		DeliverPolicy: nats.DeliverNewPolicy,
		ReplayPolicy:  nats.ReplayInstantPolicy,
	}

	_, err := c.jetStream.AddConsumer(c.config.StreamName, consumerConfig)
	return err
}
