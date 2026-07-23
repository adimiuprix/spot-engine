package main

import (
	"fmt"
	"time"

	"github.com/adimiuprix/spot-engine/engine"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

func runIcebergExample() {
	fmt.Println("=== Iceberg Orders Example ===\n")

	// Create engine
	config := engine.Config{
		Symbol:         "BTC/USDT",
		RingBufferSize: 10000,
	}

	eng := engine.New(config)

	// Start event listener
	eventCount := 0
	go func() {
		for log := range eng.Events() {
			eventCount++
			switch log.LogType {
			case event.LogTypeTrade:
				fmt.Printf("[TRADE #%d] %s @ %s (Qty: %s)\n",
					log.TradeID,
					log.MarketID,
					log.TradePrice.String(),
					log.TradeQuantity.String(),
				)
			case event.LogTypeFill:
				status := "PARTIAL"
				if log.RemainingSize.IsZero() {
					status = "FULL"
				}
				fmt.Printf("[FILL] Order %s: %s FILL - Filled: %s, Remaining: %s\n",
					log.OrderID,
					status,
					log.FilledSize.String(),
					log.RemainingSize.String(),
				)
			}
		}
	}()

	// Start engine
	eng.Start()
	defer eng.Stop()

	now := time.Now().UnixNano()

	fmt.Println("1. Placing ICEBERG SELL order:")
	fmt.Println("   Total: 2.0 BTC, Visible: 0.5 BTC per display")
	fmt.Println("   Hidden: 1.5 BTC (will replenish 3 times)\n")

	// Create large iceberg sell order
	// Total: 2.0 BTC, Visible: 0.5 BTC at a time
	icebergSell := &order.Order{
		ID:        100,
		OrderID:   "iceberg-sell-1",
		CommandID: "cmd-iceberg-1",
		UserID:    200,
		Symbol:    "BTC/USDT",
		Side:      order.Sell,
		Type:      order.Limit,
		TIF:       order.GTC,
		Price:     decimal.NewFromInt(50000),
		Quantity:  decimal.NewFromFloat(2.0),
		Filled:    decimal.Zero,
		Timestamp: now,
	}

	// Setup iceberg with 0.5 BTC visible
	err := icebergSell.SetupIceberg(decimal.NewFromFloat(0.5))
	if err != nil {
		fmt.Printf("Error setting up iceberg: %v\n", err)
		return
	}

	fmt.Printf("   Initial state: Visible=%s, Hidden=%s\n\n",
		icebergSell.VisibleQuantity().String(),
		icebergSell.HiddenSize.String(),
	)

	eng.SubmitOrder(icebergSell)
	time.Sleep(50 * time.Millisecond)

	fmt.Println("2. Sending 4 small BUY orders (0.6 BTC each):")
	fmt.Println("   This will trigger 3 replenishments as visible portions deplete\n")

	// Submit 4 buy orders that will gradually consume the iceberg
	for i := 0; i < 4; i++ {
		buyOrder := &order.Order{
			ID:        uint64(i + 1),
			OrderID:   fmt.Sprintf("buy-%d", i+1),
			CommandID: fmt.Sprintf("cmd-buy-%d", i+1),
			UserID:    uint64(100 + i),
			Symbol:    "BTC/USDT",
			Side:      order.Buy,
			Type:      order.Limit,
			TIF:       order.GTC,
			Price:     decimal.NewFromInt(50000),
			Quantity:  decimal.NewFromFloat(0.6),
			Filled:    decimal.Zero,
			Timestamp: now + int64(i+1)*1000,
		}

		fmt.Printf("   Sending Buy Order #%d (0.6 BTC)...\n", i+1)
		eng.SubmitOrder(buyOrder)
		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(200 * time.Millisecond)

	// Print final order book state
	book := eng.GetOrderBook()
	fmt.Println("\n=== Final Order Book ===")
	fmt.Printf("Symbol: %s\n", book.Symbol)

	fmt.Println("\nBids:")
	bids, _ := book.GetDepth(5)
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
	_, asks := book.GetDepth(5)
	if len(asks) == 0 {
		fmt.Println("  (empty)")
	}
	for _, level := range asks {
		totalVisible := decimal.Zero
		totalHidden := decimal.Zero
		for _, o := range level.Orders {
			totalVisible = totalVisible.Add(o.VisibleQuantity())
			totalHidden = totalHidden.Add(o.HiddenSize)
		}
		fmt.Printf("  %s: %d orders, visible: %s, hidden: %s\n",
			level.Price.String(),
			len(level.Orders),
			totalVisible.String(),
			totalHidden.String(),
		)
	}

	fmt.Printf("\nTotal events emitted: %d\n", eventCount)
	fmt.Println("\n=== Iceberg Example Complete ===")
	time.Sleep(500 * time.Millisecond)
}
