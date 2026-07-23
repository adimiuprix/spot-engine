package engine

import (
	"context"
	"fmt"
	"sync"

	"github.com/adimiuprix/spot-engine/book"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/adimiuprix/spot-engine/snapshot"
	"github.com/shopspring/decimal"
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

// PlaceOrderAsync places a new order asynchronously
func (e *MatchingEngine) PlaceOrderAsync(ctx context.Context, req *protocol.PlaceOrderRequest) (*Future[*protocol.PlaceOrderResult], error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	future := NewFuture[*protocol.PlaceOrderResult]()

	// Execute asynchronously
	go func() {
		e.mu.RLock()
		market, exists := e.markets[req.MarketID]
		e.mu.RUnlock()

		if !exists {
			rejectLog := event.NewRejectLog(
				e.seqGen.Next(),
				req.CommandID,
				req.UserID,
				req.MarketID,
				req.Timestamp,
				protocol.RejectReasonMarketNotFound,
				"market not found",
				req.OrderID,
			)
			e.publisher.Publish(rejectLog)
			future.Fail(protocol.ErrNotFound)
			return
		}

		if !market.CanPlaceOrder() {
			rejectLog := event.NewRejectLog(
				e.seqGen.Next(),
				req.CommandID,
				req.UserID,
				req.MarketID,
				req.Timestamp,
				protocol.RejectReasonMarketSuspended,
				"market does not accept new orders",
				req.OrderID,
			)
			e.publisher.Publish(rejectLog)
			future.Fail(protocol.ErrMarketSuspended)
			return
		}

		o := requestToOrder(req, e.seqGen)
		result := market.Matcher.ProcessWithTIF(o)

		if !o.IsFilled() && o.Type != order.Market {
			switch o.TIF {
			case order.GTC, order.PostOnly:
				market.OrderBook.Add(o)
			}
		}

		placeResult := &protocol.PlaceOrderResult{
			OrderID:     o.OrderID,
			Accepted:    true,
			Filled:      o.Filled,
			Remaining:   o.Remaining(),
			Trades:      convertLogsToInterface(result.Trades),
			InBook:      !o.IsFilled() && o.Type == order.Limit && (o.TIF == order.GTC || o.TIF == order.PostOnly),
			PartialFill: o.Filled.GreaterThan(decimal.Zero) && !o.IsFilled(),
		}

		future.Complete(placeResult)
	}()

	return future, nil
}

// CancelOrderAsync cancels an existing order asynchronously
func (e *MatchingEngine) CancelOrderAsync(ctx context.Context, req *protocol.CancelOrderRequest) (*Future[*protocol.CancelOrderResult], error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	future := NewFuture[*protocol.CancelOrderResult]()

	go func() {
		e.mu.RLock()
		market, exists := e.markets[req.Symbol]
		e.mu.RUnlock()

		if !exists {
			rejectLog := event.NewRejectLog(
				e.seqGen.Next(),
				req.CommandID,
				req.UserID,
				req.Symbol,
				req.Timestamp,
				protocol.RejectReasonMarketNotFound,
				"market not found",
				req.OrderID,
			)
			e.publisher.Publish(rejectLog)
			future.Fail(protocol.ErrNotFound)
			return
		}

		if !market.CanCancelOrder() {
			rejectLog := event.NewRejectLog(
				e.seqGen.Next(),
				req.CommandID,
				req.UserID,
				req.Symbol,
				req.Timestamp,
				protocol.RejectReasonMarketSuspended,
				"market does not accept cancellations",
				req.OrderID,
			)
			e.publisher.Publish(rejectLog)
			future.Fail(protocol.ErrMarketSuspended)
			return
		}

		o := market.OrderBook.FindOrder(req.OrderID)
		if o == nil {
			rejectLog := event.NewRejectLog(
				e.seqGen.Next(),
				req.CommandID,
				req.UserID,
				req.Symbol,
				req.Timestamp,
				protocol.RejectReasonOrderNotFound,
				"order not found in book",
				req.OrderID,
			)
			e.publisher.Publish(rejectLog)
			future.Fail(protocol.ErrNotFound)
			return
		}

		if o.UserID != req.UserID {
			rejectLog := event.NewRejectLog(
				e.seqGen.Next(),
				req.CommandID,
				req.UserID,
				req.Symbol,
				req.Timestamp,
				protocol.RejectReasonInvalidOrderOwner,
				"user does not own this order",
				req.OrderID,
			)
			e.publisher.Publish(rejectLog)
			future.Fail(protocol.ErrUnauthorized)
			return
		}

		removed := market.OrderBook.RemoveOrder(o)
		if !removed {
			future.Fail(protocol.ErrNotFound)
			return
		}

		cancelLog := event.NewCancelLog(
			e.seqGen.Next(),
			req.CommandID,
			req.UserID,
			req.Symbol,
			req.Timestamp,
			req.OrderID,
			sideToString(o.Side),
			o.Price,
			o.Remaining(),
		)
		e.publisher.Publish(cancelLog)

		cancelResult := &protocol.CancelOrderResult{
			OrderID:       o.OrderID,
			Cancelled:     true,
			FilledBefore:  o.Filled,
			CancelledSize: o.Remaining(),
		}

		future.Complete(cancelResult)
	}()

	return future, nil
}

// AmendOrderAsync amends an existing order asynchronously
func (e *MatchingEngine) AmendOrderAsync(ctx context.Context, req *protocol.AmendOrderRequest) (*Future[*protocol.AmendOrderResult], error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	future := NewFuture[*protocol.AmendOrderResult]()

	go func() {
		e.mu.RLock()
		market, exists := e.markets[req.MarketID]
		e.mu.RUnlock()

		if !exists {
			rejectLog := event.NewRejectLog(
				e.seqGen.Next(),
				req.CommandID,
				req.UserID,
				req.MarketID,
				req.Timestamp,
				protocol.RejectReasonMarketNotFound,
				"market not found",
				req.OrderID,
			)
			e.publisher.Publish(rejectLog)
			future.Fail(protocol.ErrNotFound)
			return
		}

		if !market.CanAmendOrder() {
			rejectLog := event.NewRejectLog(
				e.seqGen.Next(),
				req.CommandID,
				req.UserID,
				req.MarketID,
				req.Timestamp,
				protocol.RejectReasonMarketSuspended,
				"market does not accept amendments",
				req.OrderID,
			)
			e.publisher.Publish(rejectLog)
			future.Fail(protocol.ErrMarketSuspended)
			return
		}

		amendResult := market.Matcher.ProcessAmend(req)

		if !amendResult.Success {
			rejectLog := event.NewRejectLog(
				e.seqGen.Next(),
				req.CommandID,
				req.UserID,
				req.MarketID,
				req.Timestamp,
				amendResult.Reason,
				amendResult.Detail,
				req.OrderID,
			)
			e.publisher.Publish(rejectLog)

			var err error
			switch amendResult.Reason {
			case protocol.RejectReasonOrderNotFound:
				err = protocol.ErrNotFound
			case protocol.RejectReasonInvalidOrderOwner:
				err = protocol.ErrUnauthorized
			default:
				err = protocol.ErrInvalidRequest
			}

			future.Fail(err)
			return
		}

		result := &protocol.AmendOrderResult{
			OrderID:        req.OrderID,
			Amended:        true,
			NewPrice:       req.NewPrice,
			NewSize:        req.NewSize,
			Trades:         convertLogsToInterface(amendResult.Trades),
			MatchedOnAmend: len(amendResult.Trades) > 0,
		}

		future.Complete(result)
	}()

	return future, nil
}

// Helper functions
func requestToOrder(req *protocol.PlaceOrderRequest, seqGen *event.SequenceGenerator) *order.Order {
	var side order.Side
	if req.Side == "buy" {
		side = order.Buy
	} else {
		side = order.Sell
	}

	var orderType order.Type
	if req.OrderType == "market" {
		orderType = order.Market
	} else {
		orderType = order.Limit
	}

	o := &order.Order{
		ID:        seqGen.Next(),
		OrderID:   req.OrderID,
		CommandID: req.CommandID,
		UserID:    req.UserID,
		Symbol:    req.MarketID,
		Side:      side,
		Type:      orderType,
		TIF:       order.GTC,
		Price:     req.Price,
		Quantity:  req.Size,
		QuoteSize: req.QuoteSize,
		Filled:    decimal.Zero,
		Timestamp: req.Timestamp,
	}

	if req.VisibleSize.GreaterThan(decimal.Zero) {
		o.SetupIceberg(req.VisibleSize)
	}

	return o
}

func sideToString(side order.Side) string {
	if side == order.Buy {
		return "buy"
	}
	return "sell"
}

func convertLogsToInterface(logs []*event.OrderBookLog) []interface{} {
	result := make([]interface{}, len(logs))
	for i, log := range logs {
		result[i] = log
	}
	return result
}
