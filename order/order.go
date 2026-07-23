package order

import (
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	ID        uint64
	UserID    uint64
	Symbol    string
	Side      Side
	Type      Type
	TIF       TimeInForce
	Price     decimal.Decimal
	Quantity  decimal.Decimal
	Filled    decimal.Decimal
	Timestamp int64
}

func NewOrder(
	id uint64,
	userID uint64,
	symbol string,
	side Side,
	orderType Type,
	tif TimeInForce,
	price decimal.Decimal,
	quantity decimal.Decimal,

) Order {

	return Order{
		ID:        id,
		UserID:    userID,
		Symbol:    symbol,
		Side:      side,
		Type:      orderType,
		TIF:       tif,
		Price:     price,
		Quantity:  quantity,
		Filled:    decimal.Zero,
		Timestamp: time.Now().UnixNano(),
	}
}

func (o *Order) Remaining() decimal.Decimal {
	return o.Quantity.Sub(o.Filled)
}

func (o *Order) IsFilled() bool {
	return o.Filled.Equal(o.Quantity)
}
