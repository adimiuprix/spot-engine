# Async Place Order Example

Contoh lengkap cara menggunakan **Async API** untuk place order dengan Future pattern.

## 📁 Files

1. **`main.go`** - Example lengkap dengan berbagai scenario
2. **`simple_example.go`** - Example paling sederhana untuk pemula

## 🚀 Cara Run

```bash
cd example/async_place_order
go run main.go
```

Atau run simple example:
```bash
go run simple_example.go
```

## 📖 What This Example Shows

### 1. Basic Setup
```go
// Create publisher untuk event monitoring
publisher := event.NewChannelPublisher(10000)

// Create matching engine
eng := engine.NewMatchingEngine(publisher)
defer eng.Shutdown()

// Create context (untuk timeout/cancellation)
ctx := context.Background()
```

### 2. Create Market
```go
createReq := &protocol.CreateMarketRequest{
    BaseCommand: protocol.BaseCommand{
        CommandID: "cmd-1",
        UserID:    1,
        MarketID:  "BTC/USDT",
        Timestamp: time.Now().UnixNano(),
    },
    MinLotSize: decimal.NewFromFloat(0.00000001),
}

future, err := eng.CreateMarket(ctx, createReq)
if err != nil {
    // Validation error
    return err
}

_, err = future.Wait(ctx)
if err != nil {
    // Execution error
    return err
}
```

### 3. Place Limit Order
```go
placeReq := &protocol.PlaceOrderRequest{
    BaseCommand: protocol.BaseCommand{
        CommandID: "cmd-2",
        UserID:    100,
        MarketID:  "BTC/USDT",
        Timestamp: time.Now().UnixNano(),
    },
    OrderID:   "my-order-123",
    Side:      "buy",                    // "buy" or "sell"
    OrderType: "limit",                  // "limit" or "market"
    Price:     decimal.NewFromInt(50000), // Price for limit order
    Size:      decimal.NewFromFloat(1.0), // Quantity
}

future, err := eng.PlaceOrderAsync(ctx, placeReq)
if err != nil {
    // Validation error (bad request)
    return err
}

result, err := future.Wait(ctx)
if err != nil {
    // Execution error (market not found, etc.)
    return err
}

// Check result
fmt.Printf("Filled: %s\n", result.Filled.String())
fmt.Printf("In Book: %v\n", result.InBook)
fmt.Printf("Trades: %d\n", len(result.Trades))
```

### 4. Place Market Order
```go
marketReq := &protocol.PlaceOrderRequest{
    BaseCommand: protocol.BaseCommand{
        CommandID: "cmd-3",
        UserID:    200,
        MarketID:  "BTC/USDT",
        Timestamp: time.Now().UnixNano(),
    },
    OrderID:   "market-order-456",
    Side:      "buy",
    OrderType: "market",                 // Market order
    Price:     decimal.Zero,             // No price for market order
    Size:      decimal.NewFromFloat(0.5),
}

future, err := eng.PlaceOrderAsync(ctx, marketReq)
// ... same pattern ...
```

## 📊 Example Output

```
=== Async Place Order Example ===

📊 Creating market BTC/USDT...
✅ Market created successfully

📝 Placing buy order (1.0 BTC @ $50,000)...
✅ BUY Order Result:
   OrderID: order-buy-1784793799981053900
   Accepted: true
   Filled: 0 BTC
   Remaining: 1 BTC
   In Order Book: true
   Trades Generated: 0
   📖 Order is resting in book (waiting for match)

📝 Placing sell order that will match (0.5 BTC @ $50,000)...
      🔄 Trade #1: 0.5 BTC @ $50000
✅ SELL Order Result:
   OrderID: order-sell-1784793800082443700
   Accepted: true
   Filled: 0.5 BTC
   Remaining: 0 BTC
   Trades Generated: 3
   💰 Filled Amount: 0.5 BTC

📊 Final Market Statistics:
  Market ID: BTC/USDT
  State: running
  Bid Levels: 2
  Best Bid: $50000

✨ Done! All orders placed successfully.
```

## 🎯 Key Features Demonstrated

### ✅ Detailed Error Handling
```go
future, err := eng.PlaceOrderAsync(ctx, req)
if err != nil {
    // Validation errors:
    // - ErrInvalidCommandID
    // - ErrInvalidPrice
    // - ErrInvalidSize
    // - ErrInvalidSide
    // - etc.
}

result, err := future.Wait(ctx)
if err != nil {
    // Execution errors:
    // - ErrNotFound (market not found)
    // - ErrMarketSuspended
    // - ErrUnauthorized
}
```

### ✅ Detailed Results
```go
type PlaceOrderResult struct {
    OrderID      string          // Order identifier
    Accepted     bool            // True if accepted
    Filled       decimal.Decimal // Amount filled
    Remaining    decimal.Decimal // Amount remaining
    Trades       []interface{}   // Trades generated
    InBook       bool            // True if resting in book
    PartialFill  bool            // True if partially filled
}
```

### ✅ Context Support
```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

future, err := eng.PlaceOrderAsync(ctx, req)
result, err := future.Wait(ctx)
if err == context.DeadlineExceeded {
    fmt.Println("Request timed out")
}

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(2 * time.Second)
    cancel() // Cancel after 2 seconds
}()

result, err := future.Wait(ctx)
if err == context.Canceled {
    fmt.Println("Request cancelled")
}
```

### ✅ Event Monitoring
```go
go func() {
    for log := range publisher.Channel() {
        switch log.LogType {
        case "trade":
            fmt.Printf("Trade: %s @ %s\n", 
                log.TradeQuantity, log.TradePrice)
        case "fill":
            fmt.Printf("Fill: Order %s filled %s\n", 
                log.OrderID, log.FilledSize)
        case "reject":
            fmt.Printf("Reject: %s\n", log.RejectReason)
        }
    }
}()
```

## 🔍 What Gets Checked

### Request Validation (Before Execution)
- ✅ CommandID is not empty
- ✅ Timestamp is positive
- ✅ MarketID is not empty
- ✅ OrderID is not empty
- ✅ Side is "buy" or "sell"
- ✅ OrderType is "limit" or "market"
- ✅ Price is positive (for limit orders)
- ✅ Size is positive

### Execution Checks (During Processing)
- ✅ Market exists
- ✅ Market state allows new orders
- ✅ Order matching with existing orders
- ✅ Time-in-Force rules (GTC, IOC, FOK, PostOnly)
- ✅ MinLotSize constraints
- ✅ Iceberg order handling

## 📚 Related Examples

- **`example/simple/`** - Non-async basic example
- **`example/trading/`** - Non-async trading example
- **`example/async_trading/`** - Full async API demo (place, cancel, amend)

## 💡 Tips

1. **Always use unique CommandID** - untuk idempotency
   ```go
   CommandID: fmt.Sprintf("cmd-%d", time.Now().UnixNano())
   ```

2. **Use context with timeout** - untuk production
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   ```

3. **Check result details** - jangan cuma cek error
   ```go
   if result.PartialFill {
       fmt.Println("Order partially filled!")
   }
   if result.InBook {
       fmt.Println("Order is in book, waiting for match")
   }
   ```

4. **Monitor events** - untuk real-time updates
   ```go
   go func() {
       for log := range publisher.Channel() {
           // Process events
       }
   }()
   ```

## 🚀 Next Steps

After understanding this example, try:

1. **Cancel Order** - `example/async_trading/` shows CancelOrderAsync
2. **Amend Order** - `example/async_trading/` shows AmendOrderAsync
3. **Batch Operations** - Place multiple orders efficiently
4. **Error Recovery** - Implement retry logic with exponential backoff

## 📖 Documentation

- Full API docs: `ASYNC_TRADING_API.md`
- Comparison: `ASYNC_VS_SYNC_COMPARISON.md`
- Implementation status: `IMPLEMENTATION_STATUS.md`
