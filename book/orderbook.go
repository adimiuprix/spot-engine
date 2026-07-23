package book

import "github.com/adimiuprix/spot-engine/order"

type OrderBook struct {
	Symbol string
	Bids   map[string]*PriceLevel
	Asks   map[string]*PriceLevel
}

func NewOrderBook(
	symbol string,
) *OrderBook {
	return &OrderBook{
		Symbol: symbol,
		Bids:   make(map[string]*PriceLevel),
		Asks:   make(map[string]*PriceLevel),
	}
}

func (b *OrderBook) Add(
	o *order.Order,
) {
	key := o.Price.String()
	if o.Side == order.Buy {
		level, exists :=
			b.Bids[key]
		if !exists {
			level = &PriceLevel{
				Price: o.Price,
			}
			b.Bids[key] = level
		}
		level.Add(o)
	}
}
