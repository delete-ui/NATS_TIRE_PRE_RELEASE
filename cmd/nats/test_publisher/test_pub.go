package main

import (
	"fmt"
	"github.com/delete-ui/NATS_TIRE_LIBRARY/shared/config"
	"github.com/delete-ui/NATS_TIRE_LIBRARY/shared/nats"
	"github.com/delete-ui/NATS_TIRE_LIBRARY/shared/types"
	"time"
)

func main() {
	cfg := config.DefaultConfig()
	cfg.NATSURL = "nats://localhost:4222"
	cfg.ServiceName = "test-publisher"

	client, err := nats.NewClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("❌ Ошибка подключения: %v", err))
	}
	defer client.Close()

	fmt.Println("✅ Издатель запущен, отправляю сообщения...")

	correlationID := 1

	for {
		bundle := types.MatchBundle{
			CorrelationID: correlationID,
			TeamNames:     []string{"NaVi", "G2"},
			BookmakerBundle: map[types.Bookmaker]string{
				types.BookmakerFonbet: "https://fonbet.ru/match/123",
			},
		}

		err := client.PublishMatchBundle(bundle)
		if err != nil {
			fmt.Printf("❌ Ошибка публикации: %v\n", err)
		} else {
			fmt.Printf("📤 Отправлен матч (correlation_id: %d): %v\n",
				bundle.CorrelationID, bundle.TeamNames)
		}

		correlationID++
		time.Sleep(5 * time.Second)
	}
}
