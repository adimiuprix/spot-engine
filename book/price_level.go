package book

import (
	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

type PriceLevel struct {
	Price  decimal.Decimal
	Orders []*order.Order
	Volume decimal.Decimal
}

func NewPriceLevel(
	price decimal.Decimal,
) *PriceLevel {
	return &PriceLevel{
		Price:  price,
		Volume: decimal.Zero,
	}
}

func (p *PriceLevel) Add(
	o *order.Order,
) {
	remaining :=
		o.Quantity.Sub(o.Filled)
	p.Orders =
		append(
			p.Orders,
			o,
		)
	p.Volume =
		p.Volume.Add(
			remaining,
		)
}

func (p *PriceLevel) RemoveVolume(
	qty decimal.Decimal,
) {

	p.Volume = p.Volume.Sub(qty)
	if p.Volume.IsNegative() {
		p.Volume = decimal.Zero
	}
}
