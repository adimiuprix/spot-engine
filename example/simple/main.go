package main

import (
	"fmt"
	"time"

	"github.com/adimiuprix/spot-engine/engine"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("=== Simple Spot Engine Example ===\n")

	// Create engine
	config := engine.Config{
		Symbol:         "BTC/USDT",
		RingBufferSize: 10000,
	}

	eng := engine.New(config)

	// Start trade listener in background
	go func() {
		for log := range eng.Events() {
			if log.LogType == "trade" {
				fmt.Printf("✅ Trade #%d: Price=%s, Qty=%s\n",
					log.TradeID,
					log.TradePrice.String(),
					log.TradeQuantity.String(),
				)
			}
		}
	}()

	// Start engine
	eng.Start()
	defer eng.Stop()

	fmt.Println("📊 Engine started. Submitting orders...\n")

	// Submit buy orders
	buyOrder1 := &order.Order{
		ID:       1,
		UserID:   100,
		Symbol:   "BTC/USDT",
		Side:     order.Buy,
		Type:     order.Limit,
		TIF:      order.GTC,
		Price:    decimal.NewFromInt(50000),
		Quantity: decimal.NewFromFloat(0.5),
		Filled:   decimal.Zero,
	}

	buyOrder2 := &order.Order{
		ID:       2,
		UserID:   101,
		Symbol:   "BTC/USDT",
		Side:     order.Buy,
		Type:     order.Limit,
		TIF:      order.GTC,
		Price:    decimal.NewFromInt(49900),
		Quantity: decimal.NewFromFloat(1.0),
		Filled:   decimal.Zero,
	}

	// Submit sell orders that will match
	sellOrder1 := &order.Order{
		ID:       3,
		UserID:   102,
		Symbol:   "BTC/USDT",
		Side:     order.Sell,
		Type:     order.Limit,
		TIF:      order.GTC,
		Price:    decimal.NewFromInt(50000),
		Quantity: decimal.NewFromFloat(0.3),
		Filled:   decimal.Zero,
	}

	sellOrder2 := &order.Order{
		ID:       4,
		UserID:   103,
		Symbol:   "BTC/USDT",
		Side:     order.Sell,
		Type:     order.Limit,
		TIF:      order.GTC,
		Price:    decimal.NewFromInt(49950),
		Quantity: decimal.NewFromFloat(0.8),
		Filled:   decimal.Zero,
	}

	// Submit orders
	fmt.Println("📝 Submitting buy orders...")
	eng.SubmitOrder(buyOrder1)
	eng.SubmitOrder(buyOrder2)

	fmt.Println("📝 Submitting sell orders...")
	eng.SubmitOrder(sellOrder1)
	eng.SubmitOrder(sellOrder2)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Print order book
	book := eng.GetOrderBook()
	fmt.Println("\n📖 Order Book Status:")
	fmt.Printf("Symbol: %s\n", book.Symbol)

	fmt.Println("\n💰 Bids (Buy Orders):")
	bids, _ := book.GetDepth(10)
	if len(bids) == 0 {
		fmt.Println("  (empty)")
	}
	for _, level := range bids {
		fmt.Printf("  Price %s: %d orders, Volume: %s\n",
			level.Price.String(),
			len(level.Orders),
			level.Volume.String(),
		)
	}

	fmt.Println("\n💸 Asks (Sell Orders):")
	_, asks := book.GetDepth(10)
	if len(asks) == 0 {
		fmt.Println("  (empty)")
	}
	for _, level := range asks {
		fmt.Printf("  Price %s: %d orders, Volume: %s\n",
			level.Price.String(),
			len(level.Orders),
			level.Volume.String(),
		)
	}

	fmt.Println("\n✨ Done!")
	time.Sleep(100 * time.Millisecond)
}
