// Package event provides event logging and publishing for the matching engine.
//
// All state changes in the engine emit immutable events via the PublishLog interface.
// Events include sequence numbers for ordering and CommandID for idempotency.
// The event log provides a complete audit trail for compliance and replay.
//
// Key features:
//   - Immutable event log for audit trail
//   - Sequence number generation for ordering
//   - Multiple publisher implementations
//   - Non-blocking publish with overflow policy
//   - Type-safe event structures
//
// Event types:
//   - Trade: Match between two orders
//   - Fill: Order execution (partial or full)
//   - Cancel: Order cancellation
//   - Reject: Order rejection due to validation
//   - Admin: Market state changes
//   - Replenish: Iceberg order replenishment
//
// Example usage:
//
//	publisher := event.NewChannelPublisher(10000)
//	seqGen := event.NewSequenceGenerator(0)
//
//	// Listen to events
//	go func() {
//	    for log := range publisher.Channel() {
//	        switch log.LogType {
//	        case event.LogTypeTrade:
//	            fmt.Printf("Trade: %s @ %s\n", log.TradeQuantity, log.TradePrice)
//	        case event.LogTypeFill:
//	            fmt.Printf("Fill: Order %s filled %s\n", log.OrderID, log.FillQuantity)
//	        }
//	    }
//	}()
//
//	// Emit event
//	tradeLog := event.NewTradeLog(seqGen.Next(), ...)
//	publisher.Publish(tradeLog)
package event

// PublishLog is the interface for event emission
// Implementations can enrich events with local observation time (non-deterministic)
// but must not feed that back into replay or state-rebuild logic
type PublishLog interface {
	// Publish emits a canonical engine event
	// Returns error if publishing fails (e.g., queue full, network error)
	Publish(log *OrderBookLog) error

	// Close gracefully shuts down the publisher
	Close() error
}

// NoOpPublisher is a publisher that discards all events
type NoOpPublisher struct{}

func NewNoOpPublisher() *NoOpPublisher {
	return &NoOpPublisher{}
}

func (n *NoOpPublisher) Publish(log *OrderBookLog) error {
	// Discard
	return nil
}

func (n *NoOpPublisher) Close() error {
	return nil
}

// ChannelPublisher publishes events to a Go channel
type ChannelPublisher struct {
	ch chan *OrderBookLog
}

func NewChannelPublisher(bufferSize int) *ChannelPublisher {
	return &ChannelPublisher{
		ch: make(chan *OrderBookLog, bufferSize),
	}
}

func (c *ChannelPublisher) Publish(log *OrderBookLog) error {
	select {
	case c.ch <- log:
		return nil
	default:
		// Channel full - could return error or block
		// For now, drop to prevent blocking the engine
		return nil
	}
}

func (c *ChannelPublisher) Channel() <-chan *OrderBookLog {
	return c.ch
}

func (c *ChannelPublisher) Close() error {
	close(c.ch)
	return nil
}

// MultiPublisher publishes to multiple publishers
type MultiPublisher struct {
	publishers []PublishLog
}

func NewMultiPublisher(publishers ...PublishLog) *MultiPublisher {
	return &MultiPublisher{
		publishers: publishers,
	}
}

func (m *MultiPublisher) Publish(log *OrderBookLog) error {
	for _, pub := range m.publishers {
		if err := pub.Publish(log); err != nil {
			// Continue publishing to other publishers even if one fails
			// Could add error collection here
			continue
		}
	}
	return nil
}

func (m *MultiPublisher) Close() error {
	for _, pub := range m.publishers {
		_ = pub.Close()
	}
	return nil
}
