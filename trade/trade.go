package trade

import (
	"github.com/shopspring/decimal"
)

type Trade struct {
	ID          uint64          `json:"id"`
	Symbol      string          `json:"symbol"`
	BuyOrderID  uint64          `json:"buy_order_id"`
	SellOrderID uint64          `json:"sell_order_id"`
	Price       decimal.Decimal `json:"price"`
	Quantity    decimal.Decimal `json:"quantity"`
	Timestamp   int64           `json:"timestamp"`
}
