// Package order defines order types and related data structures.
//
// Orders are the fundamental unit of trading in the matching engine.
// Each order has a unique ID, side (buy/sell), type (limit/market),
// and time-in-force policy (GTC/IOC/FOK/PostOnly).
//
// Key features:
//   - Multiple order types (limit, market)
//   - Time-in-force policies (GTC, IOC, FOK, PostOnly)
//   - Iceberg order support
//   - Decimal precision (no float rounding)
//   - Deterministic timestamps
//
// Order lifecycle:
//  1. Create order with required fields
//  2. Validate order (price, size, etc.)
//  3. Submit to matching engine
//  4. Match against opposite side
//  5. Fill (partial or full)
//  6. Remove from book when fully filled
//
// Example usage:
//
//	order := &order.Order{
//	    ID:        1,
//	    OrderID:   "order-1",
//	    CommandID: "cmd-1",
//	    UserID:    1000,
//	    Symbol:    "BTC-USDT",
//	    Side:      order.Buy,
//	    Type:      order.Limit,
//	    TIF:       order.GTC,
//	    Price:     decimal.NewFromInt(50000),
//	    Quantity:  decimal.NewFromFloat(0.1),
//	    Filled:    decimal.Zero,
//	    Timestamp: time.Now().UnixNano(),
//	}
//
//	// Iceberg order
//	order.SetupIceberg(decimal.NewFromFloat(0.01)) // Show 0.01 at a time
package order

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type Order struct {
	ID           uint64
	OrderID      string // External order ID (from command)
	CommandID    string
	UserID       uint64
	Symbol       string
	Side         Side
	Type         Type
	TIF          TimeInForce
	Price        decimal.Decimal
	Quantity     decimal.Decimal
	QuoteSize    decimal.Decimal // For market orders in quote currency
	Filled       decimal.Decimal
	Timestamp    int64
	VisibleLimit decimal.Decimal // For iceberg orders: max visible quantity
	HiddenSize   decimal.Decimal // For iceberg orders: remaining hidden quantity
}

func NewOrder(
	id uint64,
	orderID string,
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
		OrderID:      orderID,
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

func (o *Order) Remaining() decimal.Decimal {
	return o.Quantity.Sub(o.Filled)
}

func (o *Order) IsFilled() bool {
	return o.Filled.Equal(o.Quantity)
}

// IsIceberg returns true if this is an iceberg order
func (o *Order) IsIceberg() bool {
	return o.VisibleLimit.GreaterThan(decimal.Zero)
}

// VisibleQuantity returns the currently visible quantity for order book depth
func (o *Order) VisibleQuantity() decimal.Decimal {
	if !o.IsIceberg() {
		return o.Remaining()
	}

	remaining := o.Remaining()
	visible := remaining.Sub(o.HiddenSize)

	// Ensure visible doesn't exceed remaining
	if visible.GreaterThan(remaining) {
		return remaining
	}

	// Ensure visible is non-negative
	if visible.LessThan(decimal.Zero) {
		return decimal.Zero
	}

	return visible
}

// NeedsReplenishment returns true if iceberg order's visible portion is depleted
func (o *Order) NeedsReplenishment() bool {
	if !o.IsIceberg() {
		return false
	}

	// Check if visible portion is depleted but hidden size remains
	visible := o.VisibleQuantity()
	return visible.LessThanOrEqual(decimal.Zero) && o.HiddenSize.GreaterThan(decimal.Zero)
}

// Replenish replenishes the visible portion from hidden size
// Returns the amount replenished
func (o *Order) Replenish() decimal.Decimal {
	if !o.NeedsReplenishment() {
		return decimal.Zero
	}

	// Calculate how much to replenish (min of VisibleLimit and HiddenSize)
	toReplenish := o.VisibleLimit
	if o.HiddenSize.LessThan(toReplenish) {
		toReplenish = o.HiddenSize
	}

	// Move from hidden to visible
	o.HiddenSize = o.HiddenSize.Sub(toReplenish)

	return toReplenish
}

// SetupIceberg initializes iceberg order state
// Called when placing an iceberg order
func (o *Order) SetupIceberg(visibleSize decimal.Decimal) error {
	if visibleSize.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("visible size must be positive")
	}

	if visibleSize.GreaterThan(o.Quantity) {
		return fmt.Errorf("visible size cannot exceed total quantity")
	}

	o.VisibleLimit = visibleSize
	// Initially, all quantity beyond visible limit is hidden
	o.HiddenSize = o.Quantity.Sub(visibleSize)

	return nil
}
