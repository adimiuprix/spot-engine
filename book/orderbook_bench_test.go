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
