package book

import (
	"fmt"
	"testing"

	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

// TestNewOrderBook tests order book creation
func TestNewOrderBook(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	if ob == nil {
		t.Fatal("Expected non-nil order book")
	}

	if ob.Symbol != "BTCUSD" {
		t.Errorf("Expected symbol BTCUSD, got %s", ob.Symbol)
	}

	if ob.BidTree == nil {
		t.Error("Expected non-nil BidTree")
	}

	if ob.AskTree == nil {
		t.Error("Expected non-nil AskTree")
	}

	if ob.OrderIndex == nil {
		t.Error("Expected non-nil OrderIndex")
	}

	if !ob.MinLotSize.Equal(decimal.NewFromInt(1)) {
		t.Errorf("Expected MinLotSize 1, got %v", ob.MinLotSize)
	}
}

// TestOrderBook_AddBuyOrder tests adding buy orders
func TestOrderBook_AddBuyOrder(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	o := order.NewOrder(
		1, "BUY-1", "CMD-1", 101, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1000,
	)

	ob.Add(&o)

	// Verify bid tree has 1 level
	if ob.BidCount() != 1 {
		t.Errorf("Expected 1 bid level, got %d", ob.BidCount())
	}

	// Verify order is in index
	found := ob.FindOrder("BUY-1")
	if found == nil {
		t.Fatal("Expected to find order in index")
	}

	if found.OrderID != "BUY-1" {
		t.Errorf("Expected order BUY-1, got %s", found.OrderID)
	}

	// Verify best bid
	bestBid := ob.BestBid()
	if bestBid == nil {
		t.Fatal("Expected non-nil best bid")
	}

	if !bestBid.Price.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Expected best bid price 100, got %v", bestBid.Price)
	}
}

// TestOrderBook_AddSellOrder tests adding sell orders
func TestOrderBook_AddSellOrder(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	o := order.NewOrder(
		1, "SELL-1", "CMD-1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(105),
		decimal.NewFromInt(10),
		1000,
	)

	ob.Add(&o)

	// Verify ask tree has 1 level
	if ob.AskCount() != 1 {
		t.Errorf("Expected 1 ask level, got %d", ob.AskCount())
	}

	// Verify order is in index
	found := ob.FindOrder("SELL-1")
	if found == nil {
		t.Fatal("Expected to find order in index")
	}

	// Verify best ask
	bestAsk := ob.BestAsk()
	if bestAsk == nil {
		t.Fatal("Expected non-nil best ask")
	}

	if !bestAsk.Price.Equal(decimal.NewFromInt(105)) {
		t.Errorf("Expected best ask price 105, got %v", bestAsk.Price)
	}
}

// TestOrderBook_AddMultipleBuyOrders tests adding multiple buy orders at different prices
func TestOrderBook_AddMultipleBuyOrders(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	// Add buy orders at prices 100, 99, 101
	prices := []int64{100, 99, 101}
	for i, price := range prices {
		o := order.NewOrder(
			uint64(i+1),
			fmt.Sprintf("BUY-%d", i+1),
			fmt.Sprintf("CMD-%d", i+1),
			101,
			"BTCUSD",
			order.Buy,
			order.Limit,
			order.GTC,
			decimal.NewFromInt(price),
			decimal.NewFromInt(10),
			int64(1000+i),
		)
		ob.Add(&o)
	}

	// Verify 3 bid levels
	if ob.BidCount() != 3 {
		t.Errorf("Expected 3 bid levels, got %d", ob.BidCount())
	}

	// Best bid should be highest price (101)
	bestBid := ob.BestBid()
	if !bestBid.Price.Equal(decimal.NewFromInt(101)) {
		t.Errorf("Expected best bid 101, got %v", bestBid.Price)
	}
}

// TestOrderBook_AddMultipleSellOrders tests adding multiple sell orders at different prices
func TestOrderBook_AddMultipleSellOrders(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	// Add sell orders at prices 105, 106, 104
	prices := []int64{105, 106, 104}
	for i, price := range prices {
		o := order.NewOrder(
			uint64(i+1),
			fmt.Sprintf("SELL-%d", i+1),
			fmt.Sprintf("CMD-%d", i+1),
			101,
			"BTCUSD",
			order.Sell,
			order.Limit,
			order.GTC,
			decimal.NewFromInt(price),
			decimal.NewFromInt(10),
			int64(1000+i),
		)
		ob.Add(&o)
	}

	// Verify 3 ask levels
	if ob.AskCount() != 3 {
		t.Errorf("Expected 3 ask levels, got %d", ob.AskCount())
	}

	// Best ask should be lowest price (104)
	bestAsk := ob.BestAsk()
	if !bestAsk.Price.Equal(decimal.NewFromInt(104)) {
		t.Errorf("Expected best ask 104, got %v", bestAsk.Price)
	}
}

// TestOrderBook_AddOrdersSamePriceLevel tests adding multiple orders at the same price
func TestOrderBook_AddOrdersSamePriceLevel(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	// Add 3 buy orders at price 100
	for i := 0; i < 3; i++ {
		o := order.NewOrder(
			uint64(i+1),
			fmt.Sprintf("BUY-%d", i+1),
			fmt.Sprintf("CMD-%d", i+1),
			101,
			"BTCUSD",
			order.Buy,
			order.Limit,
			order.GTC,
			decimal.NewFromInt(100),
			decimal.NewFromInt(10),
			int64(1000+i),
		)
		ob.Add(&o)
	}

	// Should have 1 bid level
	if ob.BidCount() != 1 {
		t.Errorf("Expected 1 bid level, got %d", ob.BidCount())
	}

	// Verify the level has 3 orders
	level := ob.GetLevel(order.Buy, decimal.NewFromInt(100))
	if level == nil {
		t.Fatal("Expected to find price level")
	}

	if level.OrderCount != 3 {
		t.Errorf("Expected 3 orders in level, got %d", level.OrderCount)
	}

	// Verify total volume
	expectedVolume := decimal.NewFromInt(30)
	if !level.Volume.Equal(expectedVolume) {
		t.Errorf("Expected volume 30, got %v", level.Volume)
	}
}

// TestOrderBook_RemoveFilledOrders tests removing filled orders
func TestOrderBook_RemoveFilledOrders(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	// Add 3 buy orders at different prices
	orders := make([]*order.Order, 3)
	prices := []int64{100, 99, 101}
	for i, price := range prices {
		o := order.NewOrder(
			uint64(i+1),
			fmt.Sprintf("BUY-%d", i+1),
			fmt.Sprintf("CMD-%d", i+1),
			101,
			"BTCUSD",
			order.Buy,
			order.Limit,
			order.GTC,
			decimal.NewFromInt(price),
			decimal.NewFromInt(10),
			int64(1000+i),
		)
		orders[i] = &o
		ob.Add(&o)
	}

	// Fill the order at price 100 completely
	orders[0].Filled = orders[0].Quantity

	// Remove filled orders
	ob.RemoveFilledOrders(order.Buy)

	// Should have 2 bid levels left
	if ob.BidCount() != 2 {
		t.Errorf("Expected 2 bid levels after removal, got %d", ob.BidCount())
	}

	// Price level 100 should be gone
	level := ob.GetLevel(order.Buy, decimal.NewFromInt(100))
	if level != nil {
		t.Error("Expected price level 100 to be removed")
	}
}

// TestOrderBook_BestBidAsk tests best bid/ask retrieval
func TestOrderBook_BestBidAsk(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	// Empty book
	if ob.BestBid() != nil {
		t.Error("Expected nil best bid for empty book")
	}
	if ob.BestAsk() != nil {
		t.Error("Expected nil best ask for empty book")
	}

	// Add bids
	for _, price := range []int64{100, 99, 101} {
		o := order.NewOrder(
			uint64(price), fmt.Sprintf("BUY-%d", price), "CMD", 101, "BTCUSD",
			order.Buy, order.Limit, order.GTC,
			decimal.NewFromInt(price), decimal.NewFromInt(10), 1000,
		)
		ob.Add(&o)
	}

	// Add asks
	for _, price := range []int64{105, 104, 106} {
		o := order.NewOrder(
			uint64(price), fmt.Sprintf("SELL-%d", price), "CMD", 101, "BTCUSD",
			order.Sell, order.Limit, order.GTC,
			decimal.NewFromInt(price), decimal.NewFromInt(10), 1000,
		)
		ob.Add(&o)
	}

	// Best bid should be 101
	bestBid := ob.BestBid()
	if !bestBid.Price.Equal(decimal.NewFromInt(101)) {
		t.Errorf("Expected best bid 101, got %v", bestBid.Price)
	}

	// Best ask should be 104
	bestAsk := ob.BestAsk()
	if !bestAsk.Price.Equal(decimal.NewFromInt(104)) {
		t.Errorf("Expected best ask 104, got %v", bestAsk.Price)
	}
}

// TestOrderBook_GetDepth tests market depth retrieval
func TestOrderBook_GetDepth(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	// Add 5 bid levels
	for i := 100; i > 95; i-- {
		o := order.NewOrder(
			uint64(i), fmt.Sprintf("BUY-%d", i), "CMD", 101, "BTCUSD",
			order.Buy, order.Limit, order.GTC,
			decimal.NewFromInt(int64(i)), decimal.NewFromInt(10), 1000,
		)
		ob.Add(&o)
	}

	// Add 5 ask levels
	for i := 105; i < 110; i++ {
		o := order.NewOrder(
			uint64(i), fmt.Sprintf("SELL-%d", i), "CMD", 101, "BTCUSD",
			order.Sell, order.Limit, order.GTC,
			decimal.NewFromInt(int64(i)), decimal.NewFromInt(10), 1000,
		)
		ob.Add(&o)
	}

	// Get depth of 3 levels
	bids, asks := ob.GetDepth(3)

	if len(bids) != 3 {
		t.Errorf("Expected 3 bid levels, got %d", len(bids))
	}

	if len(asks) != 3 {
		t.Errorf("Expected 3 ask levels, got %d", len(asks))
	}

	// Verify bids are in descending order (best first)
	if !bids[0].Price.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Expected first bid 100, got %v", bids[0].Price)
	}
	if !bids[1].Price.Equal(decimal.NewFromInt(99)) {
		t.Errorf("Expected second bid 99, got %v", bids[1].Price)
	}
	if !bids[2].Price.Equal(decimal.NewFromInt(98)) {
		t.Errorf("Expected third bid 98, got %v", bids[2].Price)
	}

	// Verify asks are in ascending order (best first)
	if !asks[0].Price.Equal(decimal.NewFromInt(105)) {
		t.Errorf("Expected first ask 105, got %v", asks[0].Price)
	}
	if !asks[1].Price.Equal(decimal.NewFromInt(106)) {
		t.Errorf("Expected second ask 106, got %v", asks[1].Price)
	}
	if !asks[2].Price.Equal(decimal.NewFromInt(107)) {
		t.Errorf("Expected third ask 107, got %v", asks[2].Price)
	}
}

// TestOrderBook_GetLevel tests getting specific price level
func TestOrderBook_GetLevel(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	o := order.NewOrder(
		1, "BUY-1", "CMD-1", 101, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1000,
	)
	ob.Add(&o)

	// Get existing level
	level := ob.GetLevel(order.Buy, decimal.NewFromInt(100))
	if level == nil {
		t.Fatal("Expected to find level")
	}

	if !level.Price.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Expected price 100, got %v", level.Price)
	}

	// Get non-existent level
	level = ob.GetLevel(order.Buy, decimal.NewFromInt(99))
	if level != nil {
		t.Error("Expected nil for non-existent level")
	}
}

// TestOrderBook_RemoveLevel tests removing price level
func TestOrderBook_RemoveLevel(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	// Add orders at different prices
	for _, price := range []int64{100, 99, 101} {
		o := order.NewOrder(
			uint64(price), fmt.Sprintf("BUY-%d", price), "CMD", 101, "BTCUSD",
			order.Buy, order.Limit, order.GTC,
			decimal.NewFromInt(price), decimal.NewFromInt(10), 1000,
		)
		ob.Add(&o)
	}

	// Remove level at price 100
	ob.RemoveLevel(order.Buy, decimal.NewFromInt(100))

	// Should have 2 levels left
	if ob.BidCount() != 2 {
		t.Errorf("Expected 2 bid levels, got %d", ob.BidCount())
	}

	// Price level 100 should be gone
	level := ob.GetLevel(order.Buy, decimal.NewFromInt(100))
	if level != nil {
		t.Error("Expected price level 100 to be removed")
	}
}

// TestOrderBook_Clear tests clearing the order book
func TestOrderBook_Clear(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	// Add orders
	for _, price := range []int64{100, 99, 101} {
		o := order.NewOrder(
			uint64(price), fmt.Sprintf("BUY-%d", price), "CMD", 101, "BTCUSD",
			order.Buy, order.Limit, order.GTC,
			decimal.NewFromInt(price), decimal.NewFromInt(10), 1000,
		)
		ob.Add(&o)
	}

	for _, price := range []int64{105, 104, 106} {
		o := order.NewOrder(
			uint64(price), fmt.Sprintf("SELL-%d", price), "CMD", 101, "BTCUSD",
			order.Sell, order.Limit, order.GTC,
			decimal.NewFromInt(price), decimal.NewFromInt(10), 1000,
		)
		ob.Add(&o)
	}

	// Clear the book
	ob.Clear()

	// Verify both sides are empty
	if ob.BidCount() != 0 {
		t.Errorf("Expected 0 bid levels after clear, got %d", ob.BidCount())
	}

	if ob.AskCount() != 0 {
		t.Errorf("Expected 0 ask levels after clear, got %d", ob.AskCount())
	}

	if ob.BestBid() != nil {
		t.Error("Expected nil best bid after clear")
	}

	if ob.BestAsk() != nil {
		t.Error("Expected nil best ask after clear")
	}
}

// TestOrderBook_FindOrder tests finding order by OrderID
func TestOrderBook_FindOrder(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	o := order.NewOrder(
		1, "BUY-123", "CMD-1", 101, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1000,
	)
	ob.Add(&o)

	// Find existing order
	found := ob.FindOrder("BUY-123")
	if found == nil {
		t.Fatal("Expected to find order")
	}

	if found.OrderID != "BUY-123" {
		t.Errorf("Expected order BUY-123, got %s", found.OrderID)
	}

	// Find non-existent order
	found = ob.FindOrder("NONEXISTENT")
	if found != nil {
		t.Error("Expected nil for non-existent order")
	}
}

// TestOrderBook_RemoveOrder tests removing specific order
func TestOrderBook_RemoveOrder(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	// Add 3 orders at same price
	orders := make([]*order.Order, 3)
	for i := 0; i < 3; i++ {
		o := order.NewOrder(
			uint64(i+1),
			fmt.Sprintf("BUY-%d", i+1),
			fmt.Sprintf("CMD-%d", i+1),
			101,
			"BTCUSD",
			order.Buy,
			order.Limit,
			order.GTC,
			decimal.NewFromInt(100),
			decimal.NewFromInt(10),
			int64(1000+i),
		)
		orders[i] = &o
		ob.Add(&o)
	}

	// Remove middle order
	removed := ob.RemoveOrder(orders[1])
	if !removed {
		t.Error("Expected successful removal")
	}

	// Order should not be findable
	found := ob.FindOrder("BUY-2")
	if found != nil {
		t.Error("Expected order to be removed from index")
	}

	// Level should still exist with 2 orders
	level := ob.GetLevel(order.Buy, decimal.NewFromInt(100))
	if level == nil {
		t.Fatal("Expected level to still exist")
	}

	if level.OrderCount != 2 {
		t.Errorf("Expected 2 orders in level, got %d", level.OrderCount)
	}

	// Remove remaining orders
	ob.RemoveOrder(orders[0])
	ob.RemoveOrder(orders[2])

	// Level should be gone
	level = ob.GetLevel(order.Buy, decimal.NewFromInt(100))
	if level != nil {
		t.Error("Expected level to be removed when empty")
	}

	// Book should be empty
	if ob.BidCount() != 0 {
		t.Errorf("Expected empty book, got %d levels", ob.BidCount())
	}
}

// TestOrderBook_RemoveOrderNonExistent tests removing non-existent order
func TestOrderBook_RemoveOrderNonExistent(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	fakeOrder := order.NewOrder(
		999, "FAKE", "FAKE", 999, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		9999,
	)

	removed := ob.RemoveOrder(&fakeOrder)
	if removed {
		t.Error("Expected removal to fail for non-existent order")
	}
}

// TestOrderBook_OrderIndexConsistency tests that OrderIndex stays consistent
func TestOrderBook_OrderIndexConsistency(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	// Add orders
	for i := 1; i <= 5; i++ {
		o := order.NewOrder(
			uint64(i),
			fmt.Sprintf("ORD-%d", i),
			fmt.Sprintf("CMD-%d", i),
			101,
			"BTCUSD",
			order.Buy,
			order.Limit,
			order.GTC,
			decimal.NewFromInt(100),
			decimal.NewFromInt(10),
			int64(1000+i),
		)
		ob.Add(&o)
	}

	// Index should have 5 entries
	if len(ob.OrderIndex) != 5 {
		t.Errorf("Expected 5 entries in index, got %d", len(ob.OrderIndex))
	}

	// Find and remove an order
	o := ob.FindOrder("ORD-3")
	if o == nil {
		t.Fatal("Expected to find order")
	}

	ob.RemoveOrder(o)

	// Index should have 4 entries
	if len(ob.OrderIndex) != 4 {
		t.Errorf("Expected 4 entries in index after removal, got %d", len(ob.OrderIndex))
	}

	// Removed order should not be findable
	if ob.FindOrder("ORD-3") != nil {
		t.Error("Expected removed order to not be in index")
	}
}

// TestOrderBook_BidAskTreeSelection tests that orders go to correct tree
func TestOrderBook_BidAskTreeSelection(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	buyOrder := order.NewOrder(
		1, "BUY-1", "CMD-1", 101, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1000,
	)

	sellOrder := order.NewOrder(
		2, "SELL-1", "CMD-2", 102, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(105),
		decimal.NewFromInt(10),
		1001,
	)

	ob.Add(&buyOrder)
	ob.Add(&sellOrder)

	// Buy order should be in bid tree only
	if ob.BidCount() != 1 {
		t.Errorf("Expected 1 bid level, got %d", ob.BidCount())
	}

	if ob.AskCount() != 1 {
		t.Errorf("Expected 1 ask level, got %d", ob.AskCount())
	}

	// Verify trees have correct orders
	bidLevel := ob.BidTree.Get(decimal.NewFromInt(100))
	if bidLevel == nil || bidLevel.OrderCount != 1 {
		t.Error("Expected buy order in bid tree")
	}

	askLevel := ob.AskTree.Get(decimal.NewFromInt(105))
	if askLevel == nil || askLevel.OrderCount != 1 {
		t.Error("Expected sell order in ask tree")
	}
}
