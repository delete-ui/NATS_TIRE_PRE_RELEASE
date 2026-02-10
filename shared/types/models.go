package types

import "time"

type SportType string

const (
	SportCSGO     SportType = "counter-strike"
	SportDota2    SportType = "dota2"
	SportLoL      SportType = "league-of-legends"
	SportValorant SportType = "valorant"
	SportRainbow6 SportType = "rainbow-six"
)

type Bookmaker string

const (
	BookmakerParivision Bookmaker = "parivision"
	BookmakerFonbet     Bookmaker = "fonbet"
	BookmakerOlimpBet   Bookmaker = "olimp-bet"
	BookmakerBetBoom    Bookmaker = "bet-boom"
	BookmakerWinline    Bookmaker = "winline"
)

type MarketType string

const (
	MarketMatchWinner MarketType = "match-winner"
	MarketTotalMaps   MarketType = "total-maps"
	MarketHandicap    MarketType = "handicap"
	MarketMainTotal   MarketType = "main-total"
	MarketPainting    MarketType = "painting"
)

type Outcome struct {
	BetTitle string  `json:"name"`
	Less     float64 `json:"less"`
	More     float64 `json:"more"`
}

type Market struct {
	Type      MarketType `json:"type"`
	Name      string     `json:"name"`
	Outcomes  []Outcome  `json:"outcomes"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// TODO: Добавить информацию о лигах/турнирах
type Match struct {
	ID         string    `json:"id"`
	ExternalID string    `json:"external_id"`
	Sport      SportType `json:"sport"`
	Teams      []string  `json:"teams"`
	StartsAt   time.Time `json:"starts_at"`
	URL        string    `json:"url,omitempty"`
}

type Odds struct {
	MatchID   string    `json:"match_id"`
	Bookmaker Bookmaker `json:"bookmaker"`
	Markets   []Market  `json:"markets"`
	Timestamp time.Time `json:"timestamp"`
}

type ArbitrageLeg struct {
	Bookmaker        Bookmaker `json:"bookmaker"`
	Market           Market    `json:"market"`
	Outcome          Outcome   `json:"outcome"`
	Stake            float64   `json:"stake"`
	Profit           float64   `json:"profit"`
	ProfitPercentage float64   `json:"profit_percentage"`
}

type Fork struct {
	ID               string         `json:"id"`
	MatchID          string         `json:"match_id"`
	Sport            SportType      `json:"sport"`
	Teams            []string       `json:"teams"`
	Profit           float64        `json:"profit"`
	ProfitPercentage float64        `json:"profit_percentage"`
	Arbitrage        []ArbitrageLeg `json:"arbitrage"`
	DetectedAt       time.Time      `json:"detected_at"`
	Notification     bool           `json:"notification"`
}
