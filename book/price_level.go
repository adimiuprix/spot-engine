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

func NewPriceLevel(price decimal.Decimal) *PriceLevel {
	return &PriceLevel{
		Price:  price,
		Orders: make([]*order.Order, 0),
		Volume: decimal.Zero,
	}
}

func (p *PriceLevel) Add(o *order.Order) {
	remaining := o.Remaining()
	p.Orders = append(p.Orders, o)
	p.Volume = p.Volume.Add(remaining)
}

func (p *PriceLevel) RemoveVolume(qty decimal.Decimal) {
	p.Volume = p.Volume.Sub(qty)
	if p.Volume.IsNegative() {
		p.Volume = decimal.Zero
	}
}

// RemoveFilledOrders removes all fully filled orders from the level
func (p *PriceLevel) RemoveFilledOrders() {
	var active []*order.Order
	recalcVolume := decimal.Zero

	for _, o := range p.Orders {
		if !o.IsFilled() {
			active = append(active, o)
			recalcVolume = recalcVolume.Add(o.Remaining())
		}
	}

	p.Orders = active
	p.Volume = recalcVolume
}

// IsEmpty returns true if there are no orders at this price level
func (p *PriceLevel) IsEmpty() bool {
	return len(p.Orders) == 0
}

// Head returns the first order (highest priority)
func (p *PriceLevel) Head() *order.Order {
	if len(p.Orders) == 0 {
		return nil
	}
	return p.Orders[0]
}
