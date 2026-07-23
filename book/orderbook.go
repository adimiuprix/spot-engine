package book

import (
	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

// OrderBook maintains buy and sell orders using efficient B-Tree structure
type OrderBook struct {
	Symbol     string
	BidTree    *PriceTree              // Descending order (highest price first)
	AskTree    *PriceTree              // Ascending order (lowest price first)
	MinLotSize decimal.Decimal         // Minimum trade unit for this market
	OrderIndex map[string]*order.Order // Fast order lookup by OrderID
}

// NewOrderBook creates a new order book for a symbol
func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol:     symbol,
		BidTree:    NewPriceTree(true),    // Descending for bids
		AskTree:    NewPriceTree(false),   // Ascending for asks
		MinLotSize: decimal.NewFromInt(1), // Default
		OrderIndex: make(map[string]*order.Order),
	}
}

// Add adds an order to the appropriate side of the book
func (b *OrderBook) Add(o *order.Order) {
	tree := b.getTree(o.Side)

	// Get or create price level
	level := tree.Get(o.Price)
	if level == nil {
		level = NewPriceLevel(o.Price)
		tree.Add(level)
	}

	level.Add(o)

	// Add to index
	b.OrderIndex[o.OrderID] = o
}

// RemoveFilledOrders removes filled orders and empty price levels
func (b *OrderBook) RemoveFilledOrders(side order.Side) {
	tree := b.getTree(side)

	// Collect empty price levels
	var emptyPrices []decimal.Decimal

	tree.Ascend(func(level *PriceLevel) bool {
		level.RemoveFilledOrders()
		if level.IsEmpty() {
			emptyPrices = append(emptyPrices, level.Price)
		}
		return true // Continue iteration
	})

	// Remove empty levels
	for _, price := range emptyPrices {
		tree.Remove(price)
	}
}

// BestAsk returns the lowest ask price level (O(1) with B-Tree)
func (b *OrderBook) BestAsk() *PriceLevel {
	return b.AskTree.Best()
}

// BestBid returns the highest bid price level (O(1) with B-Tree)
func (b *OrderBook) BestBid() *PriceLevel {
	return b.BidTree.Best()
}

// GetDepth returns market depth up to specified levels
func (b *OrderBook) GetDepth(levels int) (bids, asks []*PriceLevel) {
	count := 0
	b.BidTree.Ascend(func(level *PriceLevel) bool {
		if count >= levels {
			return false
		}
		bids = append(bids, level)
		count++
		return true
	})

	count = 0
	b.AskTree.Ascend(func(level *PriceLevel) bool {
		if count >= levels {
			return false
		}
		asks = append(asks, level)
		count++
		return true
	})

	return bids, asks
}

// GetLevel retrieves a specific price level
func (b *OrderBook) GetLevel(side order.Side, price decimal.Decimal) *PriceLevel {
	tree := b.getTree(side)
	return tree.Get(price)
}

// RemoveLevel removes a price level completely
func (b *OrderBook) RemoveLevel(side order.Side, price decimal.Decimal) {
	tree := b.getTree(side)
	tree.Remove(price)
}

// Clear removes all orders from the book
func (b *OrderBook) Clear() {
	b.BidTree.Clear()
	b.AskTree.Clear()
}

// BidCount returns the number of bid price levels
func (b *OrderBook) BidCount() int {
	return b.BidTree.Len()
}

// AskCount returns the number of ask price levels
func (b *OrderBook) AskCount() int {
	return b.AskTree.Len()
}

// getTree returns the appropriate tree for the given side
func (b *OrderBook) getTree(side order.Side) *PriceTree {
	if side == order.Buy {
		return b.BidTree
	}
	return b.AskTree
}

// FindOrder finds an order by OrderID
func (b *OrderBook) FindOrder(orderID string) *order.Order {
	return b.OrderIndex[orderID]
}

// RemoveOrder removes a specific order from the book
// Returns true if the order was found and removed
func (b *OrderBook) RemoveOrder(o *order.Order) bool {
	tree := b.getTree(o.Side)
	level := tree.Get(o.Price)
	if level == nil {
		return false
	}

	// Remove from price level
	removed := level.RemoveOrder(o)
	if !removed {
		return false
	}

	// Remove from index
	delete(b.OrderIndex, o.OrderID)

	// Remove empty price level
	if level.IsEmpty() {
		tree.Remove(o.Price)
	}

	return true
}
