package main

import (
	"fmt"
	"time"

	"github.com/adimiuprix/spot-engine/engine"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/matcher"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/shopspring/decimal"
)

func runAmendExample() {
	fmt.Println("=== Amend Orders Example ===\n")

	// Create engine components
	config := engine.Config{
		Symbol:         "BTC/USDT",
		RingBufferSize: 10000,
	}

	eng := engine.New(config)
	book := eng.GetOrderBook()

	publisher := event.NewChannelPublisher(10000)
	seqGen := event.NewSequenceGenerator(0)
	m := matcher.New(book, seqGen, publisher)

	// Start event listener
	go func() {
		for log := range publisher.Channel() {
			switch log.LogType {
			case event.LogTypeTrade:
				fmt.Printf("[TRADE] #%d @ %s (Qty: %s)\n",
					log.TradeID,
					log.TradePrice.String(),
					log.TradeQuantity.String(),
				)
			case event.LogTypeFill:
				fmt.Printf("[FILL] Order %s: Filled=%s, Remaining=%s\n",
					log.OrderID,
					log.FilledSize.String(),
					log.RemainingSize.String(),
				)
			case event.LogTypeCancel:
				fmt.Printf("[CANCEL] Order %s: Cancelled %s @ %s\n",
					log.OrderID,
					log.RemainingSize.String(),
					log.Price.String(),
				)
			case event.LogTypeReject:
				fmt.Printf("[REJECT] %s: %s\n",
					log.RejectReason,
					log.RejectDetail,
				)
			}
		}
	}()

	now := time.Now().UnixNano()

	// Test 1: Size decrease (keeps priority)
	fmt.Println("Test 1: Size Decrease (Keeps Priority)")
	fmt.Println("---------------------------------------")

	order1 := &order.Order{
		ID:        1,
		OrderID:   "order-1",
		CommandID: "cmd-place-1",
		UserID:    100,
		Symbol:    "BTC/USDT",
		Side:      order.Sell,
		Type:      order.Limit,
		Price:     decimal.NewFromInt(50000),
		Quantity:  decimal.NewFromFloat(1.0),
		Filled:    decimal.Zero,
		Timestamp: now,
	}

	book.Add(order1)
	fmt.Printf("✓ Placed sell order: 1.0 BTC @ 50000\n")

	// Amend: decrease size from 1.0 to 0.5
	amendReq1 := &protocol.AmendOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-amend-1",
			UserID:    100,
			MarketID:  "BTC/USDT",
			Timestamp: now + 1000,
		},
		OrderID:  "order-1",
		NewPrice: decimal.NewFromInt(50000), // Same price
		NewSize:  decimal.NewFromFloat(0.5), // Decrease
	}

	result1 := m.ProcessAmend(amendReq1)
	if result1.Success {
		fmt.Printf("✓ Amended order to 0.5 BTC (kept priority)\n")
	} else {
		fmt.Printf("✗ Amend failed: %s\n", result1.Detail)
	}

	time.Sleep(100 * time.Millisecond)
	fmt.Println()

	// Test 2: Size increase (loses priority)
	fmt.Println("Test 2: Size Increase (Loses Priority)")
	fmt.Println("---------------------------------------")

	order2 := &order.Order{
		ID:        2,
		OrderID:   "order-2",
		CommandID: "cmd-place-2",
		UserID:    101,
		Symbol:    "BTC/USDT",
		Side:      order.Buy,
		Type:      order.Limit,
		Price:     decimal.NewFromInt(49000),
		Quantity:  decimal.NewFromFloat(0.5),
		Filled:    decimal.Zero,
		Timestamp: now + 2000,
	}

	book.Add(order2)
	fmt.Printf("✓ Placed buy order: 0.5 BTC @ 49000\n")

	// Amend: increase size from 0.5 to 1.0
	amendReq2 := &protocol.AmendOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-amend-2",
			UserID:    101,
			MarketID:  "BTC/USDT",
			Timestamp: now + 3000,
		},
		OrderID:  "order-2",
		NewPrice: decimal.NewFromInt(49000), // Same price
		NewSize:  decimal.NewFromFloat(1.0), // Increase
	}

	result2 := m.ProcessAmend(amendReq2)
	if result2.Success {
		fmt.Printf("✓ Amended order to 1.0 BTC (lost priority)\n")
	} else {
		fmt.Printf("✗ Amend failed: %s\n", result2.Detail)
	}

	time.Sleep(100 * time.Millisecond)
	fmt.Println()

	// Test 3: Price change (loses priority and might match)
	fmt.Println("Test 3: Price Change (Loses Priority, Immediate Match)")
	fmt.Println("--------------------------------------------------------")

	order3 := &order.Order{
		ID:        3,
		OrderID:   "order-3",
		CommandID: "cmd-place-3",
		UserID:    102,
		Symbol:    "BTC/USDT",
		Side:      order.Buy,
		Type:      order.Limit,
		Price:     decimal.NewFromInt(48000),
		Quantity:  decimal.NewFromFloat(0.3),
		Filled:    decimal.Zero,
		Timestamp: now + 4000,
	}

	book.Add(order3)
	fmt.Printf("✓ Placed buy order: 0.3 BTC @ 48000\n")

	// Amend price to 50000 (will match with order-1)
	amendReq3 := &protocol.AmendOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-amend-3",
			UserID:    102,
			MarketID:  "BTC/USDT",
			Timestamp: now + 5000,
		},
		OrderID:  "order-3",
		NewPrice: decimal.NewFromInt(50000), // Price up
		NewSize:  decimal.NewFromFloat(0.3), // Same size
	}

	result3 := m.ProcessAmend(amendReq3)
	if result3.Success {
		fmt.Printf("✓ Amended price to 50000 (matched immediately!)\n")
		if len(result3.Trades) > 0 {
			fmt.Printf("  Generated %d trade(s)\n", len(result3.Trades))
		}
	} else {
		fmt.Printf("✗ Amend failed: %s\n", result3.Detail)
	}

	time.Sleep(100 * time.Millisecond)
	fmt.Println()

	// Test 4: Invalid amend (wrong user)
	fmt.Println("Test 4: Invalid Amend (Wrong User)")
	fmt.Println("-----------------------------------")

	amendReq4 := &protocol.AmendOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-amend-4",
			UserID:    999, // Wrong user
			MarketID:  "BTC/USDT",
			Timestamp: now + 6000,
		},
		OrderID:  "order-2",
		NewPrice: decimal.NewFromInt(49000),
		NewSize:  decimal.NewFromFloat(0.8),
	}

	result4 := m.ProcessAmend(amendReq4)
	if !result4.Success {
		fmt.Printf("✓ Correctly rejected: %s\n", result4.Detail)
	}

	time.Sleep(100 * time.Millisecond)

	// Print final order book
	fmt.Println("\n=== Final Order Book ===")
	bids, asks := book.GetDepth(5)

	fmt.Println("Bids:")
	if len(bids) == 0 {
		fmt.Println("  (empty)")
	}
	for _, level := range bids {
		fmt.Printf("  %s: %d orders, volume: %s\n",
			level.Price.String(),
			len(level.Orders),
			level.Volume.String(),
		)
	}

	fmt.Println("\nAsks:")
	if len(asks) == 0 {
		fmt.Println("  (empty)")
	}
	for _, level := range asks {
		fmt.Printf("  %s: %d orders, volume: %s\n",
			level.Price.String(),
			len(level.Orders),
			level.Volume.String(),
		)
	}

	publisher.Close()
	time.Sleep(200 * time.Millisecond)
	fmt.Println("\n=== Amend Example Complete ===")
}
