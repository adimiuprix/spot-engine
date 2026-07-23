package engine

import (
	"sync"
	"time"

	"github.com/adimiuprix/spot-engine/book"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/matcher"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/adimiuprix/spot-engine/queue"
	"github.com/shopspring/decimal"
)

type Engine struct {
	config       Config
	orderQueue   *queue.OrderQueue
	orderBook    *book.OrderBook
	matcher      *matcher.Matcher
	running      bool
	state        protocol.OrderBookState // Market state
	lastCmdSeqID uint64                  // Last processed command SeqID for replay
	mu           sync.RWMutex
	publisher    event.PublishLog
	seqGen       *event.SequenceGenerator
}

func New(config Config) *Engine {
	orderBook := book.NewOrderBook(config.Symbol)

	// Create event publisher (channel-based for now)
	publisher := event.NewChannelPublisher(10000)

	// Create sequence generator
	seqGen := event.NewSequenceGenerator(0)

	// Create matcher
	m := matcher.New(orderBook, seqGen, publisher)

	// Configure lot size (default if not specified)
	if config.MinLotSize.GreaterThan(decimal.Zero) {
		m.SetLotSize(config.MinLotSize)
	} else {
		m.SetLotSize(DefaultLotSize)
	}

	return &Engine{
		config:     config,
		orderQueue: queue.NewOrderQueue(config.RingBufferSize),
		orderBook:  orderBook,
		matcher:    m,
		running:    false,
		state:      protocol.StateRunning, // Start in running state
		publisher:  publisher,
		seqGen:     seqGen,
	}
}

func (e *Engine) Start() {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}
	e.running = true
	e.mu.Unlock()

	go e.processOrders()
}

func (e *Engine) Stop() {
	e.mu.Lock()
	e.running = false
	e.mu.Unlock()

	// Close publisher
	e.publisher.Close()
}

func (e *Engine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

func (e *Engine) SubmitOrder(o *order.Order) bool {
	return e.orderQueue.Push(o)
}

// Events returns the event log channel
func (e *Engine) Events() <-chan *event.OrderBookLog {
	if ch, ok := e.publisher.(*event.ChannelPublisher); ok {
		return ch.Channel()
	}
	return nil
}

func (e *Engine) processOrders() {
	ticker := time.NewTicker(1 * time.Microsecond)
	defer ticker.Stop()

	for {
		e.mu.RLock()
		running := e.running
		e.mu.RUnlock()

		if !running {
			break
		}

		<-ticker.C
		e.processNextOrder()
	}
}

func (e *Engine) processNextOrder() {
	o, ok := e.orderQueue.Pop()
	if !ok {
		return
	}

	// Check market state before processing
	e.mu.RLock()
	currentState := e.state
	e.mu.RUnlock()

	// State enforcement
	if !currentState.CanAcceptOrders() {
		// Reject order due to market state
		var reason string
		if currentState == protocol.StateSuspended {
			reason = "market is suspended"
		} else if currentState == protocol.StateHalted {
			reason = "market is halted"
		}

		rejectLog := event.NewRejectLog(
			e.seqGen.Next(),
			o.CommandID,
			o.UserID,
			o.Symbol,
			o.Timestamp,
			protocol.RejectReasonMarketSuspended,
			reason,
			o.OrderID,
		)
		e.publisher.Publish(rejectLog)
		return
	}

	// Process order based on its Time-In-Force (TIF) and Type
	result := e.matcher.ProcessWithTIF(o)

	// Events are already published by matcher
	_ = result

	// Add to book logic based on order type and TIF
	if !o.IsFilled() {
		// Market orders NEVER rest in book (always fully executed or rejected)
		if o.Type == order.Market {
			// Market orders don't rest - do nothing
			return
		}

		// Limit orders: handle based on TIF
		switch o.TIF {
		case order.GTC, order.PostOnly:
			// GTC and PostOnly can rest in book
			e.orderBook.Add(o)
		case order.IOC, order.FOK:
			// IOC and FOK never rest in book
			// Already handled in TIF logic
		}
	}

	// Update last processed command SeqID for replay checkpoint
	e.mu.Lock()
	e.lastCmdSeqID = o.ID
	e.mu.Unlock()
}

func (e *Engine) GetOrderBook() *book.OrderBook {
	return e.orderBook
}

// GetLastCmdSeqID returns the last processed command SeqID
// Used for replay checkpoint after snapshot restore
func (e *Engine) GetLastCmdSeqID() uint64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.lastCmdSeqID
}

// SetLastCmdSeqID sets the last processed command SeqID
// Used when restoring from snapshot
func (e *Engine) SetLastCmdSeqID(seqID uint64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.lastCmdSeqID = seqID
}

// GetMatcher returns the matcher for snapshot purposes
func (e *Engine) GetMatcher() *matcher.Matcher {
	return e.matcher
}

// GetState returns the current market state
func (e *Engine) GetState() protocol.OrderBookState {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.state
}

// SuspendMarket suspends the market (no new orders, can cancel existing)
func (e *Engine) SuspendMarket(reason string) {
	e.mu.Lock()
	oldState := e.state
	e.state = protocol.StateSuspended
	e.mu.Unlock()

	// Emit admin log
	adminLog := event.NewAdminLog(
		e.seqGen.Next(),
		"suspend-market",
		0, // No specific user
		e.config.Symbol,
		time.Now().UnixNano(),
		event.EventTypeMarketSuspended,
		oldState,
		protocol.StateSuspended,
		reason,
		nil,
	)
	e.publisher.Publish(adminLog)
}

// ResumeMarket resumes the market to running state
func (e *Engine) ResumeMarket(reason string) {
	e.mu.Lock()
	oldState := e.state
	e.state = protocol.StateRunning
	e.mu.Unlock()

	// Emit admin log
	adminLog := event.NewAdminLog(
		e.seqGen.Next(),
		"resume-market",
		0,
		e.config.Symbol,
		time.Now().UnixNano(),
		event.EventTypeMarketResumed,
		oldState,
		protocol.StateRunning,
		reason,
		nil,
	)
	e.publisher.Publish(adminLog)
}

// HaltMarket halts the market (emergency stop, no operations)
func (e *Engine) HaltMarket(reason string) {
	e.mu.Lock()
	oldState := e.state
	e.state = protocol.StateHalted
	e.mu.Unlock()

	// Emit admin log
	adminLog := event.NewAdminLog(
		e.seqGen.Next(),
		"halt-market",
		0,
		e.config.Symbol,
		time.Now().UnixNano(),
		event.EventTypeMarketHalted,
		oldState,
		protocol.StateHalted,
		reason,
		nil,
	)
	e.publisher.Publish(adminLog)
}
