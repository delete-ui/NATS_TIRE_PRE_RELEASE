package utils

import (
	"NATS_TIRE_SERVICE/shared/types"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type EventHelper struct {
	serviceName string
	version     string
}

func NewEventHelper(serviceName, version string) *EventHelper {
	return &EventHelper{
		serviceName: serviceName,
		version:     version,
	}
}

func (h *EventHelper) NewEventHeader(eventType types.EventType, correlationID string) types.EventHeader {
	return types.EventHeader{
		EventID:       uuid.New().String(),
		EventType:     eventType,
		Timestamp:     time.Now().UTC(),
		Source:        h.serviceName,
		Version:       h.version,
		CorrelationID: correlationID,
	}
}

func (h *EventHelper) CreateMatchFoundEvent(match types.Match, correlationID string) types.MatchFoundEvent {
	return types.MatchFoundEvent{
		EventHeader: h.NewEventHeader(types.EventTypeMatchFound, correlationID),
		Payload:     match,
	}
}

func (h *EventHelper) CreateOddsUpdatedEvent(odds types.Odds, correlationID string) types.OddsUpdatedEvent {
	return types.OddsUpdatedEvent{
		EventHeader: h.NewEventHeader(types.EventTypeOddsUpdated, correlationID),
		Payload:     odds,
	}
}

func (h *EventHelper) CreateErrorEvent(operation, err string, critical bool, correlationID string) types.ErrorEvent {
	return types.ErrorEvent{
		EventHeader: h.NewEventHeader(types.EventTypeError, correlationID),
		Payload: struct {
			Service     string `json:"service"`
			Operation   string `json:"operation"`
			Error       string `json:"error"`
			Stack       string `json:"stack,omitempty"`
			IsCritical  bool   `json:"is_critical"`
			Recoverable bool   `json:"recoverable"`
		}{
			Service:     h.serviceName,
			Operation:   operation,
			Error:       err,
			IsCritical:  critical,
			Recoverable: !critical,
		},
	}
}

func (h *EventHelper) CreateHealthCheckEvent(status, message string, metrics map[string]interface{}) types.HealthCheckEvent {
	return types.HealthCheckEvent{
		EventHeader: h.NewEventHeader(types.EventTypeHealthCheck, ""),
		Payload: struct {
			Service string                 `json:"service"`
			Status  string                 `json:"status"`
			Message string                 `json:"message,omitempty"`
			Metrics map[string]interface{} `json:"metrics,omitempty"`
		}{
			Service: h.serviceName,
			Status:  status,
			Message: message,
			Metrics: metrics,
		},
	}
}

func (h *EventHelper) SerializeEvent(event interface{}) ([]byte, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}
	return data, nil
}

func (h *EventHelper) DeserializeEvent(data []byte, eventType types.EventType) (interface{}, error) {
	switch eventType {
	case types.EventTypeMatchFound:
		var event types.MatchFoundEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return event, nil

	case types.EventTypeOddsUpdated:
		var event types.OddsUpdatedEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return event, nil

	case types.EventTypeForkFound:
		var event types.ForkFoundEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return event, nil

	case types.EventTypeError:
		var event types.ErrorEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return event, nil

	case types.EventTypeHealthCheck:
		var event types.HealthCheckEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return event, nil

	default:
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}
}

func (h *EventHelper) CreateForkFoundEvent(fork types.Fork, correlationID string) types.ForkFoundEvent {
	return types.ForkFoundEvent{
		EventHeader: h.NewEventHeader(types.EventTypeForkFound, correlationID),
		Payload:     fork,
	}
}
