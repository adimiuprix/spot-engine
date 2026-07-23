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

func main() {
	fmt.Println("=== Async Trading API Example ===\n")

	// Create publisher
	publisher := event.NewChannelPublisher(10000)

	// Create matching engine
	eng := engine.NewMatchingEngine(publisher)
	defer eng.Shutdown()

	// Monitor events
	go func() {
		for log := range publisher.Channel() {
			switch log.LogType {
			case "trade":
				fmt.Printf("  ✅ Trade #%d: Price=%s, Qty=%s\n",
					log.TradeID,
					log.TradePrice.String(),
					log.TradeQuantity.String(),
				)
			case "reject":
				fmt.Printf("  ❌ Rejected: %s - %s\n",
					log.RejectReason,
					log.RejectDetail,
				)
			}
		}
	}()

	ctx := context.Background()

	// Step 1: Create market
	fmt.Println("📊 Step 1: Creating market...")
	createReq := &protocol.CreateMarketRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-create-1",
			UserID:    1,
			MarketID:  "BTC/USDT",
			Timestamp: time.Now().UnixNano(),
		},
		MinLotSize: decimal.NewFromFloat(0.00000001),
	}

	createFuture, err := eng.CreateMarket(ctx, createReq)
	if err != nil {
		fmt.Printf("❌ Create market validation failed: %v\n", err)
		return
	}

	success, err := createFuture.Wait(ctx)
	if err != nil {
		fmt.Printf("❌ Create market failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Market created: %v\n\n", success)

	// Step 2: Place buy order (async)
	fmt.Println("📝 Step 2: Placing buy order...")
	placeReq1 := &protocol.PlaceOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-place-1",
			UserID:    100,
			MarketID:  "BTC/USDT",
			Timestamp: time.Now().UnixNano(),
		},
		OrderID:   "order-buy-1",
		Side:      "buy",
		OrderType: "limit",
		Price:     decimal.NewFromInt(50000),
		Size:      decimal.NewFromFloat(1.0),
	}

	placeFuture1, err := eng.PlaceOrderAsync(ctx, placeReq1)
	if err != nil {
		fmt.Printf("❌ Place order validation failed: %v\n", err)
		return
	}

	placeResult1, err := placeFuture1.Wait(ctx)
	if err != nil {
		fmt.Printf("❌ Place order failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Buy order placed: OrderID=%s, InBook=%v, Filled=%s\n\n",
		placeResult1.OrderID,
		placeResult1.InBook,
		placeResult1.Filled.String(),
	)

	// Step 3: Place sell order that matches (async)
	fmt.Println("📝 Step 3: Placing sell order (will match)...")
	placeReq2 := &protocol.PlaceOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-place-2",
			UserID:    101,
			MarketID:  "BTC/USDT",
			Timestamp: time.Now().UnixNano(),
		},
		OrderID:   "order-sell-1",
		Side:      "sell",
		OrderType: "limit",
		Price:     decimal.NewFromInt(50000),
		Size:      decimal.NewFromFloat(0.5),
	}

	placeFuture2, err := eng.PlaceOrderAsync(ctx, placeReq2)
	if err != nil {
		fmt.Printf("❌ Place order validation failed: %v\n", err)
		return
	}

	placeResult2, err := placeFuture2.Wait(ctx)
	if err != nil {
		fmt.Printf("❌ Place order failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Sell order placed: OrderID=%s, Filled=%s, Trades=%d\n\n",
		placeResult2.OrderID,
		placeResult2.Filled.String(),
		len(placeResult2.Trades),
	)

	time.Sleep(100 * time.Millisecond) // Wait for events

	// Step 4: Amend remaining buy order (async)
	fmt.Println("📝 Step 4: Amending buy order...")
	amendReq := &protocol.AmendOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-amend-1",
			UserID:    100,
			MarketID:  "BTC/USDT",
			Timestamp: time.Now().UnixNano(),
		},
		OrderID:  "order-buy-1",
		NewPrice: decimal.NewFromInt(49900),
		NewSize:  decimal.NewFromFloat(0.8),
	}

	amendFuture, err := eng.AmendOrderAsync(ctx, amendReq)
	if err != nil {
		fmt.Printf("❌ Amend order validation failed: %v\n", err)
		return
	}

	amendResult, err := amendFuture.Wait(ctx)
	if err != nil {
		fmt.Printf("❌ Amend order failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Order amended: OrderID=%s, NewPrice=%s, NewSize=%s\n\n",
		amendResult.OrderID,
		amendResult.NewPrice.String(),
		amendResult.NewSize.String(),
	)

	// Step 5: Cancel order (async)
	fmt.Println("📝 Step 5: Cancelling order...")
	cancelReq := &protocol.CancelOrderRequest{
		CommandID: "cmd-cancel-1",
		UserID:    100,
		Symbol:    "BTC/USDT",
		OrderID:   "order-buy-1",
		Timestamp: time.Now().UnixNano(),
	}

	cancelFuture, err := eng.CancelOrderAsync(ctx, cancelReq)
	if err != nil {
		fmt.Printf("❌ Cancel order validation failed: %v\n", err)
		return
	}

	cancelResult, err := cancelFuture.Wait(ctx)
	if err != nil {
		fmt.Printf("❌ Cancel order failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Order cancelled: OrderID=%s, CancelledSize=%s\n\n",
		cancelResult.OrderID,
		cancelResult.CancelledSize.String(),
	)

	// Step 6: Get market stats
	fmt.Println("📊 Step 6: Market statistics...")
	stats, err := eng.GetStats("BTC/USDT")
	if err != nil {
		fmt.Printf("❌ Get stats failed: %v\n", err)
		return
	}

	fmt.Printf("Market: %s\n", stats.MarketID)
	fmt.Printf("State: %s\n", stats.State)
	fmt.Printf("Bid Count: %d\n", stats.BidCount)
	fmt.Printf("Ask Count: %d\n", stats.AskCount)
	fmt.Printf("Best Bid: %s\n", stats.BestBid.String())
	fmt.Printf("Best Ask: %s\n", stats.BestAsk.String())

	fmt.Println("\n✨ Done! All async operations completed successfully.")
	time.Sleep(100 * time.Millisecond)
}
