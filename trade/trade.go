package trade

import (
	"github.com/shopspring/decimal"
)

type Trade struct {
	ID          uint64
	Symbol      string
	BuyOrderID  uint64
	SellOrderID uint64
	Price       decimal.Decimal
	Quantity    decimal.Decimal
	Timestamp   int64
}
