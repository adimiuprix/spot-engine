package protocol

// BaseCommand contains the shared header for all commands
// These fields are required for routing, audit, and deterministic replay
type BaseCommand struct {
	Type      CommandType `json:"type"`       // Command type for routing
	SeqID     uint64      `json:"seq_id"`     // Optional: upstream sequence ID for deduplication
	CommandID string      `json:"command_id"` // Required: unique command identifier
	UserID    uint64      `json:"user_id"`    // Required: actor identity
	MarketID  string      `json:"market_id"`  // Required: target market
	Timestamp int64       `json:"timestamp"`  // Required: logical event time (Unix nano)
}

// Validate checks if the base command has all required fields
func (b *BaseCommand) Validate() error {
	if b.CommandID == "" {
		return ErrInvalidCommandID
	}
	if b.Timestamp <= 0 {
		return ErrInvalidTimestamp
	}
	if b.MarketID == "" {
		return ErrInvalidMarketID
	}
	return nil
}
