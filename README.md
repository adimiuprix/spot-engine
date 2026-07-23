# Spot Engine - High-Performance Matching Engine

A deterministic, low-latency matching engine for spot trading built in Go. Designed for production use with comprehensive features including iceberg orders, order amendments, market management, and snapshot/restore capabilities.

## 🚀 Features

- **High Performance**: ~23ns best price lookup with 0 allocations using B-Tree structure
- **Deterministic Replay**: All events use upstream-assigned timestamps for reproducible behavior
- **Iceberg Orders**: Hide order quantity with automatic replenishment
- **Order Amendments**: Modify orders with proper priority rules
- **Market Management**: Create, suspend, resume markets with state enforcement
- **Snapshot & Restore**: Point-in-time recovery with CRC32 validation
- **Event Logging**: Comprehensive audit trail for all operations
- **Type-Safe Protocol**: Strongly-typed request/response with validation

## 📊 Performance

Based on benchmark results (AMD Ryzen 7 PRO 4750U):

- **BestBid/BestAsk**: ~23 ns/op, 0 B/op, 0 allocs/op
- **O(log n)** price level operations using B-Tree
- **Deterministic** single-threaded event loop for consistency

## 🏗️ Architecture

### Core Components

```
┌─────────────────────────────────────────────────────┐
│                 MatchingEngine                      │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐      │
│  │  Market   │  │  Market   │  │  Market   │      │
│  │ BTC-USDT  │  │ ETH-USDT  │  │ SOL-USDT  │      │
│  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘      │
│        │              │              │             │
│   ┌────▼──────────────▼──────────────▼────┐        │
│   │         OrderBook (B-Tree)            │        │
│   │  ┌──────────┐      ┌──────────┐       │        │
│   │  │ BidTree  │      │ AskTree  │       │        │
│   │  │(Descend) │      │(Ascend)  │       │        │
│   │  └──────────┘      └──────────┘       │        │
│   └───────────────────────────────────────┘        │
│                                                     │
│   ┌──────────────────────────────────────┐         │
│   │           Matcher                     │         │
│   │  ┌────────────┐  ┌────────────┐      │         │
│   │  │ Execute    │  │ Replenish  │      │         │
│   │  │ Amend      │  │ Cancel     │      │         │
│   │  └────────────┘  └────────────┘      │         │
│   └──────────────────────────────────────┘         │
│                      │                              │
│                      ▼                              │
│   ┌──────────────────────────────────────┐         │
│   │        Event Publisher                │         │
│   │  (Trade, Fill, Cancel, Reject, Admin) │         │
│   └──────────────────────────────────────┘         │
└─────────────────────────────────────────────────────┘
```

### Key Design Principles

1. **Deterministic Replay**: All timestamps from upstream, no `time.Now()`
2. **Event Sourcing**: Every state change emits an event log
3. **Type Safety**: Strongly-typed requests with validation
4. **State Enforcement**: Market states (running/suspended/halted) enforced
5. **Precision Control**: MinLotSize prevents micro-remainder loops

## 🎯 Quick Start

### Installation

```bash
go get github.com/adimiuprix/spot-engine
```

### Basic Usage

```go
package main

import (
    "github.com/adimiuprix/spot-engine/engine"
    "github.com/adimiuprix/spot-engine/event"
    "github.com/shopspring/decimal"
)

func main() {
    // Create engine
    publisher := event.NewChannelPublisher(10000)
    eng := engine.NewMatchingEngine(publisher)

    // Listen to events
    go func() {
        for log := range publisher.Channel() {
            switch log.LogType {
            case event.LogTypeTrade:
                fmt.Printf("Trade: %s @ %s\n", 
                    log.TradeQuantity, log.TradePrice)
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
}
```

## 📚 Examples

Run the included examples:

```bash
# Basic trading
go run ./example

# Management commands
go run ./example mgmt

# Iceberg orders
go run ./example iceberg

# Order amendments
go run ./example amend

# Snapshot & restore
go run ./example snapshot
```

## 🔑 Key Features

### 1. Iceberg Orders

Hide large orders by showing only a portion:

```go
order := &order.Order{
    OrderID:   "iceberg-1",
    Side:      order.Sell,
    Price:     decimal.NewFromInt(50000),
    Quantity:  decimal.NewFromFloat(10.0),  // Total
}
order.SetupIceberg(decimal.NewFromFloat(1.0))  // Show 1.0 at a time
```

When visible portion is consumed, automatically replenishes from hidden and moves to tail of queue.

### 2. Order Amendments

Modify orders with priority rules:

- **Size Decrease + Same Price**: Keeps priority (in-place update)
- **Size Increase**: Loses priority (re-match as fresh order)
- **Price Change**: Loses priority (re-match as fresh order)

```go
amendReq := &protocol.AmendOrderRequest{
    OrderID:  "order-1",
    NewPrice: decimal.NewFromInt(50000),
    NewSize:  decimal.NewFromFloat(0.5),
}
result := matcher.ProcessAmend(amendReq)
```

### 3. Market Management

Full lifecycle management:

```go
// Create
eng.CreateMarket(ctx, createReq)

// Suspend (only cancel allowed)
eng.SuspendMarket(ctx, suspendReq)

// Resume
eng.ResumeMarket(ctx, resumeReq)

// Update config
eng.UpdateConfig(ctx, updateReq)
```

### 4. Snapshot & Restore

Point-in-time recovery:

```go
// Take snapshot
snapshots, seqID := eng.TakeSnapshot()
writer := snapshot.NewWriter("./snapshots")
writer.WriteSnapshot(snapshots, seqID)

// Restore after crash
reader := snapshot.NewReader("./snapshots")
metadata, snapshots, _ := reader.ReadSnapshot()
eng.RestoreFromSnapshot(snapshots)
```

## 📖 Documentation

- [Architecture](./docs/ARCHITECTURE.md) - System design and principles
- [API Reference](./docs/API.md) - Complete API documentation
- [Protocol](./docs/PROTOCOL.md) - Request/response formats
- [Events](./docs/EVENTS.md) - Event model and log types
- [Benchmarks](./docs/BENCHMARKS.md) - Performance metrics

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run benchmarks
go test -bench=. ./book/...

# Run with race detector
go test -race ./...
```

## 📦 Project Structure

```
spot-engine/
├── book/           # OrderBook and price level management
├── engine/         # Core engine and market management
├── event/          # Event logging and publishing
├── matcher/        # Order matching logic
├── order/          # Order types and validation
├── protocol/       # Request/response protocol
├── queue/          # Ring buffer implementation
├── snapshot/       # Snapshot & restore
├── trade/          # Trade records
└── example/        # Usage examples
```

## 🔒 Security & Safety

- **Deterministic**: No `time.Now()` in business logic
- **Type-Safe**: Validation before enqueue
- **State Enforcement**: Market state rules enforced
- **Audit Trail**: Complete event log with CommandID
- **Checksum Validation**: Snapshot integrity verified

## 🤝 Contributing

Contributions are welcome! Please ensure:

1. All tests pass
2. Code follows Go conventions
3. Documentation updated
4. Examples provided for new features

## 📄 License

MIT License - see LICENSE file for details

## 🙏 Acknowledgments

Built with reference to production-grade matching engines and best practices from:
- Financial exchange architectures
- LMAX Disruptor pattern
- Event sourcing principles

## 📞 Support

- Issues: GitHub Issues
- Documentation: `/docs` directory
- Examples: `/example` directory

---

**Version**: 0.8.0  
**Go Version**: 1.23+  
**Status**: Production Ready
