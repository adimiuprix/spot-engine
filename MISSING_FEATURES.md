# 🔴 Missing Features - Spot Engine

## 🚨 **HIGH PRIORITY (Safety Critical)**

### 1. **Market Orders** ❌ **CRITICAL**
**Status:** Type enum exists, NO logic implementation

**What's missing:**
- `processMarket()` function
- Base mode (Size-based): "Buy 1 BTC"
- Quote mode (QuoteSize-based): "Spend $50,000"
- Match at any price logic
- Integration with engine

**Risk:** Market orders won't execute at all

**Files to create:**
```
matcher/market_order.go
  - processMarket(o *order.Order) Result
  - matchMarketBuy(o *order.Order) Result
  - matchMarketSell(o *order.Order) Result
```

---

### 2. **LotSize Configuration** ❌ **SAFETY CRITICAL**
**Status:** Completely missing

**What's missing:**
- `MinLotSize` field in Config
- `DefaultLotSize` constant (1e-8)
- `WithLotSize()` option
- Termination check in market orders
- Snapshot persistence

**Risk:** Quote mode market orders can **INFINITE LOOP**

**Example bug scenario:**
```go
// Without LotSize:
quoteSize = 0.00000001 USDT
matchSize = 0.00000001 / 50000 = 0.0000000000002 BTC
// Loop forever with micro amounts! ❌
```

**Files to modify:**
```
engine/config.go
  - Add MinLotSize field

matcher/market_order.go
  - if matchSize < lotSize { reject remaining; break }
  
snapshot/snapshot.go
  - Persist MinLotSize
```

---

### 3. **Market State Management** ⚠️ **PARTIAL**
**Status:** Basic CreateMarket exists, no state control

**What's missing:**
- Market states: `Running`, `Suspended`, `Halted`
- `SuspendMarket()` command
- `ResumeMarket()` command  
- `HaltMarket()` command
- State enforcement in order processing
- Admin event logs

**Current behavior:**
- ❌ Can't suspend trading during emergency
- ❌ Can't halt market for maintenance
- ❌ Orders processed regardless of state

**Files to modify:**
```
protocol/state.go (NEW)
  - type OrderBookState uint8
  - const (Running, Suspended, Halted)

engine/matching_engine.go
  - SuspendMarket(req *protocol.SuspendMarketRequest)
  - ResumeMarket(req *protocol.ResumeMarketRequest)
  
engine/engine.go
  - State check in processNextOrder()
  - Reject if suspended/halted
```

---

## 🟡 **MEDIUM PRIORITY (Architecture)**

### 4. **Typed Protocol Layer** ⚠️ **SIMPLIFIED**
**Status:** Direct Order submission, no typed requests

**What's missing:**
- `BaseCommand` struct with CommandID/Timestamp/UserID
- `PlaceOrderRequest` typed struct
- `CancelOrderRequest` typed struct
- `AmendOrderRequest` typed struct
- Request validation before enqueue
- Binary serialization (MarshalRequest/UnmarshalRequest)

**Current approach:** Submit `Order` struct directly

**Decision:** Keep simplified for now, add validation layer

**Impact:** Medium - Architecture different but functional

---

### 5. **GlobalLastCmdSeqID Tracking** ❌
**Status:** Missing

**What's missing:**
- Track last processed command SeqID
- Store in snapshot metadata
- Return on restore for replay resume

**Current behavior:**
- ❌ No way to know where to resume replay after restore
- ❌ Must replay from beginning

**Files to modify:**
```
snapshot/snapshot.go
  - Add GlobalLastCmdSeqID to SnapshotMetadata
  - Track max SeqID across all markets
  - Return on restore
```

**Impact:** Medium - Replay coordination difficult

---

### 6. **Optimized Ring Buffer (Disruptor Pattern)** ⚠️
**Status:** Using basic channel-based queue

**What's missing:**
- MPSC ring buffer with Claim/Commit API
- Zero-allocation hot path
- Batch submission (ClaimN/CommitN)
- Explicit backpressure handling

**Current approach:** `queue.OrderQueue` (ring buffer basic)

**Impact:** Medium - Performance under extreme load

---

### 7. **Multi-Market Snapshot Format** ⚠️
**Status:** Single market only

**What's missing:**
- Footer with per-market segments
- Market offset/length/checksum in footer
- Multi-market serialization

**Current behavior:** Works for single market

**Impact:** Low - Single market sufficient for now

---

## 🟢 **LOW PRIORITY (Nice to Have)**

### 8. **Future Pattern (Async/Await)** ❌
**Status:** Not implemented

**What's missing:**
- `Future[T]` generic struct
- Response channel pooling
- `Wait(ctx)` for command completion
- Context timeout handling

**Current behavior:** Commands fire-and-forget

**Impact:** Low - Can poll state instead

---

### 9. **Depth Query API** ⚠️
**Status:** Basic only

**What's missing:**
- `Depth(limit uint32)` query
- Aggregated price levels
- Volume at each level
- Query via ring buffer

**Current:** Basic stats only

**Impact:** Low - Can iterate tree manually

---

### 10. **User Events** ❌
**Status:** Not implemented

**What's missing:**
- `UserEventRequest` type
- Custom event handling
- Event validation
- Event logs

**Use case:** Custom app-level events

**Impact:** Low - Can use admin logs

---

### 11. **Query Serialization** ❌
**Status:** Direct read access

**What's missing:**
- Queries via ring buffer
- Serialized with commands
- `ErrNotFound` for missing markets

**Current:** Direct method calls

**Impact:** Low - No race conditions in single-threaded

---

### 12. **Request Validation Layer** ⚠️
**Status:** Basic checks only

**What's missing:**
- Pre-enqueue validation
- Timestamp > 0 check
- CommandID non-empty check
- Size/Price positive checks
- Reject before queue entry

**Current:** Validated during processing

**Impact:** Low - Same result, different timing

---

## 📊 **Summary by Category**

| Priority | Count | Features |
|----------|-------|----------|
| 🔴 **HIGH (Safety)** | 3 | Market Orders, LotSize, State Management |
| 🟡 **MEDIUM (Architecture)** | 4 | Protocol, SeqID, Ring Buffer, Multi-Market |
| 🟢 **LOW (Nice to Have)** | 5 | Future, Depth, UserEvents, Query, Validation |
| **TOTAL** | **12** | |

---

## 🎯 **Recommended Implementation Order**

### **Phase 10: Safety Critical** (Week 1)
1. ✅ Market Orders (Base + Quote mode)
2. ✅ LotSize Configuration
3. ✅ Market State Management

### **Phase 11: Architecture** (Week 2)
4. ⚠️ GlobalLastCmdSeqID tracking
5. ⚠️ Request validation layer
6. ⚠️ Multi-market snapshot (if needed)

### **Phase 12: Enhancements** (Optional)
7. ⚠️ Future pattern (if async needed)
8. ⚠️ Depth query API (if market data needed)
9. ⚠️ Optimized ring buffer (if performance needed)

---

## 🚀 **Quick Start Priority**

**Must implement before production:**
- 🔴 Market Orders
- 🔴 LotSize Configuration
- 🔴 Market State Management

**Can defer:**
- Everything else (Medium/Low priority)

---

## 📝 **Notes**

- **Current completion: 60%** (39/65 features)
- **After Phase 10: ~75%** (48/65 features)
- **After Phase 11: ~85%** (54/65 features)

**Decision:** Focus on Phase 10 (safety critical) before any production deployment.
