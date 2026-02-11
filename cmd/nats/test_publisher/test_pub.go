package main

import (
	"NATS_TIRE_SERVICE/shared/config"
	"NATS_TIRE_SERVICE/shared/nats"
	"NATS_TIRE_SERVICE/shared/types"
	"fmt"
	"github.com/google/uuid"
	"time"
)

func main() {
	cfg := config.DefaultConfig()
	cfg.NATSURL = "nats://localhost:4222"
	cfg.ServiceName = "test-publisher"

	client, err := nats.NewClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("Ошибка подключения: %v", err))
	}
	defer client.Close()

	fmt.Println("✅ Издатель запущен, шлю сообщения...")

	for {
		match := types.Match{
			ID:    uuid.New().String(),
			Sport: types.SportCSGO,
			Teams: []string{"NaVi", "G2"},
		}

		err := client.PublishMatchFound(match)
		if err != nil {
			fmt.Printf("❌ Ошибка: %v\n", err)
		} else {
			fmt.Printf("📤 Отправлен матч: %s\n", match.ID)
		}

		time.Sleep(5 * time.Second)
	}
}
