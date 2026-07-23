package engine

import (
	"sync"
	"time"

	"github.com/adimiuprix/spot-engine/book"
	"github.com/adimiuprix/spot-engine/matcher"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/adimiuprix/spot-engine/queue"
	"github.com/adimiuprix/spot-engine/trade"
)

type Engine struct {
	config     Config
	orderQueue *queue.OrderQueue
	orderBook  *book.OrderBook
	matcher    *matcher.Matcher
	running    bool
	mu         sync.RWMutex
	trades     chan trade.Trade
}

func New(config Config) *Engine {
	orderBook := book.NewOrderBook(config.Symbol)

	return &Engine{
		config:     config,
		orderQueue: queue.NewOrderQueue(config.RingBufferSize),
		orderBook:  orderBook,
		matcher:    matcher.New(orderBook),
		running:    false,
		trades:     make(chan trade.Trade, 1000),
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
}

func (e *Engine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

func (e *Engine) SubmitOrder(o *order.Order) bool {
	return e.orderQueue.Push(o)
}

func (e *Engine) Trades() <-chan trade.Trade {
	return e.trades
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

	// Publish trades
	for _, t := range result.Trades {
		select {
		case e.trades <- t:
		default:
			// Trade channel full, skip
		}
	}

	// Jika order tidak fully filled, tambahkan ke order book
	if !o.IsFilled() {
		e.orderBook.Add(o)
	}
}

func (e *Engine) GetOrderBook() *book.OrderBook {
	return e.orderBook
}
