package event

import "sync/atomic"

// SequenceGenerator generates monotonically increasing sequence IDs
// Thread-safe for concurrent access
type SequenceGenerator struct {
	counter uint64
}

// NewSequenceGenerator creates a new sequence generator starting from the given value
func NewSequenceGenerator(start uint64) *SequenceGenerator {
	return &SequenceGenerator{
		counter: start,
	}
}

// Next returns the next sequence ID
func (s *SequenceGenerator) Next() uint64 {
	return atomic.AddUint64(&s.counter, 1)
}

// Current returns the current sequence ID without incrementing
func (s *SequenceGenerator) Current() uint64 {
	return atomic.LoadUint64(&s.counter)
}

// Set sets the sequence counter to a specific value
// Used during snapshot restore
func (s *SequenceGenerator) Set(value uint64) {
	atomic.StoreUint64(&s.counter, value)
}
