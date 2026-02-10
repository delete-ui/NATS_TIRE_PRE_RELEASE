package utils

import (
	"NATS_TIRE_SERVICE/shared/types"
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
		EventID: uuid.New().String(),
		EventType: eventType,
		Timestamp: time.Now().UTC(),
		Source: h.serviceName,
		Version: h.version,
		CorrelationID: correlationID,
	}
}
 
func (h *EventHelper) CreateMatchFoundEvent(match types.Match,correlationID string)types.MatchFoundEvent{
	return types.MatchFoundEvent{
		EventHeader: h.NewEventHeader(types.EventTypeMatchFound, correlationID),
		Payload: match,
	}
}
