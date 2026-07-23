package main

import (
	"context"
	"fmt"
	"time"

	"github.com/adimiuprix/spot-engine/engine"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/shopspring/decimal"
)

// SimpleExample adalah contoh paling sederhana untuk place order dengan Async API
func SimpleExample() {
	fmt.Println("=== Simple Async Place Order ===\n")

	// 1. Setup: Buat publisher dan engine
	publisher := event.NewChannelPublisher(10000)
	eng := engine.NewMatchingEngine(publisher)
	defer eng.Shutdown()

	ctx := context.Background()

	// 2. Buat market dulu
	fmt.Println("Step 1: Creating market...")
	createReq := &protocol.CreateMarketRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-1",
			UserID:    1,
			MarketID:  "BTC/USDT",
			Timestamp: time.Now().UnixNano(),
		},
		MinLotSize: decimal.NewFromFloat(0.00000001),
	}

	createFuture, err := eng.CreateMarket(ctx, createReq)
	if err != nil {
		fmt.Printf("❌ Validation error: %v\n", err)
		return
	}

	_, err = createFuture.Wait(ctx)
	if err != nil {
		fmt.Printf("❌ Execution error: %v\n", err)
		return
	}
	fmt.Println("✅ Market created!\n")

	// 3. Place order dengan Async API
	fmt.Println("Step 2: Placing order...")
	placeReq := &protocol.PlaceOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-2",           // Unique command ID
			UserID:    100,                // User yang place order
			MarketID:  "BTC/USDT",         // Market symbol
			Timestamp: time.Now().UnixNano(), // Timestamp untuk determinism
		},
		OrderID:   "my-order-123",         // Order ID (unique per user)
		Side:      "buy",                  // "buy" atau "sell"
		OrderType: "limit",                // "limit" atau "market"
		Price:     decimal.NewFromInt(50000), // Harga (untuk limit order)
		Size:      decimal.NewFromFloat(1.0), // Quantity
	}

	// Call PlaceOrderAsync - returns Future
	future, err := eng.PlaceOrderAsync(ctx, placeReq)
	if err != nil {
		// Ini error validation (request tidak valid)
		fmt.Printf("❌ Validation error: %v\n", err)
		return
	}

	// Wait for result - bisa timeout atau cancel via context
	result, err := future.Wait(ctx)
	if err != nil {
		// Ini error execution (e.g., market not found)
		fmt.Printf("❌ Execution error: %v\n", err)
		return
	}

	// 4. Print detailed result
	fmt.Println("✅ Order placed successfully!\n")
	fmt.Println("Result:")
	fmt.Printf("  OrderID: %s\n", result.OrderID)
	fmt.Printf("  Accepted: %v\n", result.Accepted)
	fmt.Printf("  Filled: %s BTC\n", result.Filled.String())
	fmt.Printf("  Remaining: %s BTC\n", result.Remaining.String())
	fmt.Printf("  In Order Book: %v\n", result.InBook)
	fmt.Printf("  Trades: %d\n", len(result.Trades))

	if result.InBook {
		fmt.Println("\n📖 Order is now in the order book, waiting for a match!")
	}

	if result.PartialFill {
		fmt.Println("\n⚠️  Order was partially filled!")
	}
}
