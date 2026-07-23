package engine

import (
	"sync"
	"time"

	"github.com/adimiuprix/spot-engine/book"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/matcher"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/adimiuprix/spot-engine/queue"
)

type Engine struct {
	config     Config
	orderQueue *queue.OrderQueue
	orderBook  *book.OrderBook
	matcher    *matcher.Matcher
	running    bool
	mu         sync.RWMutex
	publisher  event.PublishLog
	seqGen     *event.SequenceGenerator
}

func New(config Config) *Engine {
	orderBook := book.NewOrderBook(config.Symbol)

	// Create event publisher (channel-based for now)
	publisher := event.NewChannelPublisher(10000)

	// Create sequence generator
	seqGen := event.NewSequenceGenerator(0)

	return &Engine{
		config:     config,
		orderQueue: queue.NewOrderQueue(config.RingBufferSize),
		orderBook:  orderBook,
		matcher:    matcher.New(orderBook, seqGen, publisher),
		running:    false,
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

		select {
		case <-ticker.C:
			e.processNextOrder()
		}
	}
}

func (e *Engine) processNextOrder() {
	o, ok := e.orderQueue.Pop()
	if !ok {
		return
	}

	// Match order dengan order book
	result := e.matcher.Process(o)

	// Events are already published by matcher
	_ = result

	// Jika order tidak fully filled, tambahkan ke order book
	if !o.IsFilled() {
		e.orderBook.Add(o)
	}
}

func (e *Engine) GetOrderBook() *book.OrderBook {
	return e.orderBook
}
