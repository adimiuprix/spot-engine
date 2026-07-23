package book

import (
	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

type PriceLevel struct {
	Price      decimal.Decimal
	Orders     []*order.Order
	Volume     decimal.Decimal
	OrderCount int // Total count of orders at this price level
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
	p.OrderCount = len(p.Orders)
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
	p.OrderCount = len(p.Orders)
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

// MoveToTail moves an order to the end of the queue (used for iceberg replenishment)
func (p *PriceLevel) MoveToTail(o *order.Order) {
	// Find and remove the order
	for i, existing := range p.Orders {
		if existing.ID == o.ID {
			// Remove from current position
			p.Orders = append(p.Orders[:i], p.Orders[i+1:]...)
			// Add to tail
			p.Orders = append(p.Orders, o)
			return
		}
	}
}

// ProcessReplenishments checks all orders for replenishment and moves them to tail
// Returns list of orders that were replenished
func (p *PriceLevel) ProcessReplenishments() []*order.Order {
	var replenished []*order.Order

	for _, o := range p.Orders {
		if o.NeedsReplenishment() {
			o.Replenish()
			replenished = append(replenished, o)
		}
	}

	// Move all replenished orders to tail
	for _, o := range replenished {
		p.MoveToTail(o)
	}

	return replenished
}

// RemoveOrder removes a specific order from the level
func (p *PriceLevel) RemoveOrder(target *order.Order) bool {
	for i, o := range p.Orders {
		if o.ID == target.ID {
			// Remove from slice
			p.Orders = append(p.Orders[:i], p.Orders[i+1:]...)
			// Update volume
			p.Volume = p.Volume.Sub(o.Remaining())
			if p.Volume.IsNegative() {
				p.Volume = decimal.Zero
			}
			p.OrderCount = len(p.Orders)
			return true
		}
	}
	return false
}
