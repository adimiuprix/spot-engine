package matcher

import (
	"github.com/adimiuprix/spot-engine/book"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/adimiuprix/spot-engine/trade"
	"github.com/shopspring/decimal"
)

type Matcher struct {
	book    *book.OrderBook
	tradeID uint64
}

type Result struct {
	Trades []trade.Trade
}

func New(book *book.OrderBook) *Matcher {
	return &Matcher{
		book:    book,
		tradeID: 1,
	}
}

func (m *Matcher) Process(o *order.Order) Result {
	switch o.Side {
	case order.Buy:
		return m.matchBuy(o)
	case order.Sell:
		return m.matchSell(o)
	}
	return Result{}
}

func (m *Matcher) matchBuy(buy *order.Order) Result {
	result := Result{}

	for buy.Remaining().GreaterThan(decimal.Zero) {
		ask := m.book.BestAsk()
		if ask == nil {
			break
		}

		// Check if price crosses
		if buy.Price.LessThan(ask.Price) {
			break
		}

		// Get first order at this price level
		sell := ask.Head()
		if sell == nil {
			break
		}

		// Execute trade
		trade := m.execute(buy, sell)
		result.Trades = append(result.Trades, trade)

		// Remove filled orders from the ask side
		m.book.RemoveFilledOrders(order.Sell)
	}

	return result
}

func (m *Matcher) matchSell(sell *order.Order) Result {
	result := Result{}

	for sell.Remaining().GreaterThan(decimal.Zero) {
		bid := m.book.BestBid()
		if bid == nil {
			break
		}

		// Check if price crosses
		if sell.Price.GreaterThan(bid.Price) {
			break
		}

		// Get first order at this price level
		buy := bid.Head()
		if buy == nil {
			break
		}

		// Execute trade
		trade := m.execute(buy, sell)
		result.Trades = append(result.Trades, trade)

		// Remove filled orders from the bid side
		m.book.RemoveFilledOrders(order.Buy)
	}

	return result
}

func (m *Matcher) execute(buy *order.Order, sell *order.Order) trade.Trade {
	buyRemaining := buy.Remaining()
	sellRemaining := sell.Remaining()

	// Determine trade quantity (minimum of both remainings)
	var qty decimal.Decimal
	if buyRemaining.LessThan(sellRemaining) {
		qty = buyRemaining
	} else {
		qty = sellRemaining
	}

	// Price is always the resting order's price (maker price)
	// In this case, we use the sell price as it was resting first
	price := sell.Price

	// Update filled quantities
	buy.Filled = buy.Filled.Add(qty)
	sell.Filled = sell.Filled.Add(qty)

	// Create trade record
	t := trade.Trade{
		ID:          m.tradeID,
		Symbol:      buy.Symbol,
		BuyOrderID:  buy.ID,
		SellOrderID: sell.ID,
		Price:       price,
		Quantity:    qty,
		Timestamp:   buy.Timestamp, // Use incoming order timestamp
	}

	m.tradeID++

	return t
}
