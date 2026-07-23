package main

import (
	"fmt"
	"time"

	"github.com/adimiuprix/spot-engine/engine"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

func main() {
	// Create engine
	config := engine.Config{
		Symbol:         "BTC/USDT",
		RingBufferSize: 10000,
	}

	eng := engine.New(config)

	// Start trade listener
	go func() {
		for trade := range eng.Trades() {
			fmt.Printf("Trade #%d: %s @ %s (Qty: %s)\n",
				trade.ID,
				trade.Symbol,
				trade.Price.String(),
				trade.Quantity.String(),
			)
		}
	}()

	// Start engine
	eng.Start()
	defer eng.Stop()

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

	// Submit sell orders
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
	fmt.Println("Submitting orders...")
	eng.SubmitOrder(buyOrder1)
	eng.SubmitOrder(buyOrder2)
	eng.SubmitOrder(sellOrder1)
	eng.SubmitOrder(sellOrder2)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Print order book
	book := eng.GetOrderBook()
	fmt.Println("\n=== Order Book ===")
	fmt.Printf("Symbol: %s\n", book.Symbol)
	
	fmt.Println("\nBids:")
	for price, level := range book.Bids {
		fmt.Printf("  %s: %d orders, volume: %s\n", 
			price, 
			len(level.Orders), 
			level.Volume.String(),
		)
	}

	fmt.Println("\nAsks:")
	for price, level := range book.Asks {
		fmt.Printf("  %s: %d orders, volume: %s\n", 
			price, 
			len(level.Orders), 
			level.Volume.String(),
		)
	}

	// Keep running
	time.Sleep(1 * time.Second)
}
