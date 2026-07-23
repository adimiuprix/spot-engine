package engine

import (
	"sync"

	"github.com/adimiuprix/spot-engine/book"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/matcher"
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/shopspring/decimal"
)

// Market represents a single trading market with its own order book
type Market struct {
	ID         string
	OrderBook  *book.OrderBook
	Matcher    *matcher.Matcher
	State      protocol.OrderBookState
	MinLotSize decimal.Decimal
	mu         sync.RWMutex
	seqGen     *event.SequenceGenerator
	publisher  event.PublishLog
}

// NewMarket creates a new market
func NewMarket(
	id string,
	minLotSize decimal.Decimal,
	seqGen *event.SequenceGenerator,
	publisher event.PublishLog,
) *Market {
	orderBook := book.NewOrderBook(id)
	orderBook.MinLotSize = minLotSize

	return &Market{
		ID:         id,
		OrderBook:  orderBook,
		Matcher:    matcher.New(orderBook, seqGen, publisher),
		State:      protocol.StateRunning,
		MinLotSize: minLotSize,
		seqGen:     seqGen,
		publisher:  publisher,
	}
}

// GetState returns the current market state (thread-safe)
func (m *Market) GetState() protocol.OrderBookState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.State
}

// SetState sets the market state (thread-safe)
func (m *Market) SetState(state protocol.OrderBookState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.State = state
}

// CanPlaceOrder returns true if new orders can be placed
func (m *Market) CanPlaceOrder() bool {
	return m.GetState().CanPlaceOrder()
}

// CanCancelOrder returns true if orders can be cancelled
func (m *Market) CanCancelOrder() bool {
	return m.GetState().CanCancelOrder()
}

// CanAmendOrder returns true if orders can be amended
func (m *Market) CanAmendOrder() bool {
	return m.GetState().CanAmendOrder()
}

// UpdateConfig updates market configuration
func (m *Market) UpdateConfig(minLotSize decimal.Decimal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MinLotSize = minLotSize
	m.OrderBook.MinLotSize = minLotSize
}
