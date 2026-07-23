package engine

import (
	"context"
	"fmt"
	"sync"

	"github.com/adimiuprix/spot-engine/book"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/adimiuprix/spot-engine/snapshot"
)

// Alias for cleaner code
type OrderBookSnapshot = snapshot.OrderBookSnapshot

// MatchingEngine manages multiple markets and routes commands
type MatchingEngine struct {
	markets   map[string]*Market
	mu        sync.RWMutex
	seqGen    *event.SequenceGenerator
	publisher event.PublishLog
	running   bool
}

// NewMatchingEngine creates a new matching engine
func NewMatchingEngine(publisher event.PublishLog) *MatchingEngine {
	return &MatchingEngine{
		markets:   make(map[string]*Market),
		seqGen:    event.NewSequenceGenerator(0),
		publisher: publisher,
		running:   true,
	}
}

// CreateMarket creates a new market
func (e *MatchingEngine) CreateMarket(ctx context.Context, req *protocol.CreateMarketRequest) (*Future[bool], error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	future := NewFuture[bool]()

	// Execute asynchronously on the event loop
	go func() {
		e.mu.Lock()
		defer e.mu.Unlock()

		// Check if market already exists
		if _, exists := e.markets[req.MarketID]; exists {
			// Emit reject log
			rejectLog := event.NewRejectLog(
				e.seqGen.Next(),
				req.CommandID,
				req.UserID,
				req.MarketID,
				req.Timestamp,
				protocol.RejectReasonMarketAlreadyExists,
				"market already exists",
				"",
			)
			e.publisher.Publish(rejectLog)
			future.Fail(protocol.ErrNotFound) // Using existing error for now
			return
		}

		// Create market
		market := NewMarket(req.MarketID, req.MinLotSize, e.seqGen, e.publisher)
		e.markets[req.MarketID] = market

		// Emit admin log
		adminLog := event.NewAdminLog(
			e.seqGen.Next(),
			req.CommandID,
			req.UserID,
			req.MarketID,
			req.Timestamp,
			event.EventTypeMarketCreated,
			protocol.StateRunning, // Old state (new market starts running)
			protocol.StateRunning, // New state
			"",
			map[string]interface{}{
				"min_lot_size": req.MinLotSize.String(),
			},
		)
		e.publisher.Publish(adminLog)

		future.Complete(true)
	}()

	return future, nil
}

// SuspendMarket suspends trading on a market
func (e *MatchingEngine) SuspendMarket(ctx context.Context, req *protocol.SuspendMarketRequest) (*Future[bool], error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	future := NewFuture[bool]()

	// Execute asynchronously
	go func() {
		e.mu.RLock()
		market, exists := e.markets[req.MarketID]
		e.mu.RUnlock()

		if !exists {
			// Emit reject log
			rejectLog := event.NewRejectLog(
				e.seqGen.Next(),
				req.CommandID,
				req.UserID,
				req.MarketID,
				req.Timestamp,
				protocol.RejectReasonMarketNotFound,
				"market not found",
				"",
			)
			e.publisher.Publish(rejectLog)
			future.Fail(protocol.ErrNotFound)
			return
		}

		oldState := market.GetState()
		market.SetState(protocol.StateSuspended)

		// Emit admin log
		adminLog := event.NewAdminLog(
			e.seqGen.Next(),
			req.CommandID,
			req.UserID,
			req.MarketID,
			req.Timestamp,
			event.EventTypeMarketSuspended,
			oldState,
			protocol.StateSuspended,
			req.Reason,
			nil,
		)
		e.publisher.Publish(adminLog)

		future.Complete(true)
	}()

	return future, nil
}

// ResumeMarket resumes trading on a market
func (e *MatchingEngine) ResumeMarket(ctx context.Context, req *protocol.ResumeMarketRequest) (*Future[bool], error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	future := NewFuture[bool]()

	// Execute asynchronously
	go func() {
		e.mu.RLock()
		market, exists := e.markets[req.MarketID]
		e.mu.RUnlock()

		if !exists {
			// Emit reject log
			rejectLog := event.NewRejectLog(
				e.seqGen.Next(),
				req.CommandID,
				req.UserID,
				req.MarketID,
				req.Timestamp,
				protocol.RejectReasonMarketNotFound,
				"market not found",
				"",
			)
			e.publisher.Publish(rejectLog)
			future.Fail(protocol.ErrNotFound)
			return
		}

		oldState := market.GetState()
		market.SetState(protocol.StateRunning)

		// Emit admin log
		adminLog := event.NewAdminLog(
			e.seqGen.Next(),
			req.CommandID,
			req.UserID,
			req.MarketID,
			req.Timestamp,
			event.EventTypeMarketResumed,
			oldState,
			protocol.StateRunning,
			"",
			nil,
		)
		e.publisher.Publish(adminLog)

		future.Complete(true)
	}()

	return future, nil
}

// UpdateConfig updates market configuration
func (e *MatchingEngine) UpdateConfig(ctx context.Context, req *protocol.UpdateConfigRequest) (*Future[bool], error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	future := NewFuture[bool]()

	// Execute asynchronously
	go func() {
		e.mu.RLock()
		market, exists := e.markets[req.MarketID]
		e.mu.RUnlock()

		if !exists {
			// Emit reject log
			rejectLog := event.NewRejectLog(
				e.seqGen.Next(),
				req.CommandID,
				req.UserID,
				req.MarketID,
				req.Timestamp,
				protocol.RejectReasonMarketNotFound,
				"market not found",
				"",
			)
			e.publisher.Publish(rejectLog)
			future.Fail(protocol.ErrNotFound)
			return
		}

		// Update config
		oldMinLotSize := market.MinLotSize
		market.UpdateConfig(req.MinLotSize)

		// Emit admin log
		adminLog := event.NewAdminLog(
			e.seqGen.Next(),
			req.CommandID,
			req.UserID,
			req.MarketID,
			req.Timestamp,
			event.EventTypeMarketConfigUpdated,
			market.GetState(), // Keep same state
			market.GetState(), // Keep same state
			"",
			map[string]interface{}{
				"old_min_lot_size": oldMinLotSize.String(),
				"new_min_lot_size": req.MinLotSize.String(),
			},
		)
		e.publisher.Publish(adminLog)

		future.Complete(true)
	}()

	return future, nil
}

// GetMarket retrieves a market by ID
func (e *MatchingEngine) GetMarket(marketID string) (*Market, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	market, exists := e.markets[marketID]
	if !exists {
		return nil, fmt.Errorf("market %s not found", marketID)
	}

	return market, nil
}

// GetStats returns market statistics
func (e *MatchingEngine) GetStats(marketID string) (*protocol.MarketStats, error) {
	market, err := e.GetMarket(marketID)
	if err != nil {
		return nil, err
	}

	bestBid := market.OrderBook.BestBid()
	bestAsk := market.OrderBook.BestAsk()

	stats := &protocol.MarketStats{
		MarketID:   marketID,
		State:      market.GetState(),
		BidCount:   market.OrderBook.BidCount(),
		AskCount:   market.OrderBook.AskCount(),
		MinLotSize: market.MinLotSize,
	}

	if bestBid != nil {
		stats.BestBid = bestBid.Price
	}
	if bestAsk != nil {
		stats.BestAsk = bestAsk.Price
	}

	return stats, nil
}

// Shutdown gracefully shuts down the engine
func (e *MatchingEngine) Shutdown() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.running = false
	e.publisher.Close()
}

// TakeSnapshot creates a point-in-time snapshot of all markets
func (e *MatchingEngine) TakeSnapshot() ([]*OrderBookSnapshot, uint64) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var snapshots []*OrderBookSnapshot
	var globalMaxSeq uint64

	for marketID, market := range e.markets {
		snap := e.snapshotMarket(marketID, market)
		snapshots = append(snapshots, snap)

		if snap.LastCmdSeqID > globalMaxSeq {
			globalMaxSeq = snap.LastCmdSeqID
		}
	}

	return snapshots, globalMaxSeq
}

// snapshotMarket creates a snapshot of a single market
func (e *MatchingEngine) snapshotMarket(marketID string, market *Market) *OrderBookSnapshot {
	snap := &OrderBookSnapshot{
		MarketID:     marketID,
		SeqID:        e.seqGen.Current(),
		LastCmdSeqID: e.seqGen.Current(),
		TradeID:      market.Matcher.GetTradeID(),
		State:        market.GetState(),
		MinLotSize:   market.MinLotSize,
	}

	// Collect all bid orders
	market.OrderBook.BidTree.Ascend(func(level *book.PriceLevel) bool {
		for _, o := range level.Orders {
			snap.Bids = append(snap.Bids, o)
		}
		return true
	})

	// Collect all ask orders
	market.OrderBook.AskTree.Ascend(func(level *book.PriceLevel) bool {
		for _, o := range level.Orders {
			snap.Asks = append(snap.Asks, o)
		}
		return true
	})

	return snap
}

// RestoreFromSnapshot restores engine state from snapshots
func (e *MatchingEngine) RestoreFromSnapshot(snapshots []*OrderBookSnapshot) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Clear existing markets
	e.markets = make(map[string]*Market)

	// Restore each market
	for _, snap := range snapshots {
		market := NewMarket(snap.MarketID, snap.MinLotSize, e.seqGen, e.publisher)
		market.SetState(snap.State)

		// Restore bid orders
		for _, o := range snap.Bids {
			market.OrderBook.Add(o)
		}

		// Restore ask orders
		for _, o := range snap.Asks {
			market.OrderBook.Add(o)
		}

		e.markets[snap.MarketID] = market
	}

	// Update sequence generator
	maxSeq := uint64(0)
	for _, snap := range snapshots {
		if snap.LastCmdSeqID > maxSeq {
			maxSeq = snap.LastCmdSeqID
		}
	}
	e.seqGen.Set(maxSeq)

	return nil
}
