package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/adimiuprix/spot-engine/engine"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/adimiuprix/spot-engine/snapshot"
	"github.com/shopspring/decimal"
)

func runSnapshotExample() {
	fmt.Println("=== Snapshot & Restore Example ===\n")

	// Create snapshot directory
	snapshotDir := "./test-snapshot"
	defer os.RemoveAll(snapshotDir) // Clean up after test

	// Phase 1: Create engine and place orders
	fmt.Println("Phase 1: Creating engine and placing orders")
	fmt.Println("--------------------------------------------")

	publisher := event.NewChannelPublisher(10000)
	eng := engine.NewMatchingEngine(publisher)

	ctx := context.Background()
	now := time.Now().UnixNano()

	// Create market
	createReq := &protocol.CreateMarketRequest{
		BaseCommand: protocol.BaseCommand{
			Type:      protocol.CmdCreateMarket,
			CommandID: "cmd-create-1",
			UserID:    1000,
			MarketID:  "BTC-USDT",
			Timestamp: now,
		},
		MinLotSize: decimal.NewFromFloat(0.001),
	}

	future, _ := eng.CreateMarket(ctx, createReq)
	future.Wait(ctx)
	fmt.Println("✓ Created BTC-USDT market")

	// Get market and place orders directly
	market, _ := eng.GetMarket("BTC-USDT")

	// Place some orders
	orders := []*order.Order{
		{
			ID:        1,
			OrderID:   "order-1",
			CommandID: "cmd-1",
			UserID:    100,
			Symbol:    "BTC-USDT",
			Side:      order.Buy,
			Type:      order.Limit,
			Price:     decimal.NewFromInt(50000),
			Quantity:  decimal.NewFromFloat(1.0),
			Filled:    decimal.Zero,
			Timestamp: now + 1,
		},
		{
			ID:        2,
			OrderID:   "order-2",
			CommandID: "cmd-2",
			UserID:    101,
			Symbol:    "BTC-USDT",
			Side:      order.Buy,
			Type:      order.Limit,
			Price:     decimal.NewFromInt(49000),
			Quantity:  decimal.NewFromFloat(0.5),
			Filled:    decimal.Zero,
			Timestamp: now + 2,
		},
		{
			ID:        3,
			OrderID:   "order-3",
			CommandID: "cmd-3",
			UserID:    102,
			Symbol:    "BTC-USDT",
			Side:      order.Sell,
			Type:      order.Limit,
			Price:     decimal.NewFromInt(51000),
			Quantity:  decimal.NewFromFloat(0.8),
			Filled:    decimal.Zero,
			Timestamp: now + 3,
		},
	}

	for _, o := range orders {
		market.OrderBook.Add(o)
	}

	fmt.Printf("✓ Placed %d orders\n", len(orders))

	// Print initial state
	fmt.Println("\nInitial Order Book State:")
	bids, asks := market.OrderBook.GetDepth(10)
	fmt.Printf("  Bids: %d levels\n", len(bids))
	fmt.Printf("  Asks: %d levels\n", len(asks))
	for _, level := range bids {
		fmt.Printf("    %s: %d orders\n", level.Price.String(), len(level.Orders))
	}
	for _, level := range asks {
		fmt.Printf("    %s: %d orders\n", level.Price.String(), len(level.Orders))
	}

	// Phase 2: Take snapshot
	fmt.Println("\nPhase 2: Taking snapshot")
	fmt.Println("------------------------")

	snapshots, globalSeq := eng.TakeSnapshot()
	fmt.Printf("✓ Captured %d market snapshot(s)\n", len(snapshots))
	fmt.Printf("  Global sequence ID: %d\n", globalSeq)

	// Write to disk
	writer := snapshot.NewWriter(snapshotDir)
	if err := writer.WriteSnapshot(snapshots, globalSeq); err != nil {
		fmt.Printf("✗ Failed to write snapshot: %v\n", err)
		return
	}
	fmt.Printf("✓ Snapshot written to %s\n", snapshotDir)

	// Phase 3: Simulate engine restart
	fmt.Println("\nPhase 3: Simulating engine restart")
	fmt.Println("-----------------------------------")

	// Close old engine
	eng.Shutdown()
	fmt.Println("✓ Old engine shut down")

	// Create new engine
	publisher2 := event.NewChannelPublisher(10000)
	eng2 := engine.NewMatchingEngine(publisher2)
	fmt.Println("✓ New engine created (empty)")

	// Phase 4: Restore from snapshot
	fmt.Println("\nPhase 4: Restoring from snapshot")
	fmt.Println("---------------------------------")

	reader := snapshot.NewReader(snapshotDir)
	metadata, restoredSnapshots, err := reader.ReadSnapshot()
	if err != nil {
		fmt.Printf("✗ Failed to read snapshot: %v\n", err)
		return
	}

	fmt.Printf("✓ Read snapshot (schema v%d)\n", metadata.SchemaVersion)
	fmt.Printf("  Engine version: %s\n", metadata.EngineVersion)
	fmt.Printf("  Markets: %d\n", len(restoredSnapshots))
	fmt.Printf("  Checksum: %d (verified)\n", metadata.SnapshotChecksum)

	// Restore into engine
	if err := eng2.RestoreFromSnapshot(restoredSnapshots); err != nil {
		fmt.Printf("✗ Failed to restore: %v\n", err)
		return
	}
	fmt.Println("✓ State restored successfully")

	// Phase 5: Verify restored state
	fmt.Println("\nPhase 5: Verifying restored state")
	fmt.Println("----------------------------------")

	market2, err := eng2.GetMarket("BTC-USDT")
	if err != nil {
		fmt.Printf("✗ Market not found after restore: %v\n", err)
		return
	}

	bids2, asks2 := market2.OrderBook.GetDepth(10)
	fmt.Printf("✓ Market found: BTC-USDT\n")
	fmt.Printf("  Bids: %d levels\n", len(bids2))
	fmt.Printf("  Asks: %d levels\n", len(asks2))

	// Verify order count
	totalOrders := 0
	for _, level := range bids2 {
		totalOrders += len(level.Orders)
		fmt.Printf("    %s: %d orders\n", level.Price.String(), len(level.Orders))
	}
	for _, level := range asks2 {
		totalOrders += len(level.Orders)
		fmt.Printf("    %s: %d orders\n", level.Price.String(), len(level.Orders))
	}

	if totalOrders == len(orders) {
		fmt.Printf("✓ All %d orders restored correctly\n", totalOrders)
	} else {
		fmt.Printf("✗ Order count mismatch: expected %d, got %d\n", len(orders), totalOrders)
	}

	// Verify state
	stats, _ := eng2.GetStats("BTC-USDT")
	fmt.Printf("  Market state: %s\n", stats.State)
	fmt.Printf("  MinLotSize: %s\n", stats.MinLotSize.String())

	publisher2.Close()
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n=== Snapshot Example Complete ===")
	fmt.Println("Snapshot file can be used for:")
	fmt.Println("  • Recovery after crash")
	fmt.Println("  • Deterministic replay from checkpoint")
	fmt.Println("  • State inspection and audit")
}
