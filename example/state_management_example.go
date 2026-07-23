package main

import (
	"fmt"
	"time"

	"github.com/adimiuprix/spot-engine/book"
	"github.com/adimiuprix/spot-engine/engine"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("=== Market State Management Example ===\n")

	// Create engine
	config := engine.Config{
		Symbol:         "BTCUSD",
		RingBufferSize: 1024,
		MinLotSize:     decimal.NewFromFloat(0.001),
	}
	eng := engine.New(config)
	eng.Start()
	defer eng.Stop()

	// Monitor events
	go monitorEvents(eng.Events())

	time.Sleep(100 * time.Millisecond)

	// ========================================
	// Test 1: Normal Operation (Running State)
	// ========================================
	fmt.Println("=== Test 1: Normal Operation (Running) ===")
	fmt.Printf("Current state: %s\n", eng.GetState())

	limitOrder1 := createLimitOrder("ORDER-1", 1, 1, "BTCUSD", order.Buy, 50000, 1.0)
	eng.SubmitOrder(limitOrder1)

	time.Sleep(200 * time.Millisecond)
	fmt.Println("✅ Order accepted in Running state\n")

	// ========================================
	// Test 2: Suspend Market
	// ========================================
	fmt.Println("=== Test 2: Suspend Market ===")
	eng.SuspendMarket("scheduled maintenance")
	fmt.Printf("State changed to: %s\n", eng.GetState())

	time.Sleep(200 * time.Millisecond)

	// Try to place order in suspended state
	fmt.Println("\nTrying to place order in Suspended state...")
	limitOrder2 := createLimitOrder("ORDER-2", 2, 2, "BTCUSD", order.Buy, 50100, 0.5)
	eng.SubmitOrder(limitOrder2)

	time.Sleep(200 * time.Millisecond)
	fmt.Println("Expected: Order REJECTED (market suspended)\n")

	// ========================================
	// Test 3: Resume Market
	// ========================================
	fmt.Println("=== Test 3: Resume Market ===")
	eng.ResumeMarket("maintenance complete")
	fmt.Printf("State changed to: %s\n", eng.GetState())

	time.Sleep(200 * time.Millisecond)

	// Try to place order after resume
	fmt.Println("\nPlacing order after resume...")
	limitOrder3 := createLimitOrder("ORDER-3", 3, 3, "BTCUSD", order.Sell, 50200, 0.8)
	eng.SubmitOrder(limitOrder3)

	time.Sleep(200 * time.Millisecond)
	fmt.Println("✅ Order accepted after resume\n")

	// ========================================
	// Test 4: Halt Market (Emergency)
	// ========================================
	fmt.Println("=== Test 4: Halt Market (Emergency) ===")
	eng.HaltMarket("emergency stop - suspicious activity detected")
	fmt.Printf("State changed to: %s\n", eng.GetState())

	time.Sleep(200 * time.Millisecond)

	// Try to place order in halted state
	fmt.Println("\nTrying to place order in Halted state...")
	limitOrder4 := createLimitOrder("ORDER-4", 4, 4, "BTCUSD", order.Buy, 49900, 1.0)
	eng.SubmitOrder(limitOrder4)

	time.Sleep(200 * time.Millisecond)
	fmt.Println("Expected: Order REJECTED (market halted)\n")

	// ========================================
	// Test 5: Resume from Halt
	// ========================================
	fmt.Println("=== Test 5: Resume from Halt ===")
	eng.ResumeMarket("issue resolved, resuming trading")
	fmt.Printf("State changed to: %s\n", eng.GetState())

	time.Sleep(200 * time.Millisecond)

	// Place order after resume from halt
	fmt.Println("\nPlacing order after resume from halt...")
	limitOrder5 := createLimitOrder("ORDER-5", 5, 5, "BTCUSD", order.Buy, 49800, 0.5)
	eng.SubmitOrder(limitOrder5)

	time.Sleep(200 * time.Millisecond)
	fmt.Println("✅ Order accepted after resume\n")

	// Wait for all events
	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n--- Final Orderbook State ---")
	printOrderBook(eng)

	fmt.Println("\n=== State Management Example Complete ===")
}

func createLimitOrder(orderID string, id uint64, userID uint64, symbol string, side order.Side, price float64, qty float64) *order.Order {
	return &order.Order{
		ID:        id,
		OrderID:   orderID,
		CommandID: fmt.Sprintf("cmd-%d", id),
		UserID:    userID,
		Symbol:    symbol,
		Side:      side,
		Type:      order.Limit,
		TIF:       order.GTC,
		Price:     decimal.NewFromFloat(price),
		Quantity:  decimal.NewFromFloat(qty),
		Filled:    decimal.Zero,
		Timestamp: time.Now().UnixNano(),
	}
}

func monitorEvents(events <-chan *event.OrderBookLog) {
	for log := range events {
		switch log.LogType {
		case event.LogTypeReject:
			fmt.Printf("  [REJECT] Order %s: %s - %s\n",
				log.OrderID,
				log.RejectReason,
				log.RejectDetail,
			)

		case event.LogTypeAdmin:
			fmt.Printf("  [ADMIN] %s: %s → %s (%s)\n",
				log.EventType,
				log.OldState,
				log.NewState,
				log.AdminReason,
			)
		}
	}
}

func printOrderBook(eng *engine.Engine) {
	bk := eng.GetOrderBook()

	fmt.Println("Asks (Sell Orders):")
	count := 0
	bk.AskTree.Ascend(func(level *book.PriceLevel) bool {
		if level.Volume.GreaterThan(decimal.Zero) && count < 5 {
			fmt.Printf("  %s: %s BTC (%d orders)\n",
				level.Price.String(),
				level.Volume.String(),
				level.OrderCount,
			)
			count++
		}
		return count < 5
	})

	fmt.Println("Bids (Buy Orders):")
	count = 0
	bk.BidTree.Ascend(func(level *book.PriceLevel) bool {
		if level.Volume.GreaterThan(decimal.Zero) && count < 5 {
			fmt.Printf("  %s: %s BTC (%d orders)\n",
				level.Price.String(),
				level.Volume.String(),
				level.OrderCount,
			)
			count++
		}
		return count < 5
	})
}
