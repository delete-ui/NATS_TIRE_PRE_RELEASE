package main

import (
	"fmt"
	"github.com/delete-ui/NATS_TIRE_LIBRARY/shared/config"
	"github.com/delete-ui/NATS_TIRE_LIBRARY/shared/nats"
	"github.com/delete-ui/NATS_TIRE_LIBRARY/shared/types"
	"os"
	"os/signal"
)

type Handler struct{}

func (h *Handler) HandleMatchBundleFound(event types.MatchBundleEvent) error {
	fmt.Printf("\n🎯 ПОЛУЧЕН МАТЧ!\n")
	fmt.Printf("   Матч: %s vs %s\n", event.Payload.TeamNames[0], event.Payload.TeamNames[1])
	fmt.Printf("   Correlation ID: %d\n", event.Payload.CorrelationID)
	fmt.Printf("   Букмекеров: %d\n", len(event.Payload.BookmakerBundle))

	for bm := range event.Payload.BookmakerBundle {
		fmt.Printf("     - %s\n - %s\n", bm, event.Payload.BookmakerBundle[bm])
	}
	return nil
}

func (h *Handler) HandleMatchMonitoring(event types.MatchMonitoringEvent) error {
	fmt.Printf("\n📊 ПОЛУЧЕН МОНИТОРИНГ МАТЧА!\n")
	fmt.Printf("   Матч: %s vs %s\n", event.Payload.TeamNames[0], event.Payload.TeamNames[1])
	fmt.Printf("   Correlation ID: %d\n", event.Payload.CorrelationID)

	for market, bet := range event.Payload.Bets {
		fmt.Printf("     %s: Less=%.2f More=%.2f\n", market, bet.Less, bet.More)
	}
	return nil
}

func (h *Handler) HandleForkFound(event types.ForkFoundEvent) error {
	fmt.Printf("\n💰 ПОЛУЧЕНА ВИЛКА!\n")
	fmt.Printf("   Матч: %s vs %s\n", event.Payload.TeamNames[0], event.Payload.TeamNames[1])
	fmt.Printf("   Correlation ID: %d\n", event.Payload.CorrelationID)

	return nil
}

func main() {
	cfg := config.DefaultConfig()
	cfg.NATSURL = "nats://localhost:4222"
	cfg.ServiceName = "test-subscriber"
	cfg.ConsumerName = "test-subscriber"
	cfg.LogLevel = "debug"

	client, err := nats.NewClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("❌ Ошибка подключения: %v", err))
	}
	defer client.Close()

	fmt.Println("👂 Подписчик запущен")
	fmt.Println("   Ожидание сообщений...")
	fmt.Println("   (Нажмите Ctrl+C для выхода)")
	fmt.Println()

	handler := &Handler{}

	if err := client.SubscribeToMatchBundle(handler); err != nil {
		panic(fmt.Sprintf("❌ Ошибка подписки на матчи: %v", err))
	}
	fmt.Println("   ✅ Подписка на MatchBundle")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	fmt.Println("\n👋 Завершаем работу")
}
