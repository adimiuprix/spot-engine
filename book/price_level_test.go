package book

import (
	"fmt"
	"testing"

	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

// TestNewPriceLevel tests price level creation
func TestNewPriceLevel(t *testing.T) {
	price := decimal.NewFromInt(100)
	level := NewPriceLevel(price)

	if level == nil {
		t.Fatal("Expected non-nil price level")
	}

	if !level.Price.Equal(price) {
		t.Errorf("Expected price %v, got %v", price, level.Price)
	}

	if level.Volume.IsPositive() {
		t.Errorf("Expected zero volume, got %v", level.Volume)
	}

	if level.OrderCount != 0 {
		t.Errorf("Expected order count 0, got %d", level.OrderCount)
	}

	if len(level.Orders) != 0 {
		t.Errorf("Expected empty orders slice, got %d orders", len(level.Orders))
	}
}

// TestPriceLevel_Add tests adding orders to a level
func TestPriceLevel_Add(t *testing.T) {
	level := NewPriceLevel(decimal.NewFromInt(100))

	// Create test order
	o := order.NewOrder(
		1,
		"ORD-1",
		"CMD-1",
		101,
		"BTCUSD",
		order.Buy,
		order.Limit,
		order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1000,
	)

	level.Add(&o)

	// Verify order was added
	if len(level.Orders) != 1 {
		t.Fatalf("Expected 1 order, got %d", len(level.Orders))
	}

	if level.OrderCount != 1 {
		t.Errorf("Expected OrderCount 1, got %d", level.OrderCount)
	}

	// Verify volume
	expectedVolume := decimal.NewFromInt(10)
	if !level.Volume.Equal(expectedVolume) {
		t.Errorf("Expected volume %v, got %v", expectedVolume, level.Volume)
	}

	// Verify the order
	if level.Orders[0].OrderID != "ORD-1" {
		t.Errorf("Expected order ID ORD-1, got %s", level.Orders[0].OrderID)
	}
}

// TestPriceLevel_AddMultiple tests adding multiple orders (FIFO)
func TestPriceLevel_AddMultiple(t *testing.T) {
	level := NewPriceLevel(decimal.NewFromInt(100))

	// Add 3 orders
	for i := 1; i <= 3; i++ {
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
			decimal.NewFromInt(int64(i*10)),
			int64(1000+i),
		)
		level.Add(&o)
	}

	// Verify order count
	if len(level.Orders) != 3 {
		t.Fatalf("Expected 3 orders, got %d", len(level.Orders))
	}

	// Verify FIFO order (first added should be first in queue)
	if level.Orders[0].OrderID != "ORD-1" {
		t.Errorf("Expected first order ORD-1, got %s", level.Orders[0].OrderID)
	}
	if level.Orders[1].OrderID != "ORD-2" {
		t.Errorf("Expected second order ORD-2, got %s", level.Orders[1].OrderID)
	}
	if level.Orders[2].OrderID != "ORD-3" {
		t.Errorf("Expected third order ORD-3, got %s", level.Orders[2].OrderID)
	}

	// Verify total volume (10 + 20 + 30 = 60)
	expectedVolume := decimal.NewFromInt(60)
	if !level.Volume.Equal(expectedVolume) {
		t.Errorf("Expected volume %v, got %v", expectedVolume, level.Volume)
	}
}

// TestPriceLevel_RemoveVolume tests volume reduction
func TestPriceLevel_RemoveVolume(t *testing.T) {
	level := NewPriceLevel(decimal.NewFromInt(100))
	level.Volume = decimal.NewFromInt(100)

	// Remove 30
	level.RemoveVolume(decimal.NewFromInt(30))
	if !level.Volume.Equal(decimal.NewFromInt(70)) {
		t.Errorf("Expected volume 70, got %v", level.Volume)
	}

	// Remove more than available (should not go negative)
	level.RemoveVolume(decimal.NewFromInt(80))
	if !level.Volume.Equal(decimal.Zero) {
		t.Errorf("Expected volume 0 (clamped), got %v", level.Volume)
	}
}

// TestPriceLevel_RemoveFilledOrders tests removing filled orders
func TestPriceLevel_RemoveFilledOrders(t *testing.T) {
	level := NewPriceLevel(decimal.NewFromInt(100))

	// Add 3 orders
	orders := make([]*order.Order, 3)
	for i := 0; i < 3; i++ {
		o := order.NewOrder(
			uint64(i+1),
			fmt.Sprintf("ORD-%d", i+1),
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
		level.Add(&o)
	}

	// Fill the middle order completely
	orders[1].Filled = orders[1].Quantity

	// Remove filled orders
	level.RemoveFilledOrders()

	// Should have 2 orders left
	if len(level.Orders) != 2 {
		t.Fatalf("Expected 2 orders after removal, got %d", len(level.Orders))
	}

	if level.OrderCount != 2 {
		t.Errorf("Expected OrderCount 2, got %d", level.OrderCount)
	}

	// Verify the remaining orders are the right ones
	if level.Orders[0].OrderID != "ORD-1" {
		t.Errorf("Expected first remaining order ORD-1, got %s", level.Orders[0].OrderID)
	}
	if level.Orders[1].OrderID != "ORD-3" {
		t.Errorf("Expected second remaining order ORD-3, got %s", level.Orders[1].OrderID)
	}

	// Volume should be recalculated (2 orders * 10 each = 20)
	expectedVolume := decimal.NewFromInt(20)
	if !level.Volume.Equal(expectedVolume) {
		t.Errorf("Expected volume %v, got %v", expectedVolume, level.Volume)
	}
}

// TestPriceLevel_IsEmpty tests empty state detection
func TestPriceLevel_IsEmpty(t *testing.T) {
	level := NewPriceLevel(decimal.NewFromInt(100))

	// Should be empty initially
	if !level.IsEmpty() {
		t.Error("Expected empty level")
	}

	// Add an order
	o := order.NewOrder(
		1, "ORD-1", "CMD-1", 101, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100), decimal.NewFromInt(10), 1000,
	)
	level.Add(&o)

	// Should not be empty
	if level.IsEmpty() {
		t.Error("Expected non-empty level after adding order")
	}

	// Fill and remove the order
	o.Filled = o.Quantity
	level.RemoveFilledOrders()

	// Should be empty again
	if !level.IsEmpty() {
		t.Error("Expected empty level after removing filled order")
	}
}

// TestPriceLevel_Head tests getting the first order
func TestPriceLevel_Head(t *testing.T) {
	level := NewPriceLevel(decimal.NewFromInt(100))

	// Empty level should return nil
	if level.Head() != nil {
		t.Error("Expected nil head for empty level")
	}

	// Add orders
	o1 := order.NewOrder(1, "ORD-1", "CMD-1", 101, "BTCUSD", order.Buy, order.Limit, order.GTC, decimal.NewFromInt(100), decimal.NewFromInt(10), 1000)
	o2 := order.NewOrder(2, "ORD-2", "CMD-2", 102, "BTCUSD", order.Buy, order.Limit, order.GTC, decimal.NewFromInt(100), decimal.NewFromInt(20), 1001)

	level.Add(&o1)
	level.Add(&o2)

	// Head should be the first order
	head := level.Head()
	if head == nil {
		t.Fatal("Expected non-nil head")
	}

	if head.OrderID != "ORD-1" {
		t.Errorf("Expected head to be ORD-1, got %s", head.OrderID)
	}
}

// TestPriceLevel_MoveToTail tests moving order to end of queue
func TestPriceLevel_MoveToTail(t *testing.T) {
	level := NewPriceLevel(decimal.NewFromInt(100))

	// Add 3 orders
	orders := make([]*order.Order, 3)
	for i := 0; i < 3; i++ {
		o := order.NewOrder(
			uint64(i+1),
			fmt.Sprintf("ORD-%d", i+1),
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
		level.Add(&o)
	}

	// Initial order: ORD-1, ORD-2, ORD-3

	// Move ORD-2 to tail
	level.MoveToTail(orders[1])

	// New order: ORD-1, ORD-3, ORD-2
	if level.Orders[0].OrderID != "ORD-1" {
		t.Errorf("Position 0: expected ORD-1, got %s", level.Orders[0].OrderID)
	}
	if level.Orders[1].OrderID != "ORD-3" {
		t.Errorf("Position 1: expected ORD-3, got %s", level.Orders[1].OrderID)
	}
	if level.Orders[2].OrderID != "ORD-2" {
		t.Errorf("Position 2: expected ORD-2, got %s", level.Orders[2].OrderID)
	}

	// Move ORD-1 (head) to tail
	level.MoveToTail(orders[0])

	// New order: ORD-3, ORD-2, ORD-1
	if level.Orders[0].OrderID != "ORD-3" {
		t.Errorf("Position 0: expected ORD-3, got %s", level.Orders[0].OrderID)
	}
	if level.Orders[2].OrderID != "ORD-1" {
		t.Errorf("Position 2: expected ORD-1, got %s", level.Orders[2].OrderID)
	}
}

// TestPriceLevel_ProcessReplenishments tests iceberg replenishment
func TestPriceLevel_ProcessReplenishments(t *testing.T) {
	level := NewPriceLevel(decimal.NewFromInt(100))

	// Create iceberg order
	icebergOrder := order.NewOrder(
		1, "ICE-1", "CMD-1", 101, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(100), // Total quantity
		1000,
	)

	// Setup as iceberg with 10 visible
	err := icebergOrder.SetupIceberg(decimal.NewFromInt(10))
	if err != nil {
		t.Fatalf("Failed to setup iceberg: %v", err)
	}

	// Create regular order
	regularOrder := order.NewOrder(
		2, "ORD-2", "CMD-2", 102, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(20),
		1001,
	)

	level.Add(&icebergOrder)
	level.Add(&regularOrder)

	// Initial order: ICE-1, ORD-2

	// Fill the visible portion of iceberg
	icebergOrder.Filled = decimal.NewFromInt(10)

	// Process replenishments
	replenished := level.ProcessReplenishments()

	// Should have replenished the iceberg
	if len(replenished) != 1 {
		t.Fatalf("Expected 1 replenishment, got %d", len(replenished))
	}

	if replenished[0].OrderID != "ICE-1" {
		t.Errorf("Expected ICE-1 to be replenished, got %s", replenished[0].OrderID)
	}

	// Iceberg should be moved to tail
	// New order: ORD-2, ICE-1
	if level.Orders[0].OrderID != "ORD-2" {
		t.Errorf("Position 0: expected ORD-2, got %s", level.Orders[0].OrderID)
	}
	if level.Orders[1].OrderID != "ICE-1" {
		t.Errorf("Position 1: expected ICE-1, got %s", level.Orders[1].OrderID)
	}
}

// TestPriceLevel_RemoveOrder tests removing specific order
func TestPriceLevel_RemoveOrder(t *testing.T) {
	level := NewPriceLevel(decimal.NewFromInt(100))

	// Add 3 orders
	orders := make([]*order.Order, 3)
	for i := 0; i < 3; i++ {
		o := order.NewOrder(
			uint64(i+1),
			fmt.Sprintf("ORD-%d", i+1),
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
		level.Add(&o)
	}

	// Remove middle order
	removed := level.RemoveOrder(orders[1])
	if !removed {
		t.Error("Expected successful removal")
	}

	// Should have 2 orders left
	if len(level.Orders) != 2 {
		t.Fatalf("Expected 2 orders, got %d", len(level.Orders))
	}

	if level.OrderCount != 2 {
		t.Errorf("Expected OrderCount 2, got %d", level.OrderCount)
	}

	// Verify remaining orders
	if level.Orders[0].OrderID != "ORD-1" {
		t.Errorf("Expected ORD-1, got %s", level.Orders[0].OrderID)
	}
	if level.Orders[1].OrderID != "ORD-3" {
		t.Errorf("Expected ORD-3, got %s", level.Orders[1].OrderID)
	}

	// Volume should be reduced (30 - 10 = 20)
	expectedVolume := decimal.NewFromInt(20)
	if !level.Volume.Equal(expectedVolume) {
		t.Errorf("Expected volume %v, got %v", expectedVolume, level.Volume)
	}

	// Try to remove non-existent order
	fakeOrder := order.NewOrder(999, "FAKE", "FAKE", 999, "BTCUSD", order.Buy, order.Limit, order.GTC, decimal.NewFromInt(100), decimal.NewFromInt(10), 9999)
	removed = level.RemoveOrder(&fakeOrder)
	if removed {
		t.Error("Expected removal to fail for non-existent order")
	}
}

// TestPriceLevel_RemoveOrderVolumeClamping tests that volume doesn't go negative
func TestPriceLevel_RemoveOrderVolumeClamping(t *testing.T) {
	level := NewPriceLevel(decimal.NewFromInt(100))

	o := order.NewOrder(
		1, "ORD-1", "CMD-1", 101, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1000,
	)
	level.Add(&o)

	// Manually set volume lower than order remaining
	level.Volume = decimal.NewFromInt(5)

	// Remove order
	level.RemoveOrder(&o)

	// Volume should be clamped to zero
	if !level.Volume.Equal(decimal.Zero) {
		t.Errorf("Expected volume to be clamped to zero, got %v", level.Volume)
	}
}
