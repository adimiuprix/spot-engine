# Async vs Non-Async API Comparison

## Overview

Project ini memiliki **2 API** untuk trading:

1. **Non-Async API** (Basic Engine) - Simple, langsung
2. **Async API** (MatchingEngine) - Professional, dengan Future pattern

---

## Perbedaan Utama

| Aspek | Non-Async (Basic Engine) | Async (MatchingEngine) |
|-------|--------------------------|------------------------|
| **Return Type** | `bool` (success/fail) | `Future[Result]` (detailed) |
| **Error Info** | ❌ Tidak ada detail | ✅ Ada detail error |
| **Result Info** | ❌ Tidak tahu filled berapa | ✅ Tahu filled, remaining, trades |
| **Validation** | ❌ Tidak ada pre-validation | ✅ Ada pre-validation |
| **Context Support** | ❌ Tidak ada | ✅ Ada (timeout, cancel) |
| **Event Logging** | ✅ Ada | ✅ Ada |
| **Blocking** | ✅ Non-blocking (push ke queue) | ✅ Non-blocking (goroutine) |
| **Use Case** | Simple apps, testing | Production, enterprise |

---

## Contoh Kode

### 1. Place Order

#### ❌ Non-Async (Basic Engine)

```go
// Create basic engine
config := engine.Config{
    Symbol:         "BTC/USDT",
    RingBufferSize: 10000,
}
eng := engine.New(config)
eng.Start()

// Submit order - hanya return bool
order := &order.Order{
    ID:       1,
    OrderID:  "order-1",
    UserID:   100,
    Symbol:   "BTC/USDT",
    Side:     order.Buy,
    Type:     order.Limit,
    Price:    decimal.NewFromInt(50000),
    Quantity: decimal.NewFromFloat(1.0),
    Filled:   decimal.Zero,
}

success := eng.SubmitOrder(order)

if !success {
    // ❌ MASALAH: Tidak tahu kenapa gagal!
    // - Queue penuh?
    // - Market tidak ada?
    // - Order invalid?
    // - Tidak ada info!
    fmt.Println("Failed, tapi gak tau kenapa...")
}

// ❌ MASALAH: Tidak tahu order di-fill berapa
// - Apakah fully filled?
// - Apakah partially filled?
// - Apakah masuk order book?
// - Berapa trade yang terjadi?
// - Tidak ada info!
```

#### ✅ Async (MatchingEngine)

```go
// Create matching engine
publisher := event.NewChannelPublisher(10000)
eng := engine.NewMatchingEngine(publisher)

// Create market first
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
    // ✅ VALIDATION ERROR - tahu langsung sebelum diproses
    fmt.Printf("Invalid request: %v\n", err)
    return
}

_, err = createFuture.Wait(ctx)
if err != nil {
    // ✅ EXECUTION ERROR - tahu kenapa gagal
    fmt.Printf("Execution failed: %v\n", err)
    return
}

// Place order dengan detailed request
placeReq := &protocol.PlaceOrderRequest{
    BaseCommand: protocol.BaseCommand{
        CommandID: "cmd-place-1",
        UserID:    100,
        MarketID:  "BTC/USDT",
        Timestamp: time.Now().UnixNano(),
    },
    OrderID:   "order-1",
    Side:      "buy",
    OrderType: "limit",
    Price:     decimal.NewFromInt(50000),
    Size:      decimal.NewFromFloat(1.0),
}

placeFuture, err := eng.PlaceOrderAsync(ctx, placeReq)
if err != nil {
    // ✅ VALIDATION ERROR - langsung tahu error apa
    // - Invalid CommandID
    // - Invalid Price
    // - Invalid Size
    fmt.Printf("Validation error: %v\n", err)
    return
}

result, err := placeFuture.Wait(ctx)
if err != nil {
    // ✅ EXECUTION ERROR - tahu kenapa gagal
    // - Market not found
    // - Market suspended
    // - Unauthorized
    fmt.Printf("Execution error: %v\n", err)
    return
}

// ✅ DETAILED RESULT - tahu semua info!
fmt.Printf("Order placed successfully!\n")
fmt.Printf("  OrderID: %s\n", result.OrderID)
fmt.Printf("  Filled: %s\n", result.Filled.String())
fmt.Printf("  Remaining: %s\n", result.Remaining.String())
fmt.Printf("  In Book: %v\n", result.InBook)
fmt.Printf("  Partial Fill: %v\n", result.PartialFill)
fmt.Printf("  Trades: %d\n", len(result.Trades))
```

---

### 2. Cancel Order

#### ❌ Non-Async (Basic Engine)

```go
// Basic engine tidak punya cancel API!
// Harus manual remove dari order book atau pakai matcher

// Option 1: Manual remove (tidak aman)
removed := eng.GetOrderBook().RemoveOrder(order)
if !removed {
    // ❌ Tidak tahu kenapa gagal
    fmt.Println("Failed to remove")
}

// ❌ MASALAH:
// - Tidak ada ownership check
// - Tidak ada state check
// - Tidak ada event log
// - Tidak ada error detail
```

#### ✅ Async (MatchingEngine)

```go
cancelReq := &protocol.CancelOrderRequest{
    CommandID: "cmd-cancel-1",
    UserID:    100,
    Symbol:    "BTC/USDT",
    OrderID:   "order-1",
    Timestamp: time.Now().UnixNano(),
}

cancelFuture, err := eng.CancelOrderAsync(ctx, cancelReq)
if err != nil {
    // ✅ Validation error
    fmt.Printf("Invalid request: %v\n", err)
    return
}

result, err := cancelFuture.Wait(ctx)
if err != nil {
    // ✅ Execution error dengan detail
    // - Order not found
    // - Not owner (unauthorized)
    // - Market suspended
    fmt.Printf("Cancel failed: %v\n", err)
    return
}

// ✅ DETAILED RESULT
fmt.Printf("Order cancelled!\n")
fmt.Printf("  OrderID: %s\n", result.OrderID)
fmt.Printf("  Filled Before: %s\n", result.FilledBefore.String())
fmt.Printf("  Cancelled Size: %s\n", result.CancelledSize.String())

// ✅ PLUS:
// - Ownership check (user 100 cannot cancel user 101's order)
// - State check (cannot cancel if market halted)
// - Event log emitted (for audit trail)
```

---

### 3. Timeout & Cancellation

#### ❌ Non-Async (Basic Engine)

```go
// Tidak ada cara untuk set timeout atau cancel operation
success := eng.SubmitOrder(order)

// Jika stuck atau lambat, tidak bisa cancel
// Harus tunggu selesai atau kill process
```

#### ✅ Async (MatchingEngine)

```go
// Set timeout 5 detik
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

placeFuture, err := eng.PlaceOrderAsync(ctx, placeReq)
if err != nil {
    return err
}

result, err := placeFuture.Wait(ctx)
if err == context.DeadlineExceeded {
    // ✅ Timeout setelah 5 detik
    fmt.Println("Request timeout after 5 seconds")
    return
}

// Atau manual cancel
go func() {
    time.Sleep(2 * time.Second)
    cancel() // Cancel operation after 2 seconds
}()

result, err := placeFuture.Wait(ctx)
if err == context.Canceled {
    // ✅ Operation di-cancel
    fmt.Println("Request cancelled")
    return
}
```

---

## Error Handling Comparison

### Non-Async
```go
success := eng.SubmitOrder(order)
if !success {
    // ❌ Tidak tahu kenapa gagal!
    // Harus cek event log manual untuk tahu
    log.Println("Failed") 
}
```

### Async
```go
// Validation error (immediate)
future, err := eng.PlaceOrderAsync(ctx, req)
if err != nil {
    switch err {
    case protocol.ErrInvalidCommandID:
        // Handle invalid command ID
    case protocol.ErrInvalidPrice:
        // Handle invalid price
    case protocol.ErrInvalidSize:
        // Handle invalid size
    }
}

// Execution error (from future)
result, err := future.Wait(ctx)
if err != nil {
    switch err {
    case protocol.ErrNotFound:
        // Market or order not found
    case protocol.ErrMarketSuspended:
        // Market is suspended
    case protocol.ErrUnauthorized:
        // User doesn't own the order
    }
}
```

---

## Use Cases

### ✅ Kapan Pakai Non-Async (Basic Engine)?

1. **Simple Applications**
   - Testing
   - Prototyping
   - Internal tools

2. **Fire-and-Forget**
   - Tidak perlu tahu hasilnya
   - Hanya perlu submit order

3. **High Throughput, Low Latency**
   - Push to ring buffer langsung
   - Minimal overhead
   - Event log untuk monitoring

**Example:**
```go
// Testing matching logic
eng := engine.New(config)
eng.Start()

order1 := createBuyOrder()
order2 := createSellOrder()

eng.SubmitOrder(order1)
eng.SubmitOrder(order2)

// Monitor events untuk lihat hasil
go func() {
    for log := range eng.Events() {
        if log.LogType == "trade" {
            fmt.Printf("Trade: %s @ %s\n", 
                log.TradeQuantity, log.TradePrice)
        }
    }
}()
```

---

### ✅ Kapan Pakai Async (MatchingEngine)?

1. **Production Applications**
   - Need error handling
   - Need result details
   - Need audit trail

2. **Client-Facing APIs**
   - REST API
   - WebSocket API
   - gRPC API

3. **Complex Workflows**
   - Multi-step operations
   - Need to know if succeeded
   - Need to handle errors

4. **Enterprise Systems**
   - Audit requirements
   - Compliance
   - Monitoring & alerting

**Example:**
```go
// REST API handler
func handlePlaceOrder(w http.ResponseWriter, r *http.Request) {
    var req protocol.PlaceOrderRequest
    json.NewDecoder(r.Body).Decode(&req)

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    // Async with proper error handling
    future, err := engine.PlaceOrderAsync(ctx, &req)
    if err != nil {
        // Validation error - 400 Bad Request
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    result, err := future.Wait(ctx)
    if err != nil {
        // Execution error - 500 or specific error code
        if err == protocol.ErrNotFound {
            http.Error(w, "Market not found", http.StatusNotFound)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    // Success - return detailed result
    json.NewEncoder(w).Encode(result)
}
```

---

## Performance Comparison

| Aspect | Non-Async | Async |
|--------|-----------|-------|
| **Throughput** | ⚡ Highest (direct push) | 🔵 High (goroutine overhead) |
| **Latency** | ⚡ Lowest (no validation) | 🔵 Low (pre-validation) |
| **Memory** | ⚡ Minimal | 🔵 Slightly more (Future objects) |
| **CPU** | ⚡ Minimal | 🔵 Slightly more (goroutines) |
| **Error Handling** | ❌ Poor | ✅ Excellent |
| **Observability** | ⚠️ Event logs only | ✅ Results + event logs |
| **Production Ready** | ⚠️ Limited | ✅ Full featured |

**Verdict:** 
- Non-Async: Fastest, tapi minimal features
- Async: Sedikit lebih lambat (negligible), tapi production-ready

---

## Code Size Comparison

### Non-Async
```go
// 5 lines - simple tapi tidak robust
eng := engine.New(config)
eng.Start()
success := eng.SubmitOrder(order)
if !success {
    log.Println("Failed")
}
```

### Async
```go
// 20 lines - lebih verbose tapi robust
eng := engine.NewMatchingEngine(publisher)

req := &protocol.PlaceOrderRequest{...}

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

future, err := eng.PlaceOrderAsync(ctx, req)
if err != nil {
    return fmt.Errorf("validation: %w", err)
}

result, err := future.Wait(ctx)
if err != nil {
    return fmt.Errorf("execution: %w", err)
}

fmt.Printf("Filled: %s, InBook: %v\n", 
    result.Filled, result.InBook)
```

**Trade-off:** Async lebih verbose, tapi jauh lebih reliable dan informative.

---

## Migration Path

Jika Anda sudah pakai Non-Async, bisa migrate ke Async secara bertahap:

### Step 1: Keep both
```go
// Old code (masih jalan)
basicEngine := engine.New(config)
basicEngine.Start()
basicEngine.SubmitOrder(order)

// New code (parallel)
matchingEngine := engine.NewMatchingEngine(publisher)
// ... create market ...
matchingEngine.PlaceOrderAsync(ctx, req)
```

### Step 2: Migrate critical paths
```go
// Critical operations → Async
if isCriticalOrder {
    future, err := matchingEngine.PlaceOrderAsync(ctx, req)
    // ... handle result ...
} else {
    // Non-critical → keep simple
    basicEngine.SubmitOrder(order)
}
```

### Step 3: Full migration
```go
// Semua pakai async
matchingEngine := engine.NewMatchingEngine(publisher)
// ... all operations use async API ...
```

---

## Kesimpulan

### ✅ Non-Async (Basic Engine)
**Pros:**
- ⚡ Fastest
- 🎯 Simple API
- 💡 Easy to understand

**Cons:**
- ❌ No error details
- ❌ No result details
- ❌ No timeout/cancellation
- ❌ Not production-ready

**Best for:** Testing, prototyping, internal tools

---

### ✅ Async (MatchingEngine)
**Pros:**
- ✅ Detailed error handling
- ✅ Detailed results (filled, trades, etc.)
- ✅ Context support (timeout, cancel)
- ✅ Validation before execution
- ✅ Production-ready
- ✅ Enterprise features

**Cons:**
- 🔵 Slightly more overhead
- 📝 More verbose code

**Best for:** Production, client APIs, enterprise systems

---

## Rekomendasi

| Scenario | Use API |
|----------|---------|
| 🧪 Testing & development | Non-Async (simple) |
| 🚀 Production REST/gRPC API | **Async** ✅ |
| 🏢 Enterprise system | **Async** ✅ |
| ⚡ Ultra-high throughput | Non-Async (measure first!) |
| 📊 Trading platform | **Async** ✅ |
| 🔧 Internal tools | Non-Async or Async |
| 🌐 Public API | **Async** ✅ |

**Default choice untuk production: Async API** ✅

Performance difference is negligible, but robustness difference is HUGE!
