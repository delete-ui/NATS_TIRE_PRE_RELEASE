package types

import "time"

type EventType string

const (
	EventTypeMatchFound   EventType = "match.found"
	EventTypeOddsUpdated  EventType = "odds.updated"
	EventTypeForkFound    EventType = "fork.found"
	EventTypeAlertCreated EventType = "alert.created"
	EventTypeHealthCheck  EventType = "health.check"
	EventTypeError        EventType = "error"
)

// EventHeader
type EventHeader struct {
	EventID       string    `json:"event_id"`
	EventType     EventType `json:"event_type"`
	Timestamp     time.Time `json:"timestamp"`
	Source        string    `json:"source"`
	Version       string    `json:"version"`
	CorrelationID string    `json:"correlation_id,omitempty"`
}

type MatchFoundEvent struct {
	EventHeader
	Payload Match `json:"payload"`
}

type ForkFoundEvent struct {
	EventHeader
	Payload Fork `json:"payload"`
}

type OddsUpdatedEvent struct {
	EventHeader
	Payload Odds `json:"payload"`
}

// AlertCreatedEvent
type AlertCreatedEvent struct {
	EventHeader
	Payload struct {
		ForkID    string    `json:"fork_id"`
		Channel   string    `json:"channel"` // telegram, web, etc
		Message   string    `json:"message"`
		Priority  string    `json:"priority"` // low, medium, high
		CreatedAt time.Time `json:"created_at"`
	} `json:"payload"`
}

// HealthCheckEvent
type HealthCheckEvent struct {
	EventHeader
	Payload struct {
		Service string                 `json:"service"`
		Status  string                 `json:"status"` // healthy, unhealthy
		Message string                 `json:"message,omitempty"`
		Metrics map[string]interface{} `json:"metrics,omitempty"`
	} `json:"payload"`
}

// ErrorEvent
type ErrorEvent struct {
	EventHeader
	Payload struct {
		Service     string `json:"service"`
		Operation   string `json:"operation"`
		Error       string `json:"error"`
		Stack       string `json:"stack,omitempty"`
		IsCritical  bool   `json:"is_critical"`
		Recoverable bool   `json:"recoverable"`
	} `json:"payload"`
}
