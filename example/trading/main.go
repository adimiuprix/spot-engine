package main

import (
	"fmt"
	"time"

	"github.com/adimiuprix/spot-engine/engine"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("=== Trading Example ===\n")

	// Create engine
	config := engine.Config{
		Symbol:         "BTC/USDT",
		RingBufferSize: 10000,
	}

	eng := engine.New(config)

	// Start trade listener
	go func() {
		for log := range eng.Events() {
			if log.LogType == "trade" {
				fmt.Printf("✅ Trade #%d: BuyOrder=%s, SellOrder=%s, Price=%s, Qty=%s\n",
					log.TradeID,
					log.MakerOrderID,
					log.TakerOrderID,
					log.TradePrice.String(),
					log.TradeQuantity.String(),
				)
			}
		}
	}()

	// Start engine
	eng.Start()
	defer eng.Stop()

	fmt.Println("📊 Submitting orders...\n")

	// Submit buy orders
	eng.SubmitOrder(&order.Order{
		ID:       1,
		UserID:   100,
		Symbol:   "BTC/USDT",
		Side:     order.Buy,
		Type:     order.Limit,
		TIF:      order.GTC,
		Price:    decimal.NewFromInt(50000),
		Quantity: decimal.NewFromFloat(1.0),
		Filled:   decimal.Zero,
	})

	eng.SubmitOrder(&order.Order{
		ID:       2,
		UserID:   101,
		Symbol:   "BTC/USDT",
		Side:     order.Buy,
		Type:     order.Limit,
		TIF:      order.GTC,
		Price:    decimal.NewFromInt(49900),
		Quantity: decimal.NewFromFloat(2.0),
		Filled:   decimal.Zero,
	})

	// Submit sell orders that will match
	eng.SubmitOrder(&order.Order{
		ID:       3,
		UserID:   102,
		Symbol:   "BTC/USDT",
		Side:     order.Sell,
		Type:     order.Limit,
		TIF:      order.GTC,
		Price:    decimal.NewFromInt(50000),
		Quantity: decimal.NewFromFloat(0.5),
		Filled:   decimal.Zero,
	})

	eng.SubmitOrder(&order.Order{
		ID:       4,
		UserID:   103,
		Symbol:   "BTC/USDT",
		Side:     order.Sell,
		Type:     order.Limit,
		TIF:      order.GTC,
		Price:    decimal.NewFromInt(49950),
		Quantity: decimal.NewFromFloat(1.5),
		Filled:   decimal.Zero,
	})

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Print order book
	book := eng.GetOrderBook()
	fmt.Println("\n📖 Final Order Book:")
	fmt.Printf("Symbol: %s\n\n", book.Symbol)

	fmt.Println("💰 Bids:")
	bids, _ := book.GetDepth(10)
	for _, level := range bids {
		fmt.Printf("  %s: %d orders, %s volume\n", level.Price.String(), len(level.Orders), level.Volume.String())
	}

	fmt.Println("\n💸 Asks:")
	_, asks := book.GetDepth(10)
	for _, level := range asks {
		fmt.Printf("  %s: %d orders, %s volume\n", level.Price.String(), len(level.Orders), level.Volume.String())
	}

	fmt.Println("\n✨ Done!")
}
