package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type Bet struct {
	BetMarket string  `json:"bet_market"`
	TargetBet string  `json:"target_bet"`
	Less      float64 `json:"less"`
	More      float64 `json:"more"`
}

type Payload struct {
	CorrelationID   int               `json:"correlation_id"`
	SportType       string            `json:"sport_type"`
	TeamNames       []string          `json:"team_names"`
	BookmakerBundle map[string]string `json:"bookmaker_bundle"`
	Bets            map[string][]Bet  `json:"bets"`
	Timestamp       time.Time         `json:"timestamp"`
}

type Event struct {
	Payload Payload `json:"payload"`
}

func main() {
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	event := Event{
		Payload: Payload{
			CorrelationID: 999,
			SportType:     "counter-strike",
			TeamNames:     []string{"NaVi", "G2"},
			BookmakerBundle: map[string]string{
				"fonbet":     "https://fonbet.ru/match/999",
				"parivision": "https://parivision.com/match/999",
			},
			Bets: map[string][]Bet{
				"fonbet": {
					{
						BetMarket: "match-winner",
						TargetBet: "NaVi",
						Less:      1.85,
						More:      1.95,
					},
				},
				"parivision": {
					{
						BetMarket: "match-winner",
						TargetBet: "NaVi",
						Less:      2.10,
						More:      1.75,
					},
				},
			},
			Timestamp: time.Now(),
		},
	}

	data, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("📤 Отправляемые данные:")
	fmt.Println(string(data))
	fmt.Println()

	err = nc.Publish("events.match.monitoring", data)
	if err != nil {
		log.Fatal(err)
	}

	nc.Flush()

	fmt.Println("✅ Данные отправлены!")
	fmt.Printf("📊 Размер: %d байт\n", len(data))

	if nc.IsConnected() {
		fmt.Println("✅ NATS подключен")
	}

	time.Sleep(2 * time.Second)
}
