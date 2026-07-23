package matcher

import (
	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

// processLimit handles Limit orders (GTC behavior)
// Limit orders have a specific price and will only match if price crosses
// Remaining quantity rests in the order book
func (m *Matcher) processLimit(o *order.Order) Result {
	switch o.Side {
	case order.Buy:
		return m.matchLimitBuy(o)
	case order.Sell:
		return m.matchLimitSell(o)
	}
	return Result{}
}

// matchLimitBuy matches a limit buy order
// Only matches with asks at or below the limit price
func (m *Matcher) matchLimitBuy(buy *order.Order) Result {
	result := Result{}

	for buy.Remaining().GreaterThan(decimal.Zero) {
		ask := m.book.BestAsk()
		if ask == nil {
			// No asks available - order will rest in book
			break
		}

		// LIMIT ORDER RULE: Only match if price crosses
		// Buy price must be >= ask price
		if buy.Price.LessThan(ask.Price) {
			// Price doesn't cross - order will rest in book
			break
		}

		// Get first order at this price level (FIFO)
		sell := ask.Head()
		if sell == nil {
			break
		}

		// Execute trade at maker price (ask price)
		tradeLogs := m.execute(buy, sell)
		result.Trades = append(result.Trades, tradeLogs...)

		// Remove filled orders from the ask side
		m.book.RemoveFilledOrders(order.Sell)

		// Process iceberg replenishments on ask side
		m.processReplenishments(order.Sell)
	}

	return result
}

// matchLimitSell matches a limit sell order
// Only matches with bids at or above the limit price
func (m *Matcher) matchLimitSell(sell *order.Order) Result {
	result := Result{}

	for sell.Remaining().GreaterThan(decimal.Zero) {
		bid := m.book.BestBid()
		if bid == nil {
			// No bids available - order will rest in book
			break
		}

		// LIMIT ORDER RULE: Only match if price crosses
		// Sell price must be <= bid price
		if sell.Price.GreaterThan(bid.Price) {
			// Price doesn't cross - order will rest in book
			break
		}

		// Get first order at this price level (FIFO)
		buy := bid.Head()
		if buy == nil {
			break
		}

		// Execute trade at maker price (bid price)
		tradeLogs := m.execute(buy, sell)
		result.Trades = append(result.Trades, tradeLogs...)

		// Remove filled orders from the bid side
		m.book.RemoveFilledOrders(order.Buy)

		// Process iceberg replenishments on bid side
		m.processReplenishments(order.Buy)
	}

	return result
}
