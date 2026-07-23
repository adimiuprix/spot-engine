package book

import (
	"github.com/shopspring/decimal"
)

type PriceItem struct {
	Price decimal.Decimal
	Level *PriceLevel
}
