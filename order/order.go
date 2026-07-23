package order

import (
	"github.com/shopspring/decimal"
)

type Order struct {
	ID           uint64
	CommandID    string
	UserID       uint64
	Symbol       string
	Side         Side
	Type         Type
	TIF          TimeInForce
	Price        decimal.Decimal
	Quantity     decimal.Decimal
	Filled       decimal.Decimal
	Timestamp    int64
	VisibleLimit decimal.Decimal // For iceberg orders: max visible quantity
	HiddenSize   decimal.Decimal // For iceberg orders: remaining hidden quantity
}

func NewOrder(
	id uint64,
	commandID string,
	userID uint64,
	symbol string,
	side Side,
	orderType Type,
	tif TimeInForce,
	price decimal.Decimal,
	quantity decimal.Decimal,
	timestamp int64,
) Order {
	return Order{
		ID:           id,
		CommandID:    commandID,
		UserID:       userID,
		Symbol:       symbol,
		Side:         side,
		Type:         orderType,
		TIF:          tif,
		Price:        price,
		Quantity:     quantity,
		Filled:       decimal.Zero,
		Timestamp:    timestamp,
		VisibleLimit: decimal.Zero,
		HiddenSize:   decimal.Zero,
	}
}

func (order *Order) Remaining() decimal.Decimal {
	return order.Quantity.Sub(order.Filled)
}

func (order *Order) IsFilled() bool {
	return order.Filled.Equal(order.Quantity)
}

// IsIceberg returns true if this is an iceberg order
func (order *Order) IsIceberg() bool {
	return order.VisibleLimit.GreaterThan(decimal.Zero)
}

// VisibleQuantity returns the currently visible quantity for order book depth
func (order *Order) VisibleQuantity() decimal.Decimal {
	if !order.IsIceberg() {
		return order.Remaining()
	}

	remaining := order.Remaining()
	if remaining.LessThanOrEqual(order.VisibleLimit) {
		return remaining
	}
	return order.VisibleLimit
}
