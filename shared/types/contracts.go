package types

// EventHandler
type EventHandler interface {
	HandleMatchFound(event MatchFoundEvent) error
	HandleOddsUpdated(event OddsUpdatedEvent) error
	HandleForkFound(event ForkFoundEvent) error
	HandleError(event ErrorEvent) error
}

// Publisher
type Publisher interface {
	PublishMatchFound(match Match) error
	PublishOddsUpdated(odds Odds) error
	PublishForkFound(fork Fork) error
	PublishError(service, operation, err string, critical bool) error
	PublishHealthCheck(service, status, message string) error
}

// Consumer
type Consumer interface {
	SubscribeToMatches(handler EventHandler) error
	SubscribeToOdds(handler EventHandler) error
	SubscribeToForks(handler EventHandler) error
	SubscribeToErrors(handler EventHandler) error
	Unsubscribe() error
}

// Validator
type Validator interface {
	ValidateMatch(match Match) error
	ValidateOdds(odds Odds) error
	ValidateFork(fork Fork) error
	ValidateEvent(event interface{}) error
}
