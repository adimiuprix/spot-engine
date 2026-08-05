/*
Package spot-engine provides a high-performance, deterministic matching engine for spot trading.

# Overview

The spot-engine is a production-ready matching engine designed for cryptocurrency and
financial exchanges. It implements a complete order book with advanced features including
iceberg orders, order amendments, time-in-force policies, and state management.

# Key Features

  - High Performance: ~23ns best price lookup with zero allocations
  - Deterministic Replay: All events use upstream-assigned timestamps
  - Time-in-Force (TIF): GTC, IOC, FOK, and PostOnly support
  - Iceberg Orders: Hide order quantity with automatic replenishment
  - Order Amendments: Modify orders with proper priority rules
  - Market Management: Create, suspend, resume markets with state enforcement
  - Snapshot & Restore: Point-in-time recovery with CRC32 validation
  - Event Logging: Comprehensive audit trail for all operations

# Architecture

The engine follows a single-threaded event-driven architecture for determinism:

	┌─────────────────────────────────────────────┐
	│           MatchingEngine                    │
	│  ┌─────────┐  ┌─────────┐  ┌─────────┐     │
	│  │ Market  │  │ Market  │  │ Market  │     │
	│  │ BTC-USD │  │ ETH-USD │  │ SOL-USD │     │
	│  └────┬────┘  └────┬────┘  └────┬────┘     │
	│       │            │            │          │
	│  ┌────▼────────────▼────────────▼────┐     │
	│  │      OrderBook (B-Tree)           │     │
	│  │  ┌──────┐      ┌──────┐           │     │
	│  │  │ Bids │      │ Asks │           │     │
	│  │  └──────┘      └──────┘           │     │
	│  └───────────────────────────────────┘     │
	│  ┌───────────────────────────────────┐     │
	│  │         Matcher                   │     │
	│  │  Execute │ Amend │ Cancel         │     │
	│  └──────────┬────────────────────────┘     │
	│             ▼                              │
	│  ┌──────────────────────────────────┐     │
	│  │     Event Publisher              │     │
	│  │  (Trade, Fill, Cancel, Reject)   │     │
	│  └──────────────────────────────────┘     │
	└─────────────────────────────────────────────┘

# Quick Start

Basic usage with a single market:

	package main

	import (
		"context"
		"fmt"
		"time"

		"github.com/adimiuprix/spot-engine/engine"
		"github.com/adimiuprix/spot-engine/event"
		"github.com/adimiuprix/spot-engine/order"
		"github.com/adimiuprix/spot-engine/protocol"
		"github.com/shopspring/decimal"
	)

	func main() {
		// Create matching engine
		publisher := event.NewChannelPublisher(10000)
		eng := engine.NewMatchingEngine(publisher)

		// Start engine
		eng.Start()
		defer eng.Stop()

		// Listen to events
		go func() {
			for log := range publisher.Channel() {
				switch log.LogType {
				case event.LogTypeTrade:
					fmt.Printf("Trade: %s @ %s\n",
						log.TradeQuantity, log.TradePrice)
				case event.LogTypeFill:
					fmt.Printf("Fill: Order %s, %s\n",
						log.OrderID, log.FillQuantity)
				}
			}
		}()

		// Create market
		ctx := context.Background()
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

		// Place orders
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
			Price:     decimal.NewFromInt(50000),
			Size:      decimal.NewFromFloat(0.1),
		}
		eng.SubmitOrder(placeReq)
	}

# Performance

Benchmark results on Intel Core i5-3330 @ 3.00GHz:

  - BestBid/BestAsk: 23ns, 0 allocations (44M ops/sec)
  - Market Order: 288ns (3.5M ops/sec)
  - Limit Order: 687ns (1.5M ops/sec)
  - Full Match: 3.2µs with event emission

The engine achieves sub-microsecond latency for critical operations,
making it suitable for high-frequency trading applications.

# Packages

The SDK is organized into focused packages:

  - book: OrderBook and price level management using B-Tree
  - engine: Core engine and multi-market orchestration
  - event: Event logging and publishing system
  - matcher: Order matching logic and execution
  - order: Order types and validation
  - protocol: Request/response protocol with validation
  - queue: Ring buffer for command queue
  - snapshot: Snapshot and restore functionality
  - trade: Trade record structures

# Design Principles

1. Deterministic Replay: All timestamps come from upstream, no time.Now()
2. Event Sourcing: Every state change emits an immutable event log
3. Type Safety: Strongly-typed requests with validation before processing
4. State Enforcement: Market states (running/suspended/halted) strictly enforced
5. Precision Control: Decimal arithmetic, no float rounding errors

# Event Sourcing

All operations emit events for audit trail and replay:

  - Trade: A match between two orders
  - Fill: Partial or full order execution
  - Cancel: Order cancelled by user or system
  - Reject: Order rejected due to validation or state
  - Admin: Market state changes (suspend, resume, halt)
  - Replenish: Iceberg order replenishment

Events include CommandID for idempotency and Timestamp for ordering.

# Thread Safety

The engine uses a single-threaded event loop for deterministic execution.
Multiple markets run on the same thread, ensuring consistent ordering.
External callers submit commands via a ring buffer, and receive results
via Future objects or event callbacks.

# Snapshot & Recovery

The engine supports point-in-time snapshots for disaster recovery:

	// Take snapshot
	snapshots, seqID := eng.TakeSnapshot()
	writer := snapshot.NewWriter("./snapshots")
	writer.WriteSnapshot(snapshots, seqID)

	// Restore from snapshot
	reader := snapshot.NewReader("./snapshots")
	metadata, snapshots, _ := reader.ReadSnapshot()
	eng.RestoreFromSnapshot(snapshots)

Snapshots include CRC32 checksums for integrity validation and use
atomic file writes (temp + rename) to prevent corruption.

# Production Readiness

The engine has been thoroughly tested and benchmarked:

  - 142 unit tests with 97.7% coverage on critical paths
  - 12+ integration examples demonstrating real-world usage
  - Comprehensive benchmarks showing HFT-grade performance
  - Production-ready audit score: 9.1/10

See PRODUCTION_READINESS_AUDIT.md for detailed analysis.

# License

MIT License - see LICENSE file for details.

# Links

  - GitHub: https://github.com/adimiuprix/spot-engine
  - Documentation: https://pkg.go.dev/github.com/adimiuprix/spot-engine
  - Examples: https://github.com/adimiuprix/spot-engine/tree/main/example

*/
package spotengine
