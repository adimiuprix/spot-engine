# Spot Engine Examples

Koleksi contoh penggunaan Spot Engine untuk berbagai use case.

## 📁 Struktur Examples

### 1. **Simple** (`simple/`)
Contoh paling dasar menggunakan Spot Engine:
- Membuat engine
- Submit order buy/sell
- Melihat trades
- Melihat order book

```bash
cd example/simple
go run main.go
```

### 2. **Trading** (`trading/`)
Contoh trading normal dengan limit orders:
- Multiple buy dan sell orders
- Order matching
- Trade execution

```bash
cd example/trading
go run main.go
```

### 3. **Async Place Order** (`async_place_order/`) ⭐ RECOMMENDED
Contoh lengkap place order dengan Async API:
- PlaceOrderAsync untuk limit orders
- PlaceOrderAsync untuk market orders  
- Detailed results (filled, remaining, trades)
- Error handling (validation vs execution)
- Event monitoring real-time
- Context support (timeout, cancellation)

```bash
cd example/async_place_order
go run main.go
```

### 4. **Async Trading** (`async_trading/`) ⭐ FULL DEMO
Contoh async trading API dengan Future pattern:
- PlaceOrderAsync dengan detailed results
- CancelOrderAsync dengan ownership check
- AmendOrderAsync dengan priority rules
- Context support (timeout, cancellation)
- Full error handling

```bash
cd example/async_trading
go run main.go
```

### 4. **Market Order** (`market_order/`)
Contoh market order execution:
- Market buy orders
- Market sell orders
- Immediate execution tanpa price limit

```bash
cd example/market_order
go run main.go
```

### 4. **Time-in-Force (TIF)** (`tif/`)
Contoh berbagai tipe Time-in-Force:
- **GTC** (Good-Til-Cancel): Order tetap di book sampai filled atau cancelled
- **IOC** (Immediate-or-Cancel): Match langsung atau cancel
- **FOK** (Fill-or-Kill): Harus fully filled atau reject seluruhnya
- **PostOnly**: Harus jadi maker, reject jika match langsung

```bash
cd example/tif
go run main.go
```

### 5. **Amend Order** (`amend/`)
Contoh modify order yang sudah ada:
- Ubah harga order
- Ubah quantity order
- Priority loss handling

```bash
cd example/amend
go run main.go
```

### 6. **Iceberg Order** (`iceberg/`)
Contoh iceberg orders (hidden quantity):
- Display quantity vs total quantity
- Automatic replenishment
- Hidden liquidity

```bash
cd example/iceberg
go run main.go
```

### 7. **Snapshot** (`snapshot/`)
Contoh snapshot dan restore order book:
- Save order book state
- Restore dari snapshot
- Persistence

```bash
cd example/snapshot
go run main.go
```

### 8. **State Management** (`state_management/`)
Contoh state management lengkap:
- Order lifecycle
- Event handling
- State tracking

```bash
cd example/state_management
go run main.go
```

### 9. **Management** (`management/`)
Contoh management API:
- Query order book
- Cancel orders
- Get order status

```bash
cd example/management
go run main.go
```

## 🚀 Cara Menjalankan Semua Examples

```bash
# Run all examples
cd example
for /D %d in (*) do (cd %d && if exist main.go go run main.go && cd ..)
```

## 📝 Notes

- Semua examples menggunakan package `main`
- Setiap example berdiri sendiri di folder masing-masing
- Tidak ada dependency antar examples
- Bisa dijalankan secara independen

## 🔧 Requirements

- Go 1.23+
- Dependencies (otomatis diinstall via `go mod`):
  - `github.com/shopspring/decimal`
  - `github.com/google/btree`

## 📖 Learn More

Lihat dokumentasi lengkap di:
- [Design Docs](../docs/design/)
- [Benchmark](../docs/benchmark.md)
- [README](../README.md)
