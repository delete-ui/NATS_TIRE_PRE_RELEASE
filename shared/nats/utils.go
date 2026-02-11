// pkg/nats/utils.go
package nats

import (
	"fmt"
	"time"
)

// HealthCheck проверяет состояние подключения
func (c *Client) HealthCheck() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isClosed {
		return fmt.Errorf("client is closed")
	}

	if !c.conn.IsConnected() {
		return fmt.Errorf("not connected to NATS")
	}

	if !c.conn.IsReconnecting() {
		// Проверяем доступность стрима
		_, err := c.jetStream.StreamInfo(c.config.StreamName)
		if err != nil {
			return fmt.Errorf("stream not available: %w", err)
		}
	}

	return nil
}

// GetStats возвращает статистику клиента
func (c *Client) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := map[string]interface{}{
		"is_connected":  c.conn.IsConnected(),
		"is_closed":     c.isClosed,
		"url":           c.conn.ConnectedUrl(),
		"subscriptions": len(c.subscriptions),
		"handlers":      len(c.handlers),
	}

	if js, err := c.conn.JetStream(); js != nil {
		if info, err := js.StreamInfo(c.config.StreamName); err == nil {
			stats["stream_messages"] = info.State.Msgs
			stats["stream_bytes"] = info.State.Bytes
		}
		c.logger.Warn(err.Error())
	}

	return stats
}

// Close закрывает клиент и освобождает ресурсы
func (c *Client) Close() error {
	c.mu.Lock()
	if c.isClosed {
		c.mu.Unlock()
		return nil
	}
	c.isClosed = true
	c.mu.Unlock()

	// Отменяем контекст чтобы остановить все горутины
	c.cancelFunc()

	// Отписываемся от всех подписок
	c.Unsubscribe()

	// Ждем завершения всех горутин
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Все горутины завершились
	case <-time.After(10 * time.Second):
		c.logger.Warn("timed out waiting for goroutines to finish")
	}

	// Закрываем соединение
	c.conn.Close()

	// Синхронизируем логгер
	c.logger.Sync()

	c.logger.Info("NATS client closed successfully")
	return nil
}

// IsConnected проверяет подключение
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.isClosed && c.conn.IsConnected()
}
