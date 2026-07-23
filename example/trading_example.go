package main

import (
	"fmt"
	"time"

	"github.com/adimiuprix/spot-engine/engine"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

func tradingExample() {
	// Create engine
	config := engine.Config{
		Symbol:         "BTC/USDT",
		RingBufferSize: 10000,
	}

	eng := engine.New(config)

	// Start event listener
	go func() {
		for log := range eng.Events() {
			switch log.LogType {
			case event.LogTypeTrade:
				fmt.Printf("[TRADE] #%d: %s @ %s (Qty: %s) | Maker: %s, Taker: %s\n",
					log.TradeID,
					log.MarketID,
					log.TradePrice.String(),
					log.TradeQuantity.String(),
					log.MakerOrderID,
					log.TakerOrderID,
				)
			case event.LogTypeFill:
				fmt.Printf("[FILL] Order %s: Filled %s, Remaining %s\n",
					log.OrderID,
					log.FilledSize.String(),
					log.RemainingSize.String(),
				)
			case event.LogTypeCancel:
				fmt.Printf("[CANCEL] Order %s: Cancelled %s\n",
					log.OrderID,
					log.RemainingSize.String(),
				)
			case event.LogTypeReject:
				fmt.Printf("[REJECT] Order %s: %s - %s\n",
					log.OrderID,
					log.RejectReason,
					log.RejectDetail,
				)
			}
		}
	}()

	// Start engine
	eng.Start()
	defer eng.Stop()

	// Get current timestamp for deterministic behavior
	now := time.Now().UnixNano()

	// Submit buy orders with proper CommandID and Timestamp
	buyOrder1 := &order.Order{
		ID:        1,
		OrderID:   "order-buy-1",
		CommandID: "cmd-buy-1",
		UserID:    100,
		Symbol:    "BTC/USDT",
		Side:      order.Buy,
		Type:      order.Limit,
		TIF:       order.GTC,
		Price:     decimal.NewFromInt(50000),
		Quantity:  decimal.NewFromFloat(0.5),
		Filled:    decimal.Zero,
		Timestamp: now,
	}

	buyOrder2 := &order.Order{
		ID:        2,
		OrderID:   "order-buy-2",
		CommandID: "cmd-buy-2",
		UserID:    101,
		Symbol:    "BTC/USDT",
		Side:      order.Buy,
		Type:      order.Limit,
		TIF:       order.GTC,
		Price:     decimal.NewFromInt(49900),
		Quantity:  decimal.NewFromFloat(1.0),
		Filled:    decimal.Zero,
		Timestamp: now + 1,
	}

	// Submit sell orders with proper CommandID and Timestamp
	sellOrder1 := &order.Order{
		ID:        3,
		OrderID:   "order-sell-1",
		CommandID: "cmd-sell-1",
		UserID:    102,
		Symbol:    "BTC/USDT",
		Side:      order.Sell,
		Type:      order.Limit,
		TIF:       order.GTC,
		Price:     decimal.NewFromInt(50000),
		Quantity:  decimal.NewFromFloat(0.3),
		Filled:    decimal.Zero,
		Timestamp: now + 2,
	}

	sellOrder2 := &order.Order{
		ID:        4,
		OrderID:   "order-sell-2",
		CommandID: "cmd-sell-2",
		UserID:    103,
		Symbol:    "BTC/USDT",
		Side:      order.Sell,
		Type:      order.Limit,
		TIF:       order.GTC,
		Price:     decimal.NewFromInt(49950),
		Quantity:  decimal.NewFromFloat(0.8),
		Filled:    decimal.Zero,
		Timestamp: now + 3,
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
	bids, _ := book.GetDepth(10)
	for _, level := range bids {
		fmt.Printf("  %s: %d orders, volume: %s\n",
			level.Price.String(),
			len(level.Orders),
			level.Volume.String(),
		)
	}

	fmt.Println("\nAsks:")
	_, asks := book.GetDepth(10)
	for _, level := range asks {
		fmt.Printf("  %s: %d orders, volume: %s\n",
			level.Price.String(),
			len(level.Orders),
			level.Volume.String(),
		)
	}

	// Keep running
	time.Sleep(1 * time.Second)
}
