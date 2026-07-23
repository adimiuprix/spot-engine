package matcher

import (
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/shopspring/decimal"
)

// processMarket handles Market orders
// Market orders execute at any available price
// Supports two modes:
//  1. Base mode (Size-based): "Buy 1 BTC"
//  2. Quote mode (Budget-based): "Spend $50,000"
func (m *Matcher) processMarket(o *order.Order) Result {
	switch o.Side {
	case order.Buy:
		return m.matchMarketBuy(o)
	case order.Sell:
		return m.matchMarketSell(o)
	}
	return Result{}
}

// matchMarketBuy matches a market buy order
// Buys at any available ask price until quantity or budget is exhausted
func (m *Matcher) matchMarketBuy(buy *order.Order) Result {
	result := Result{}

	// Determine mode: Base (size-based) or Quote (budget-based)
	useQuoteMode := buy.Quantity.IsZero() && buy.QuoteSize.GreaterThan(decimal.Zero)
	remainingQuote := buy.QuoteSize // For quote mode tracking

	// For base mode, track remaining quantity
	remainingQty := buy.Quantity

	for {
		// Check if we still have quantity to fill
		if !useQuoteMode && remainingQty.LessThanOrEqual(decimal.Zero) {
			break // Base mode: all quantity filled
		}
		if useQuoteMode && remainingQuote.LessThanOrEqual(decimal.Zero) {
			break // Quote mode: all budget spent
		}

		// Get best ask (lowest sell price)
		ask := m.book.BestAsk()
		if ask == nil {
			// No liquidity available
			if useQuoteMode && remainingQuote.GreaterThan(decimal.Zero) {
				// Emit reject for remaining quote
				rejectLog := event.NewRejectLog(
					m.seqGen.Next(),
					buy.CommandID,
					buy.UserID,
					buy.Symbol,
					buy.Timestamp,
					protocol.RejectReasonNoLiquidity,
					"no liquidity available for market order",
					buy.OrderID,
				)
				result.Trades = append(result.Trades, rejectLog)
				m.publisher.Publish(rejectLog)
			}
			break
		}

		sell := ask.Head()
		if sell == nil {
			break
		}

		// Calculate match size
		var matchSize decimal.Decimal
		if useQuoteMode {
			// Quote mode: calculate how much we can buy with remaining budget
			maxSize := remainingQuote.Div(sell.Price)
			matchSize = maxSize
		} else {
			// Base mode: match remaining quantity
			matchSize = remainingQty
		}

		// Can't buy more than available
		if matchSize.GreaterThan(sell.Remaining()) {
			matchSize = sell.Remaining()
		}

		// ⚠️ SAFETY CHECK: Prevent infinite micro-remainder loop
		if matchSize.LessThan(m.lotSize) {
			// Match size below minimum trade unit
			if useQuoteMode && remainingQuote.GreaterThan(decimal.Zero) {
				// Reject remaining quote (can't execute below lotSize)
				rejectLog := event.NewRejectLog(
					m.seqGen.Next(),
					buy.CommandID,
					buy.UserID,
					buy.Symbol,
					buy.Timestamp,
					protocol.RejectReasonBelowMinLotSize,
					"remaining size below minimum trade unit",
					buy.OrderID,
				)
				result.Trades = append(result.Trades, rejectLog)
				m.publisher.Publish(rejectLog)
			}
			break
		}

		// Execute trade by manually updating filled amounts
		buy.Filled = buy.Filled.Add(matchSize)
		sell.Filled = sell.Filled.Add(matchSize)

		// Emit trade log
		tradeLog := event.NewTradeLog(
			m.seqGen.Next(),
			buy.CommandID,
			buy.UserID,
			buy.Symbol,
			buy.Timestamp,
			m.tradeID,
			sell.OrderID, // Maker
			buy.OrderID,  // Taker
			sell.Price,   // Maker price
			matchSize,
			"buy", // Taker side
		)
		result.Trades = append(result.Trades, tradeLog)
		m.publisher.Publish(tradeLog)

		// Emit fill logs
		makerFillLog := event.NewFillLog(
			m.seqGen.Next(),
			sell.CommandID,
			sell.UserID,
			sell.Symbol,
			buy.Timestamp,
			sell.OrderID,
			"sell",
			sell.Price,
			sell.Filled,
			sell.Remaining(),
			sell.IsFilled(),
		)
		result.Trades = append(result.Trades, makerFillLog)
		m.publisher.Publish(makerFillLog)

		takerFillLog := event.NewFillLog(
			m.seqGen.Next(),
			buy.CommandID,
			buy.UserID,
			buy.Symbol,
			buy.Timestamp,
			buy.OrderID,
			"buy",
			sell.Price,
			buy.Filled,
			decimal.Zero, // Market orders don't have predefined remaining
			false,
		)
		result.Trades = append(result.Trades, takerFillLog)
		m.publisher.Publish(takerFillLog)

		m.tradeID++

		// Update remaining budget (quote mode)
		if useQuoteMode {
			spent := matchSize.Mul(sell.Price)
			remainingQuote = remainingQuote.Sub(spent)
		} else {
			remainingQty = remainingQty.Sub(matchSize)
		}

		// Remove filled orders
		m.book.RemoveFilledOrders(order.Sell)
		m.processReplenishments(order.Sell)
	}

	return result
}

// matchMarketSell matches a market sell order
// Sells at any available bid price until quantity or budget is exhausted
func (m *Matcher) matchMarketSell(sell *order.Order) Result {
	result := Result{}

	// Determine mode: Base (size-based) or Quote (budget-based)
	useQuoteMode := sell.Quantity.IsZero() && sell.QuoteSize.GreaterThan(decimal.Zero)
	remainingQuote := sell.QuoteSize // For quote mode tracking

	// For base mode, track remaining quantity
	remainingQty := sell.Quantity

	for {
		// Check if we still have quantity to fill
		if !useQuoteMode && remainingQty.LessThanOrEqual(decimal.Zero) {
			break // Base mode: all quantity filled
		}
		if useQuoteMode && remainingQuote.LessThanOrEqual(decimal.Zero) {
			break // Quote mode: all budget received
		}

		// Get best bid (highest buy price)
		bid := m.book.BestBid()
		if bid == nil {
			// No liquidity available
			if useQuoteMode && remainingQuote.GreaterThan(decimal.Zero) {
				// Emit reject for remaining quote
				rejectLog := event.NewRejectLog(
					m.seqGen.Next(),
					sell.CommandID,
					sell.UserID,
					sell.Symbol,
					sell.Timestamp,
					protocol.RejectReasonNoLiquidity,
					"no liquidity available for market order",
					sell.OrderID,
				)
				result.Trades = append(result.Trades, rejectLog)
				m.publisher.Publish(rejectLog)
			}
			break
		}

		buy := bid.Head()
		if buy == nil {
			break
		}

		// Calculate match size
		var matchSize decimal.Decimal
		if useQuoteMode {
			// Quote mode: calculate how much we need to sell to get remaining quote
			maxSize := remainingQuote.Div(buy.Price)
			matchSize = maxSize
		} else {
			// Base mode: match remaining quantity
			matchSize = remainingQty
		}

		// Can't sell more than available
		if matchSize.GreaterThan(buy.Remaining()) {
			matchSize = buy.Remaining()
		}

		// ⚠️ SAFETY CHECK: Prevent infinite micro-remainder loop
		if matchSize.LessThan(m.lotSize) {
			// Match size below minimum trade unit
			if useQuoteMode && remainingQuote.GreaterThan(decimal.Zero) {
				// Reject remaining quote (can't execute below lotSize)
				rejectLog := event.NewRejectLog(
					m.seqGen.Next(),
					sell.CommandID,
					sell.UserID,
					sell.Symbol,
					sell.Timestamp,
					protocol.RejectReasonBelowMinLotSize,
					"remaining size below minimum trade unit",
					sell.OrderID,
				)
				result.Trades = append(result.Trades, rejectLog)
				m.publisher.Publish(rejectLog)
			}
			break
		}

		// Execute trade by manually updating filled amounts
		buy.Filled = buy.Filled.Add(matchSize)
		sell.Filled = sell.Filled.Add(matchSize)

		// Emit trade log
		tradeLog := event.NewTradeLog(
			m.seqGen.Next(),
			sell.CommandID,
			sell.UserID,
			sell.Symbol,
			sell.Timestamp,
			m.tradeID,
			buy.OrderID,  // Maker
			sell.OrderID, // Taker
			buy.Price,    // Maker price
			matchSize,
			"sell", // Taker side
		)
		result.Trades = append(result.Trades, tradeLog)
		m.publisher.Publish(tradeLog)

		// Emit fill logs
		makerFillLog := event.NewFillLog(
			m.seqGen.Next(),
			buy.CommandID,
			buy.UserID,
			buy.Symbol,
			sell.Timestamp,
			buy.OrderID,
			"buy",
			buy.Price,
			buy.Filled,
			buy.Remaining(),
			buy.IsFilled(),
		)
		result.Trades = append(result.Trades, makerFillLog)
		m.publisher.Publish(makerFillLog)

		takerFillLog := event.NewFillLog(
			m.seqGen.Next(),
			sell.CommandID,
			sell.UserID,
			sell.Symbol,
			sell.Timestamp,
			sell.OrderID,
			"sell",
			buy.Price,
			sell.Filled,
			decimal.Zero, // Market orders don't have predefined remaining
			false,
		)
		result.Trades = append(result.Trades, takerFillLog)
		m.publisher.Publish(takerFillLog)

		m.tradeID++

		// Update remaining budget (quote mode)
		if useQuoteMode {
			received := matchSize.Mul(buy.Price)
			remainingQuote = remainingQuote.Sub(received)
		} else {
			remainingQty = remainingQty.Sub(matchSize)
		}

		// Remove filled orders
		m.book.RemoveFilledOrders(order.Buy)
		m.processReplenishments(order.Buy)
	}

	return result
}
