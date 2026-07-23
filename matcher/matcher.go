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

func New(
	book *book.OrderBook,
) *Matcher {
	return &Matcher{
		book:    book,
		tradeID: 1,
	}
}

func (m *Matcher) Process(
	o *order.Order,
) Result {
	switch o.Side {
	case order.Buy:
		return m.matchBuy(o)
	case order.Sell:
		return m.matchSell(o)
	}
	return Result{}
}

func (m *Matcher) matchBuy(
	buy *order.Order,
) Result {
	result := Result{}
	for buy.Remaining().GreaterThan(decimal.Zero) {
		ask := m.book.BestAsk()
		if ask == nil {
			break
		}

		if buy.Price.LessThan(ask.Price) {
			break
		}

		sell := ask.Orders[0]
		trade := m.execute(buy, sell)
		result.Trades = append(result.Trades, trade)
	}
	return result
}

func (m *Matcher) matchSell(
	sell *order.Order,
) Result {
	result := Result{}
	for sell.Remaining().GreaterThan(decimal.Zero) {
		bid := m.book.BestBid()
		if bid == nil {
			break
		}

		if sell.Price.GreaterThan(bid.Price) {
			break
		}

		buy := bid.Orders[0]
		trade := m.execute(buy, sell)
		result.Trades = append(result.Trades, trade)
	}
	return result
}

func (m *Matcher) execute(
	buy *order.Order,
	sell *order.Order,
) trade.Trade {
	buyRemaining := buy.Remaining()
	sellRemaining := sell.Remaining()

	var qty decimal.Decimal
	if buyRemaining.LessThan(sellRemaining) {
		qty = buyRemaining
	} else {
		qty = sellRemaining
	}

	price := sell.Price

	buy.Filled = buy.Filled.Add(qty)
	sell.Filled = sell.Filled.Add(qty)

	t := trade.Trade{
		ID:          m.tradeID,
		Symbol:      buy.Symbol,
		BuyOrderID:  buy.ID,
		SellOrderID: sell.ID,
		Price:       price,
		Quantity:    qty,
	}

	m.tradeID++

	return t
}
