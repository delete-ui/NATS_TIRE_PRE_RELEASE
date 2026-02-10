package constants

import "time"

// NATS Topics
const (
	//Main topics
	TopicMatchFound  = "events.match.found"
	TopicOddsUpdated = "events.odds.updated"
	TopicForkFound   = "events.fork.found"
	TopicAlerts      = "events.alerts"
	TopicErrors      = "events.errors"
	TopicHealth      = "events.health"

	//Topic commands
	TopicCommands = "commands.*"

	//JetStream
	StreamEvents   = "EVENTS"
	StreamAlerts   = "ALERTS"
	StreamCommands = "COMMANDS"

	//Consumers default
	ConsumerParser    = "parser-consumer"
	ConsumerArbitrage = "arbitrage-consumer"
	ConsumerNotifier  = "notifier-consumer"
	ConsumerMonitor   = "monitor-consumer"
)

// JetStream configurations
const (
	MaxMessageAge         = 24 * time.Hour
	MaxMessagesPerSubject = 100000
	Replicas              = 1
)

// Subscribe settings
const (
	DefaultAckWait = 30 * time.Second
	MaxDeliver     = 5
	MaxAckPending  = 256
	PullBatchSize  = 100
)

// Timeout settings
const (
	ConnectTimeout = 5 * time.Second
	RequestTimeout = 10 * time.Second
	ReconnectWait  = 1 * time.Second
	MaxReconnects  = -1
)

// Protocol version
const (
	ProtocolVersion = "1.0.0"
	EventVersion    = "1.0"
)
