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
	fmt.Println("=== Async Place Order Example ===\n")

	// Step 1: Create publisher for event monitoring
	publisher := event.NewChannelPublisher(10000)

	// Step 2: Create matching engine
	eng := engine.NewMatchingEngine(publisher)
	defer eng.Shutdown()

	// Step 3: Monitor events in background
	go monitorEvents(publisher)

	// Step 4: Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Step 5: Create market first
	fmt.Println("📊 Creating market BTC/USDT...")
	if err := createMarket(ctx, eng); err != nil {
		fmt.Printf("❌ Failed to create market: %v\n", err)
		return
	}
	fmt.Println("✅ Market created successfully\n")

	// Step 6: Place buy order
	fmt.Println("📝 Placing buy order (1.0 BTC @ $50,000)...")
	buyResult, err := placeBuyOrder(ctx, eng)
	if err != nil {
		fmt.Printf("❌ Failed to place buy order: %v\n", err)
		return
	}
	printOrderResult("BUY", buyResult)

	// Step 7: Place another buy order
	fmt.Println("\n📝 Placing another buy order (2.0 BTC @ $49,900)...")
	buyResult2, err := placeBuyOrder2(ctx, eng)
	if err != nil {
		fmt.Printf("❌ Failed to place buy order: %v\n", err)
		return
	}
	printOrderResult("BUY", buyResult2)

	// Wait a bit for orders to settle
	time.Sleep(100 * time.Millisecond)

	// Step 8: Place sell order that will match
	fmt.Println("\n📝 Placing sell order that will match (0.5 BTC @ $50,000)...")
	sellResult, err := placeSellOrder(ctx, eng)
	if err != nil {
		fmt.Printf("❌ Failed to place sell order: %v\n", err)
		return
	}
	printOrderResult("SELL", sellResult)

	// Wait for events to be processed
	time.Sleep(200 * time.Millisecond)

	// Step 9: Place market buy order
	fmt.Println("\n📝 Placing market buy order (0.3 BTC)...")
	marketBuyResult, err := placeMarketBuyOrder(ctx, eng)
	if err != nil {
		fmt.Printf("❌ Failed to place market order: %v\n", err)
		return
	}
	printOrderResult("MARKET BUY", marketBuyResult)

	// Wait for events
	time.Sleep(200 * time.Millisecond)

	// Step 10: Get market stats
	fmt.Println("\n📊 Final Market Statistics:")
	stats, err := eng.GetStats("BTC/USDT")
	if err != nil {
		fmt.Printf("❌ Failed to get stats: %v\n", err)
		return
	}

	fmt.Printf("  Market ID: %s\n", stats.MarketID)
	fmt.Printf("  State: %s\n", stats.State)
	fmt.Printf("  Bid Levels: %d\n", stats.BidCount)
	fmt.Printf("  Ask Levels: %d\n", stats.AskCount)
	if stats.BestBid.GreaterThan(decimal.Zero) {
		fmt.Printf("  Best Bid: $%s\n", stats.BestBid.String())
	}
	if stats.BestAsk.GreaterThan(decimal.Zero) {
		fmt.Printf("  Best Ask: $%s\n", stats.BestAsk.String())
	}

	fmt.Println("\n✨ Done! All orders placed successfully.")
	time.Sleep(100 * time.Millisecond)
}

// createMarket creates a new trading market
func createMarket(ctx context.Context, eng *engine.MatchingEngine) error {
	req := &protocol.CreateMarketRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: fmt.Sprintf("cmd-create-%d", time.Now().UnixNano()),
			UserID:    1,
			MarketID:  "BTC/USDT",
			Timestamp: time.Now().UnixNano(),
		},
		MinLotSize: decimal.NewFromFloat(0.00000001), // 1 satoshi
	}

	future, err := eng.CreateMarket(ctx, req)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	_, err = future.Wait(ctx)
	if err != nil {
		return fmt.Errorf("execution error: %w", err)
	}

	return nil
}

// placeBuyOrder places a limit buy order
func placeBuyOrder(ctx context.Context, eng *engine.MatchingEngine) (*protocol.PlaceOrderResult, error) {
	req := &protocol.PlaceOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: fmt.Sprintf("cmd-buy-%d", time.Now().UnixNano()),
			UserID:    100,
			MarketID:  "BTC/USDT",
			Timestamp: time.Now().UnixNano(),
		},
		OrderID:   fmt.Sprintf("order-buy-%d", time.Now().UnixNano()),
		Side:      "buy",
		OrderType: "limit",
		Price:     decimal.NewFromInt(50000),
		Size:      decimal.NewFromFloat(1.0),
	}

	future, err := eng.PlaceOrderAsync(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	result, err := future.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("execution error: %w", err)
	}

	return result, nil
}

// placeBuyOrder2 places another buy order at different price
func placeBuyOrder2(ctx context.Context, eng *engine.MatchingEngine) (*protocol.PlaceOrderResult, error) {
	req := &protocol.PlaceOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: fmt.Sprintf("cmd-buy2-%d", time.Now().UnixNano()),
			UserID:    101,
			MarketID:  "BTC/USDT",
			Timestamp: time.Now().UnixNano(),
		},
		OrderID:   fmt.Sprintf("order-buy2-%d", time.Now().UnixNano()),
		Side:      "buy",
		OrderType: "limit",
		Price:     decimal.NewFromInt(49900),
		Size:      decimal.NewFromFloat(2.0),
	}

	future, err := eng.PlaceOrderAsync(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	result, err := future.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("execution error: %w", err)
	}

	return result, nil
}

// placeSellOrder places a limit sell order that will match
func placeSellOrder(ctx context.Context, eng *engine.MatchingEngine) (*protocol.PlaceOrderResult, error) {
	req := &protocol.PlaceOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: fmt.Sprintf("cmd-sell-%d", time.Now().UnixNano()),
			UserID:    200,
			MarketID:  "BTC/USDT",
			Timestamp: time.Now().UnixNano(),
		},
		OrderID:   fmt.Sprintf("order-sell-%d", time.Now().UnixNano()),
		Side:      "sell",
		OrderType: "limit",
		Price:     decimal.NewFromInt(50000),
		Size:      decimal.NewFromFloat(0.5),
	}

	future, err := eng.PlaceOrderAsync(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	result, err := future.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("execution error: %w", err)
	}

	return result, nil
}

// placeMarketBuyOrder places a market buy order
func placeMarketBuyOrder(ctx context.Context, eng *engine.MatchingEngine) (*protocol.PlaceOrderResult, error) {
	req := &protocol.PlaceOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: fmt.Sprintf("cmd-market-%d", time.Now().UnixNano()),
			UserID:    300,
			MarketID:  "BTC/USDT",
			Timestamp: time.Now().UnixNano(),
		},
		OrderID:   fmt.Sprintf("order-market-%d", time.Now().UnixNano()),
		Side:      "buy",
		OrderType: "market",
		Price:     decimal.Zero, // Market order doesn't need price
		Size:      decimal.NewFromFloat(0.3),
	}

	future, err := eng.PlaceOrderAsync(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	result, err := future.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("execution error: %w", err)
	}

	return result, nil
}

// printOrderResult prints detailed order result
func printOrderResult(orderType string, result *protocol.PlaceOrderResult) {
	fmt.Printf("✅ %s Order Result:\n", orderType)
	fmt.Printf("   OrderID: %s\n", result.OrderID)
	fmt.Printf("   Accepted: %v\n", result.Accepted)
	fmt.Printf("   Filled: %s BTC\n", result.Filled.String())
	fmt.Printf("   Remaining: %s BTC\n", result.Remaining.String())
	fmt.Printf("   In Order Book: %v\n", result.InBook)
	fmt.Printf("   Partial Fill: %v\n", result.PartialFill)
	fmt.Printf("   Trades Generated: %d\n", len(result.Trades))

	if result.Filled.GreaterThan(decimal.Zero) {
		fmt.Printf("   💰 Filled Amount: %s BTC\n", result.Filled.String())
	}

	if result.InBook {
		fmt.Printf("   📖 Order is resting in book (waiting for match)\n")
	}

	if result.PartialFill {
		fmt.Printf("   ⚠️  Partially filled (some matched, some in book)\n")
	}
}

// monitorEvents monitors and prints events
func monitorEvents(publisher *event.ChannelPublisher) {
	for log := range publisher.Channel() {
		switch log.LogType {
		case "trade":
			fmt.Printf("      🔄 Trade #%d: %s BTC @ $%s (Maker: %s, Taker: %s)\n",
				log.TradeID,
				log.TradeQuantity.String(),
				log.TradePrice.String(),
				log.MakerOrderID,
				log.TakerOrderID,
			)
		case "fill":
			fmt.Printf("      💰 Fill: Order %s (%s) - Filled: %s BTC\n",
				log.OrderID,
				log.Side,
				log.FilledSize.String(),
			)
		case "reject":
			fmt.Printf("      ❌ Rejected: Order %s - %s (%s)\n",
				log.OrderID,
				log.RejectReason,
				log.RejectDetail,
			)
		}
	}
}
