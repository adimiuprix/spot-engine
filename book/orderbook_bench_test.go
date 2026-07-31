package book

import (
	"testing"

	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

func BenchmarkBestBid(b *testing.B) {
	book := NewOrderBook("BTC-USDT")

	// Add 1000 price levels
	for i := 0; i < 1000; i++ {
		o := &order.Order{
			ID:        uint64(i),
			OrderID:   "ORDER-" + string(rune(i)),
			CommandID: "cmd-" + string(rune(i)),
			UserID:    100,
			Symbol:    "BTC-USDT",
			Side:      order.Buy,
			Price:     decimal.NewFromInt(int64(50000 + i)),
			Quantity:  decimal.NewFromFloat(0.1),
			Filled:    decimal.Zero,
			Timestamp: int64(i),
		}
		book.Add(o)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = book.BestBid()
	}
}

func BenchmarkBestAsk(b *testing.B) {
	book := NewOrderBook("BTC-USDT")

	// Add 1000 price levels
	for i := 0; i < 1000; i++ {
		o := &order.Order{
			ID:        uint64(i),
			OrderID:   "ORDER-" + string(rune(i)),
			CommandID: "cmd-" + string(rune(i)),
			UserID:    100,
			Symbol:    "BTC-USDT",
			Side:      order.Sell,
			Price:     decimal.NewFromInt(int64(50000 + i)),
			Quantity:  decimal.NewFromFloat(0.1),
			Filled:    decimal.Zero,
			Timestamp: int64(i),
		}
		book.Add(o)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = book.BestAsk()
	}
}

func BenchmarkAddOrder(b *testing.B) {
	book := NewOrderBook("BTC-USDT")

	orders := make([]*order.Order, b.N)
	for i := 0; i < b.N; i++ {
		orders[i] = &order.Order{
			ID:        uint64(i),
			CommandID: "cmd-" + string(rune(i)),
			UserID:    100,
			Symbol:    "BTC-USDT",
			Side:      order.Buy,
			Price:     decimal.NewFromInt(int64(50000 + (i % 100))),
			Quantity:  decimal.NewFromFloat(0.1),
			Filled:    decimal.Zero,
			Timestamp: int64(i),
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		book.Add(orders[i])
	}
}

func BenchmarkGetDepth(b *testing.B) {
	book := NewOrderBook("BTC-USDT")

	// Add 500 bid levels and 500 ask levels
	for i := 0; i < 500; i++ {
		bid := &order.Order{
			ID:        uint64(i),
			CommandID: "cmd-bid-" + string(rune(i)),
			UserID:    100,
			Symbol:    "BTC-USDT",
			Side:      order.Buy,
			Price:     decimal.NewFromInt(int64(50000 - i)),
			Quantity:  decimal.NewFromFloat(0.1),
			Filled:    decimal.Zero,
			Timestamp: int64(i),
		}
		book.Add(bid)

		ask := &order.Order{
			ID:        uint64(i + 500),
			CommandID: "cmd-ask-" + string(rune(i)),
			UserID:    101,
			Symbol:    "BTC-USDT",
			Side:      order.Sell,
			Price:     decimal.NewFromInt(int64(51000 + i)),
			Quantity:  decimal.NewFromFloat(0.1),
			Filled:    decimal.Zero,
			Timestamp: int64(i + 500),
		}
		book.Add(ask)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = book.GetDepth(10)
	}
}

// BenchmarkRemoveOrder tests order removal performance
func BenchmarkRemoveOrder(b *testing.B) {
	book := NewOrderBook("BTC-USDT")

	// Pre-add orders
	orders := make([]*order.Order, b.N)
	for i := 0; i < b.N; i++ {
		o := &order.Order{
			ID:        uint64(i),
			CommandID: "cmd-" + string(rune(i)),
			UserID:    100,
			Symbol:    "BTC-USDT",
			Side:      order.Buy,
			Price:     decimal.NewFromInt(int64(50000 + (i % 100))),
			Quantity:  decimal.NewFromFloat(0.1),
			Filled:    decimal.Zero,
			Timestamp: int64(i),
		}
		orders[i] = o
		book.Add(o)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		book.RemoveOrder(orders[i])
	}
}

// BenchmarkFindOrder tests order lookup performance
func BenchmarkFindOrder(b *testing.B) {
	book := NewOrderBook("BTC-USDT")

	// Add 10000 orders
	orders := make([]*order.Order, 10000)
	for i := 0; i < 10000; i++ {
		o := &order.Order{
			ID:        uint64(i),
			OrderID:   "ORDER-" + string(rune(i)),
			CommandID: "cmd-" + string(rune(i)),
			UserID:    100,
			Symbol:    "BTC-USDT",
			Side:      order.Buy,
			Price:     decimal.NewFromInt(int64(50000 + (i % 100))),
			Quantity:  decimal.NewFromFloat(0.1),
			Filled:    decimal.Zero,
			Timestamp: int64(i),
		}
		orders[i] = o
		book.Add(o)
	}

	// Build OrderID strings
	orderIDs := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		orderIDs[i] = orders[i].OrderID
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = book.FindOrder(orderIDs[i%10000])
	}
}

// BenchmarkAddOrder_MultipleUsers tests realistic multi-user scenario
func BenchmarkAddOrder_MultipleUsers(b *testing.B) {
	book := NewOrderBook("BTC-USDT")

	orders := make([]*order.Order, b.N)
	for i := 0; i < b.N; i++ {
		side := order.Buy
		if i%2 == 0 {
			side = order.Sell
		}

		orders[i] = &order.Order{
			ID:        uint64(i),
			CommandID: "cmd-" + string(rune(i)),
			UserID:    uint64(100 + (i % 50)), // 50 different users
			Symbol:    "BTC-USDT",
			Side:      side,
			Price:     decimal.NewFromInt(int64(50000 + (i % 200) - 100)), // Spread across 200 price levels
			Quantity:  decimal.NewFromFloat(0.1 * float64(1+(i%10))),      // Varying sizes
			Filled:    decimal.Zero,
			Timestamp: int64(i),
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		book.Add(orders[i])
	}
}

// BenchmarkGetLevel tests price level lookup
func BenchmarkGetLevel(b *testing.B) {
	book := NewOrderBook("BTC-USDT")

	// Add 1000 price levels
	prices := make([]decimal.Decimal, 1000)
	for i := 0; i < 1000; i++ {
		price := decimal.NewFromInt(int64(50000 + i))
		prices[i] = price

		o := &order.Order{
			ID:        uint64(i),
			CommandID: "cmd-" + string(rune(i)),
			UserID:    100,
			Symbol:    "BTC-USDT",
			Side:      order.Buy,
			Price:     price,
			Quantity:  decimal.NewFromFloat(0.1),
			Filled:    decimal.Zero,
			Timestamp: int64(i),
		}
		book.Add(o)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = book.GetLevel(order.Buy, prices[i%1000])
	}
}

// BenchmarkOrderBook_MixedOperations simulates realistic workload
func BenchmarkOrderBook_MixedOperations(b *testing.B) {
	book := NewOrderBook("BTC-USDT")

	// Pre-populate with 1000 orders
	activeOrders := make([]*order.Order, 0, 1000)
	for i := 0; i < 1000; i++ {
		side := order.Buy
		if i%2 == 0 {
			side = order.Sell
		}

		o := &order.Order{
			ID:        uint64(i),
			CommandID: "cmd-" + string(rune(i)),
			UserID:    uint64(100 + (i % 50)),
			Symbol:    "BTC-USDT",
			Side:      side,
			Price:     decimal.NewFromInt(int64(50000 + (i % 100) - 50)),
			Quantity:  decimal.NewFromFloat(0.1),
			Filled:    decimal.Zero,
			Timestamp: int64(i),
		}
		book.Add(o)
		activeOrders = append(activeOrders, o)
	}

	orderCounter := uint64(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		op := i % 10

		switch {
		case op < 5: // 50% add
			side := order.Buy
			if i%2 == 0 {
				side = order.Sell
			}

			o := &order.Order{
				ID:        orderCounter,
				CommandID: "cmd-" + string(rune(int(orderCounter))),
				UserID:    uint64(100 + (i % 50)),
				Symbol:    "BTC-USDT",
				Side:      side,
				Price:     decimal.NewFromInt(int64(50000 + (i % 100) - 50)),
				Quantity:  decimal.NewFromFloat(0.1),
				Filled:    decimal.Zero,
				Timestamp: int64(i + 1000),
			}
			book.Add(o)
			activeOrders = append(activeOrders, o)
			orderCounter++

		case op < 8: // 30% remove
			if len(activeOrders) > 0 {
				idx := i % len(activeOrders)
				book.RemoveOrder(activeOrders[idx])
				// Remove from slice
				activeOrders = append(activeOrders[:idx], activeOrders[idx+1:]...)
			}

		case op == 8: // 10% BestBid/Ask
			_ = book.BestBid()
			_ = book.BestAsk()

		case op == 9: // 10% GetDepth
			_, _ = book.GetDepth(5)
		}
	}
}

// BenchmarkOrderBook_DeepBook tests performance with deep order book
func BenchmarkOrderBook_DeepBook(b *testing.B) {
	book := NewOrderBook("BTC-USDT")

	// Add 10000 orders across 1000 price levels (10 orders per level)
	for i := 0; i < 10000; i++ {
		side := order.Buy
		if i%2 == 0 {
			side = order.Sell
		}

		o := &order.Order{
			ID:        uint64(i),
			CommandID: "cmd-" + string(rune(i)),
			UserID:    uint64(100 + (i % 100)),
			Symbol:    "BTC-USDT",
			Side:      side,
			Price:     decimal.NewFromInt(int64(50000 + (i % 1000))),
			Quantity:  decimal.NewFromFloat(0.1),
			Filled:    decimal.Zero,
			Timestamp: int64(i),
		}
		book.Add(o)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		op := i % 3

		switch op {
		case 0:
			_ = book.BestBid()
		case 1:
			_ = book.BestAsk()
		case 2:
			_, _ = book.GetDepth(20)
		}
	}
}
