package event

// LogType represents the category of event being logged
type LogType string

const (
	LogTypeTrade  LogType = "trade"  // Successful trade execution
	LogTypeFill   LogType = "fill"   // Order fill notification
	LogTypeCancel LogType = "cancel" // Order cancellation
	LogTypeReject LogType = "reject" // Business rejection
	LogTypeAdmin  LogType = "admin"  // Management command success
)

func (l LogType) String() string {
	return string(l)
}

// EventType represents specific event subtypes
type EventType string

const (
	// Trade events
	EventTypeTrade EventType = "trade"

	// Fill events
	EventTypeFill        EventType = "fill"
	EventTypePartialFill EventType = "partial_fill"

	// Cancel events
	EventTypeCancel EventType = "cancel"

	// Reject events
	EventTypeReject EventType = "reject"

	// Admin events
	EventTypeMarketCreated       EventType = "market_created"
	EventTypeMarketSuspended     EventType = "market_suspended"
	EventTypeMarketResumed       EventType = "market_resumed"
	EventTypeMarketHalted        EventType = "market_halted"
	EventTypeMarketConfigUpdated EventType = "market_config_updated"
)

func (e EventType) String() string {
	return string(e)
}
