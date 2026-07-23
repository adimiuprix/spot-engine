package book

import (
	"github.com/google/btree"
	"github.com/shopspring/decimal"
)

// PriceItem wraps a price level for storage in B-Tree
type PriceItem struct {
	Price decimal.Decimal
	Level *PriceLevel
}

// Less implements btree.Item interface for ascending order (asks)
// Returns true if this item should sort before the other item
func (p *PriceItem) Less(than btree.Item) bool {
	other := than.(*PriceItem)
	return p.Price.LessThan(other.Price)
}
