package main

import (
	"NATS_TIRE_SERVICE/shared/config"
	"NATS_TIRE_SERVICE/shared/nats"
	"NATS_TIRE_SERVICE/shared/types"
	"fmt"
	"os"
	"os/signal"
)

type Handler struct{}

func (h *Handler) HandleMatchFound(event types.MatchFoundEvent) error {
	fmt.Printf("\n🎯 ПОЛУЧЕНО СООБЩЕНИЕ В ДРУГОМ ПРОЕКТЕ!\n")
	fmt.Printf("   Матч: %s vs %s\n", event.Payload.Teams[0], event.Payload.Teams[1])
	fmt.Printf("   ID: %s\n", event.Payload.ID)
	fmt.Printf("   Время: %s\n\n", event.Timestamp)
	return nil
}

func (h *Handler) HandleOddsUpdated(event types.OddsUpdatedEvent) error { return nil }
func (h *Handler) HandleForkFound(event types.ForkFoundEvent) error     { return nil }
func (h *Handler) HandleError(event types.ErrorEvent) error             { return nil }

func main() {
	cfg := config.DefaultConfig()
	cfg.NATSURL = "nats://localhost:4222" // Тот же адрес!
	cfg.ServiceName = "test-subscriber-other-project"
	cfg.ConsumerName = "other-project-consumer"

	client, err := nats.NewClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("❌ Ошибка подключения: %v", err))
	}
	defer client.Close()

	fmt.Println("👂 Подписчик из ДРУГОГО ПРОЕКТА запущен")
	fmt.Println("   Жду сообщения от первого проекта...")

	err = client.SubscribeToMatches(&Handler{})
	if err != nil {
		panic(fmt.Sprintf("❌ Ошибка подписки: %v", err))
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	fmt.Println("\n👋 Завершаем работу")
}
