package book

import (
	"github.com/google/btree"
	"github.com/shopspring/decimal"
)

// PriceTree provides an ordered index of price levels using B-Tree
// Supports both ascending (asks) and descending (bids) order
type PriceTree struct {
	tree       *btree.BTreeG[PriceItem]
	descending bool
}

// NewPriceTree creates a new price tree
// descending=true for bids (highest price first)
// descending=false for asks (lowest price first)
func NewPriceTree(descending bool) *PriceTree {
	return &PriceTree{
		tree: btree.NewG(
			32, // degree
			func(a, b PriceItem) bool {
				if descending {
					return a.Price.GreaterThan(b.Price)
				}
				return a.Price.LessThan(b.Price)
			},
		),
		descending: descending,
	}
}

// Add inserts or updates a price level in the tree
func (p *PriceTree) Add(level *PriceLevel) {
	item := PriceItem{
		Price: level.Price,
		Level: level,
	}
	p.tree.ReplaceOrInsert(item)
}

// Get retrieves a price level by exact price
func (p *PriceTree) Get(price decimal.Decimal) *PriceLevel {
	searchItem := PriceItem{Price: price}

	item, found := p.tree.Get(searchItem)
	if !found {
		return nil
	}
	return item.Level
}

// Remove deletes a price level from the tree
func (p *PriceTree) Remove(price decimal.Decimal) {
	searchItem := PriceItem{Price: price}
	p.tree.Delete(searchItem)
}

// Best returns the best price level (min for asks, max for bids)
func (p *PriceTree) Best() *PriceLevel {
	var result *PriceLevel

	// Min always returns the "smallest" item according to the comparator
	// For descending (bids), smallest means highest price
	// For ascending (asks), smallest means lowest price
	p.tree.Ascend(func(item PriceItem) bool {
		result = item.Level
		return false // Stop after first item
	})

	return result
}

// Len returns the number of price levels
func (p *PriceTree) Len() int {
	return p.tree.Len()
}

// Clear removes all price levels
func (p *PriceTree) Clear() {
	p.tree.Clear(false)
}

// Ascend iterates over all price levels in order
// For bids: highest to lowest price
// For asks: lowest to highest price
func (p *PriceTree) Ascend(iterator func(*PriceLevel) bool) {
	p.tree.Ascend(func(item PriceItem) bool {
		return iterator(item.Level)
	})
}

// DescendLessOrEqual iterates over price levels <= maxPrice in descending order
func (p *PriceTree) DescendLessOrEqual(maxPrice decimal.Decimal, iterator func(*PriceLevel) bool) {
	pivot := PriceItem{Price: maxPrice}
	p.tree.DescendLessOrEqual(pivot, func(item PriceItem) bool {
		return iterator(item.Level)
	})
}

// AscendGreaterOrEqual iterates over price levels >= minPrice in ascending order
func (p *PriceTree) AscendGreaterOrEqual(minPrice decimal.Decimal, iterator func(*PriceLevel) bool) {
	pivot := PriceItem{Price: minPrice}
	p.tree.AscendGreaterOrEqual(pivot, func(item PriceItem) bool {
		return iterator(item.Level)
	})
}
