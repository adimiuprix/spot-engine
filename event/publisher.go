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
