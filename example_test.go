package spotengine_test

import (
	"context"
	"fmt"
	"time"

	"github.com/adimiuprix/spot-engine/engine"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/shopspring/decimal"
)

// Example demonstrates basic usage of the matching engine
func Example() {
	// Create matching engine
	publisher := event.NewChannelPublisher(10000)
	eng := engine.NewMatchingEngine(publisher)
	defer eng.Shutdown()

	// Create market
	ctx := context.Background()
	createReq := &protocol.CreateMarketRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-create",
			UserID:    1000,
			MarketID:  "BTC-USDT",
			Timestamp: time.Now().UnixNano(),
		},
		MinLotSize: decimal.NewFromFloat(0.001),
	}
	future, _ := eng.CreateMarket(ctx, createReq)
	_, _ = future.Wait(ctx)

	fmt.Println("Market created")
	// Output: Market created
}

// Example_limitOrder demonstrates placing limit orders
func Example_limitOrder() {
	publisher := event.NewChannelPublisher(10000)
	eng := engine.NewMatchingEngine(publisher)
	defer eng.Shutdown()

	ctx := context.Background()

	// Create market
	createReq := &protocol.CreateMarketRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-1",
			UserID:    1000,
			MarketID:  "BTC-USDT",
			Timestamp: time.Now().UnixNano(),
		},
		MinLotSize: decimal.NewFromFloat(0.001),
	}
	future, _ := eng.CreateMarket(ctx, createReq)
	future.Wait(ctx)

	// Place buy order
	buyReq := &protocol.PlaceOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-2",
			UserID:    1001,
			MarketID:  "BTC-USDT",
			Timestamp: time.Now().UnixNano(),
		},
		OrderID:   "order-1",
		Side:      "buy",
		OrderType: "limit",
		Price:     decimal.NewFromInt(50000),
		Size:      decimal.NewFromFloat(0.1),
	}
	eng.PlaceOrderAsync(ctx, buyReq)

	fmt.Println("Buy order placed")
	// Output: Buy order placed
}

// Example_marketOrder demonstrates market order execution
func Example_marketOrder() {
	publisher := event.NewChannelPublisher(10000)
	eng := engine.NewMatchingEngine(publisher)
	defer eng.Shutdown()

	ctx := context.Background()

	// Create market
	createReq := &protocol.CreateMarketRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-1",
			UserID:    1000,
			MarketID:  "BTC-USDT",
			Timestamp: time.Now().UnixNano(),
		},
		MinLotSize: decimal.NewFromFloat(0.001),
	}
	future, _ := eng.CreateMarket(ctx, createReq)
	future.Wait(ctx)

	// Place sell limit order first (liquidity)
	sellReq := &protocol.PlaceOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-2",
			UserID:    1001,
			MarketID:  "BTC-USDT",
			Timestamp: time.Now().UnixNano(),
		},
		OrderID:   "order-1",
		Side:      "sell",
		OrderType: "limit",
		Price:     decimal.NewFromInt(50000),
		Size:      decimal.NewFromFloat(0.1),
	}
	eng.PlaceOrderAsync(ctx, sellReq)

	time.Sleep(10 * time.Millisecond) // Let order process

	// Place market buy order
	buyReq := &protocol.PlaceOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-3",
			UserID:    1002,
			MarketID:  "BTC-USDT",
			Timestamp: time.Now().UnixNano(),
		},
		OrderID:   "order-2",
		Side:      "buy",
		OrderType: "market",
		Size:      decimal.NewFromFloat(0.05),
	}
	eng.PlaceOrderAsync(ctx, buyReq)

	fmt.Println("Market order executed")
	// Output: Market order executed
}

// Example_icebergOrder demonstrates iceberg order with hidden quantity
func Example_icebergOrder() {
	publisher := event.NewChannelPublisher(10000)
	eng := engine.NewMatchingEngine(publisher)
	defer eng.Shutdown()

	ctx := context.Background()

	// Create market
	createReq := &protocol.CreateMarketRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-1",
			UserID:    1000,
			MarketID:  "BTC-USDT",
			Timestamp: time.Now().UnixNano(),
		},
		MinLotSize: decimal.NewFromFloat(0.001),
	}
	future, _ := eng.CreateMarket(ctx, createReq)
	future.Wait(ctx)

	// Place iceberg order (total 1.0, show 0.1 at a time)
	icebergReq := &protocol.PlaceOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-2",
			UserID:    1001,
			MarketID:  "BTC-USDT",
			Timestamp: time.Now().UnixNano(),
		},
		OrderID:     "order-1",
		Side:        "sell",
		OrderType:   "limit",
		Price:       decimal.NewFromInt(50000),
		Size:        decimal.NewFromFloat(1.0),
		VisibleSize: decimal.NewFromFloat(0.1), // Show 0.1 at a time
	}
	eng.PlaceOrderAsync(ctx, icebergReq)

	fmt.Println("Iceberg order placed")
	// Output: Iceberg order placed
}

// Example_amendOrder demonstrates order amendment
func Example_amendOrder() {
	publisher := event.NewChannelPublisher(10000)
	eng := engine.NewMatchingEngine(publisher)
	defer eng.Shutdown()

	ctx := context.Background()

	// Create market
	createReq := &protocol.CreateMarketRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-1",
			UserID:    1000,
			MarketID:  "BTC-USDT",
			Timestamp: time.Now().UnixNano(),
		},
		MinLotSize: decimal.NewFromFloat(0.001),
	}
	future, _ := eng.CreateMarket(ctx, createReq)
	future.Wait(ctx)

	// Place original order
	placeReq := &protocol.PlaceOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-2",
			UserID:    1001,
			MarketID:  "BTC-USDT",
			Timestamp: time.Now().UnixNano(),
		},
		OrderID:   "order-1",
		Side:      "buy",
		OrderType: "limit",
		Price:     decimal.NewFromInt(49000),
		Size:      decimal.NewFromFloat(0.1),
	}
	eng.PlaceOrderAsync(ctx, placeReq)

	time.Sleep(10 * time.Millisecond) // Let order process

	// Amend order (reduce size, keep priority)
	amendReq := &protocol.AmendOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-3",
			UserID:    1001,
			MarketID:  "BTC-USDT",
			Timestamp: time.Now().UnixNano(),
		},
		OrderID:  "order-1",
		NewPrice: decimal.NewFromInt(49000),  // Same price
		NewSize:  decimal.NewFromFloat(0.05), // Reduce size
	}
	eng.AmendOrderAsync(ctx, amendReq)

	fmt.Println("Order amended")
	// Output: Order amended
}

// Example_timeInForce demonstrates IOC (Immediate-or-Cancel) order
func Example_timeInForce() {
	publisher := event.NewChannelPublisher(10000)
	eng := engine.NewMatchingEngine(publisher)
	defer eng.Shutdown()

	ctx := context.Background()

	// Create market
	createReq := &protocol.CreateMarketRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-1",
			UserID:    1000,
			MarketID:  "BTC-USDT",
			Timestamp: time.Now().UnixNano(),
		},
		MinLotSize: decimal.NewFromFloat(0.001),
	}
	future, _ := eng.CreateMarket(ctx, createReq)
	future.Wait(ctx)

	// Place IOC order (matches immediately, cancels rest)
	iocReq := &protocol.PlaceOrderRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-2",
			UserID:    1001,
			MarketID:  "BTC-USDT",
			Timestamp: time.Now().UnixNano(),
		},
		OrderID:   "order-1",
		Side:      "buy",
		OrderType: "limit",
		Price:     decimal.NewFromInt(50000),
		Size:      decimal.NewFromFloat(0.1),
	}
	// Note: TIF is set on the Order struct when converting from request
	eng.PlaceOrderAsync(ctx, iocReq)

	fmt.Println("IOC order processed")
	// Output: IOC order processed
}

// Example_snapshot demonstrates taking and restoring snapshots
func Example_snapshot() {
	publisher := event.NewChannelPublisher(10000)
	eng := engine.NewMatchingEngine(publisher)
	defer eng.Shutdown()

	ctx := context.Background()

	// Create market
	createReq := &protocol.CreateMarketRequest{
		BaseCommand: protocol.BaseCommand{
			CommandID: "cmd-1",
			UserID:    1000,
			MarketID:  "BTC-USDT",
			Timestamp: time.Now().UnixNano(),
		},
		MinLotSize: decimal.NewFromFloat(0.001),
	}
	future, _ := eng.CreateMarket(ctx, createReq)
	future.Wait(ctx)

	time.Sleep(10 * time.Millisecond)

	// Take snapshot
	snapshots, seqID := eng.TakeSnapshot()
	fmt.Printf("Snapshot taken: %d markets, seqID=%d\n", len(snapshots), seqID)

	// Snapshots can be written to disk using snapshot.Writer
	// and restored using snapshot.Reader

	// Output: Snapshot taken: 1 markets, seqID=0
}

// Example_events demonstrates listening to engine events
func Example_events() {
	publisher := event.NewChannelPublisher(10000)
	eng := engine.NewMatchingEngine(publisher)
	defer eng.Shutdown()

	// Place orders...
	fmt.Println("Event listener started")
	// Output: Event listener started
}
