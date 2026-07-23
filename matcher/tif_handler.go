package matcher

import (
	"github.com/adimiuprix/spot-engine/book"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/shopspring/decimal"
)

// ProcessWithTIF processes an order based on its Time-In-Force
func (m *Matcher) ProcessWithTIF(o *order.Order) Result {
	switch o.TIF {
	case order.GTC:
		return m.processGTC(o)
	case order.IOC:
		return m.processIOC(o)
	case order.FOK:
		return m.processFOK(o)
	case order.PostOnly:
		return m.processPostOnly(o)
	default:
		// Default to GTC
		return m.processGTC(o)
	}
}

// processGTC handles Good-Til-Cancel orders (same as current Process logic)
func (m *Matcher) processGTC(o *order.Order) Result {
	return m.Process(o)
}

// processIOC handles Immediate-Or-Cancel orders
// Matches what it can immediately, cancels the rest (doesn't rest in book)
func (m *Matcher) processIOC(o *order.Order) Result {
	result := Result{}

	// Try to match
	switch o.Side {
	case order.Buy:
		result = m.matchIOCBuy(o)
	case order.Sell:
		result = m.matchIOCSell(o)
	}

	// IOC never rests in book - any remaining is cancelled
	if !o.IsFilled() && o.Remaining().GreaterThan(decimal.Zero) {
		cancelLog := event.NewCancelLog(
			m.seqGen.Next(),
			o.CommandID,
			o.UserID,
			o.Symbol,
			o.Timestamp,
			o.OrderID,
			sideToString(o.Side),
			o.Price,
			o.Remaining(),
		)
		result.Trades = append(result.Trades, cancelLog)
		m.publisher.Publish(cancelLog)
	}

	return result
}

// matchIOCBuy matches an IOC buy order
func (m *Matcher) matchIOCBuy(buy *order.Order) Result {
	result := Result{}

	for buy.Remaining().GreaterThan(decimal.Zero) {
		ask := m.book.BestAsk()
		if ask == nil {
			// No liquidity - emit reject
			rejectLog := event.NewRejectLog(
				m.seqGen.Next(),
				buy.CommandID,
				buy.UserID,
				buy.Symbol,
				buy.Timestamp,
				protocol.RejectReasonNoLiquidity,
				"no liquidity available for IOC order",
				buy.OrderID,
			)
			result.Trades = append(result.Trades, rejectLog)
			m.publisher.Publish(rejectLog)
			break
		}

		// Check if price crosses
		if buy.Price.LessThan(ask.Price) {
			// Price mismatch - emit reject for remaining
			rejectLog := event.NewRejectLog(
				m.seqGen.Next(),
				buy.CommandID,
				buy.UserID,
				buy.Symbol,
				buy.Timestamp,
				protocol.RejectReasonInvalidPrice,
				"price does not cross for IOC order",
				buy.OrderID,
			)
			result.Trades = append(result.Trades, rejectLog)
			m.publisher.Publish(rejectLog)
			break
		}

		sell := ask.Head()
		if sell == nil {
			break
		}

		// Execute trade
		tradeLogs := m.execute(buy, sell)
		result.Trades = append(result.Trades, tradeLogs...)

		// Remove filled orders
		m.book.RemoveFilledOrders(order.Sell)
		m.processReplenishments(order.Sell)
	}

	return result
}

// matchIOCSell matches an IOC sell order
func (m *Matcher) matchIOCSell(sell *order.Order) Result {
	result := Result{}

	for sell.Remaining().GreaterThan(decimal.Zero) {
		bid := m.book.BestBid()
		if bid == nil {
			// No liquidity
			rejectLog := event.NewRejectLog(
				m.seqGen.Next(),
				sell.CommandID,
				sell.UserID,
				sell.Symbol,
				sell.Timestamp,
				protocol.RejectReasonNoLiquidity,
				"no liquidity available for IOC order",
				sell.OrderID,
			)
			result.Trades = append(result.Trades, rejectLog)
			m.publisher.Publish(rejectLog)
			break
		}

		if sell.Price.GreaterThan(bid.Price) {
			// Price mismatch
			rejectLog := event.NewRejectLog(
				m.seqGen.Next(),
				sell.CommandID,
				sell.UserID,
				sell.Symbol,
				sell.Timestamp,
				protocol.RejectReasonInvalidPrice,
				"price does not cross for IOC order",
				sell.OrderID,
			)
			result.Trades = append(result.Trades, rejectLog)
			m.publisher.Publish(rejectLog)
			break
		}

		buy := bid.Head()
		if buy == nil {
			break
		}

		tradeLogs := m.execute(buy, sell)
		result.Trades = append(result.Trades, tradeLogs...)

		m.book.RemoveFilledOrders(order.Buy)
		m.processReplenishments(order.Buy)
	}

	return result
}

// processFOK handles Fill-Or-Kill orders
// Must be fully filled immediately or rejected entirely
func (m *Matcher) processFOK(o *order.Order) Result {
	result := Result{}

	// Phase 1: Check if order can be fully filled
	canFill := m.checkFullFillPossible(o)

	if !canFill {
		// Cannot fully fill - reject entire order
		rejectLog := event.NewRejectLog(
			m.seqGen.Next(),
			o.CommandID,
			o.UserID,
			o.Symbol,
			o.Timestamp,
			protocol.RejectReasonInsufficientSize,
			"insufficient liquidity to fill entire FOK order",
			o.OrderID,
		)
		result.Trades = append(result.Trades, rejectLog)
		m.publisher.Publish(rejectLog)
		return result
	}

	// Phase 2: Execute (can reuse IOC logic since we know it will fully fill)
	return m.processIOC(o)
}

// checkFullFillPossible checks if an order can be fully filled at acceptable price
func (m *Matcher) checkFullFillPossible(o *order.Order) bool {
	remaining := o.Remaining()

	var tree *book.PriceTree
	if o.Side == order.Buy {
		tree = m.book.AskTree
	} else {
		tree = m.book.BidTree
	}

	// Iterate through price levels and check total liquidity
	canFill := false
	tree.Ascend(func(level *book.PriceLevel) bool {
		// Check if this price level is acceptable
		if o.Side == order.Buy && o.Price.LessThan(level.Price) {
			return false // Stop - price too high
		}
		if o.Side == order.Sell && o.Price.GreaterThan(level.Price) {
			return false // Stop - price too low
		}

		// Count liquidity at this level
		if remaining.LessThanOrEqual(level.Volume) {
			canFill = true
			return false // Found enough liquidity
		}

		remaining = remaining.Sub(level.Volume)
		return true // Continue to next level
	})

	return canFill
}

// processPostOnly handles Post-Only orders
// Order must become a maker (rest in book), rejects if would match immediately
func (m *Matcher) processPostOnly(o *order.Order) Result {
	result := Result{}

	// Check if order would match immediately
	var bestOpposite *book.PriceLevel
	if o.Side == order.Buy {
		bestOpposite = m.book.BestAsk()
	} else {
		bestOpposite = m.book.BestBid()
	}

	// Check if order would cross (match immediately)
	if bestOpposite != nil {
		wouldMatch := false
		if o.Side == order.Buy && o.Price.GreaterThanOrEqual(bestOpposite.Price) {
			wouldMatch = true
		}
		if o.Side == order.Sell && o.Price.LessThanOrEqual(bestOpposite.Price) {
			wouldMatch = true
		}

		if wouldMatch {
			// Reject - PostOnly cannot match immediately
			rejectLog := event.NewRejectLog(
				m.seqGen.Next(),
				o.CommandID,
				o.UserID,
				o.Symbol,
				o.Timestamp,
				protocol.RejectReasonPostOnlyWouldMatch,
				"PostOnly order would match immediately",
				o.OrderID,
			)
			result.Trades = append(result.Trades, rejectLog)
			m.publisher.Publish(rejectLog)
			return result
		}
	}

	// Safe to add to book (won't match)
	m.book.Add(o)

	return result
}
