package matcher

import (
	"testing"

	"github.com/adimiuprix/spot-engine/book"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/shopspring/decimal"
)

// MockPublisher for benchmarking (minimal overhead)
type BenchPublisher struct{}

func (p *BenchPublisher) Publish(log *event.OrderBookLog) error { return nil }
func (p *BenchPublisher) Close() error                          { return nil }

// BenchmarkLimitOrder_NoMatch tests limit order placement with no matching
func BenchmarkLimitOrder_NoMatch(b *testing.B) {
	orderBook := book.NewOrderBook("BTC-USDT")
	publisher := &BenchPublisher{}
	seqGen := event.NewSequenceGenerator(0)
	m := New(orderBook, seqGen, publisher)

	// Pre-populate with opposite side orders (no crossing)
	for i := 0; i < 100; i++ {
		sellOrder := &order.Order{
			ID:        uint64(i),
			CommandID: "cmd-sell-" + string(rune(i)),
			UserID:    100,
			Symbol:    "BTC-USDT",
			Side:      order.Sell,
			Price:     decimal.NewFromInt(51000), // High price
			Quantity:  decimal.NewFromFloat(0.1),
			Filled:    decimal.Zero,
			Timestamp: int64(i),
		}
		m.ProcessWithTIF(sellOrder)
	}

	orders := make([]*order.Order, b.N)
	for i := 0; i < b.N; i++ {
		orders[i] = &order.Order{
			ID:        uint64(i + 100),
			CommandID: "cmd-buy-" + string(rune(i)),
			UserID:    200,
			Symbol:    "BTC-USDT",
			Side:      order.Buy,
			Price:     decimal.NewFromInt(49000), // Low price (no match)
			Quantity:  decimal.NewFromFloat(0.1),
			Filled:    decimal.Zero,
			Timestamp: int64(i + 100),
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.ProcessWithTIF(orders[i])
	}
}

// BenchmarkLimitOrder_FullMatch tests limit order with immediate full match
func BenchmarkLimitOrder_FullMatch(b *testing.B) {
	orderBook := book.NewOrderBook("BTC-USDT")
	publisher := &BenchPublisher{}
	seqGen := event.NewSequenceGenerator(0)
	m := New(orderBook, seqGen, publisher)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Add sell order
		sellOrder := &order.Order{
			ID:        uint64(i*2 + 1),
			CommandID: "cmd-sell-" + string(rune(i)),
			UserID:    100,
			Symbol:    "BTC-USDT",
			Side:      order.Sell,
			Price:     decimal.NewFromInt(50000),
			Quantity:  decimal.NewFromFloat(0.1),
			Filled:    decimal.Zero,
			Timestamp: int64(i * 2),
		}
		m.ProcessWithTIF(sellOrder)

		// Add matching buy order
		buyOrder := &order.Order{
			ID:        uint64(i*2 + 2),
			CommandID: "cmd-buy-" + string(rune(i)),
			UserID:    200,
			Symbol:    "BTC-USDT",
			Side:      order.Buy,
			Price:     decimal.NewFromInt(50000),
			Quantity:  decimal.NewFromFloat(0.1),
			Filled:    decimal.Zero,
			Timestamp: int64(i*2 + 1),
		}
		m.ProcessWithTIF(buyOrder)
	}
}

// BenchmarkLimitOrder_PartialMatch tests partial fill scenarios
func BenchmarkLimitOrder_PartialMatch(b *testing.B) {
	orderBook := book.NewOrderBook("BTC-USDT")
	publisher := &BenchPublisher{}
	seqGen := event.NewSequenceGenerator(0)
	m := New(orderBook, seqGen, publisher)

	// Add large sell order
	largeSell := &order.Order{
		ID:        1,
		CommandID: "cmd-sell-large",
		UserID:    100,
		Symbol:    "BTC-USDT",
		Side:      order.Sell,
		Price:     decimal.NewFromInt(50000),
		Quantity:  decimal.NewFromInt(1000), // Large quantity
		Filled:    decimal.Zero,
		Timestamp: 1,
	}
	m.ProcessWithTIF(largeSell)

	orders := make([]*order.Order, b.N)
	for i := 0; i < b.N; i++ {
		orders[i] = &order.Order{
			ID:        uint64(i + 2),
			CommandID: "cmd-buy-" + string(rune(i)),
			UserID:    200,
			Symbol:    "BTC-USDT",
			Side:      order.Buy,
			Price:     decimal.NewFromInt(50000),
			Quantity:  decimal.NewFromFloat(0.1), // Small buys
			Filled:    decimal.Zero,
			Timestamp: int64(i + 2),
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.ProcessWithTIF(orders[i])
	}
}

// BenchmarkLimitOrder_MultiLevel tests matching across multiple price levels
func BenchmarkLimitOrder_MultiLevel(b *testing.B) {
	orderBook := book.NewOrderBook("BTC-USDT")
	publisher := &BenchPublisher{}
	seqGen := event.NewSequenceGenerator(0)
	m := New(orderBook, seqGen, publisher)

	// Add sell orders at different price levels
	for i := 0; i < 10; i++ {
		sellOrder := &order.Order{
			ID:        uint64(i),
			CommandID: "cmd-sell-" + string(rune(i)),
			UserID:    100,
			Symbol:    "BTC-USDT",
			Side:      order.Sell,
			Price:     decimal.NewFromInt(int64(50000 + i*10)),
			Quantity:  decimal.NewFromFloat(1.0),
			Filled:    decimal.Zero,
			Timestamp: int64(i),
		}
		m.ProcessWithTIF(sellOrder)
	}

	orders := make([]*order.Order, b.N)
	for i := 0; i < b.N; i++ {
		orders[i] = &order.Order{
			ID:        uint64(i + 100),
			CommandID: "cmd-buy-" + string(rune(i)),
			UserID:    200,
			Symbol:    "BTC-USDT",
			Side:      order.Buy,
			Price:     decimal.NewFromInt(50100), // Crosses multiple levels
			Quantity:  decimal.NewFromFloat(5.0),
			Filled:    decimal.Zero,
			Timestamp: int64(i + 100),
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.ProcessWithTIF(orders[i])
	}
}

// BenchmarkMarketOrder tests market order execution
func BenchmarkMarketOrder(b *testing.B) {
	orderBook := book.NewOrderBook("BTC-USDT")
	publisher := &BenchPublisher{}
	seqGen := event.NewSequenceGenerator(0)
	m := New(orderBook, seqGen, publisher)

	// Pre-populate order book
	for i := 0; i < 100; i++ {
		sellOrder := &order.Order{
			ID:        uint64(i),
			CommandID: "cmd-sell-" + string(rune(i)),
			UserID:    100,
			Symbol:    "BTC-USDT",
			Side:      order.Sell,
			Price:     decimal.NewFromInt(int64(50000 + i)),
			Quantity:  decimal.NewFromFloat(1.0),
			Filled:    decimal.Zero,
			Timestamp: int64(i),
		}
		m.ProcessWithTIF(sellOrder)
	}

	orders := make([]*order.Order, b.N)
	for i := 0; i < b.N; i++ {
		orders[i] = &order.Order{
			ID:        uint64(i + 100),
			CommandID: "cmd-market-" + string(rune(i)),
			UserID:    200,
			Symbol:    "BTC-USDT",
			Side:      order.Buy,
			Type:      order.Market,
			Quantity:  decimal.NewFromFloat(0.5),
			Filled:    decimal.Zero,
			Timestamp: int64(i + 100),
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.ProcessWithTIF(orders[i])
	}
}

// BenchmarkCancelOrder tests order cancellation
func BenchmarkCancelOrder(b *testing.B) {
	orderBook := book.NewOrderBook("BTC-USDT")
	publisher := &BenchPublisher{}
	seqGen := event.NewSequenceGenerator(0)
	m := New(orderBook, seqGen, publisher)

	// Pre-add orders to cancel
	orders := make([]*order.Order, b.N)
	for i := 0; i < b.N; i++ {
		o := &order.Order{
			ID:        uint64(i),
			CommandID: "cmd-" + string(rune(i)),
			UserID:    100,
			Symbol:    "BTC-USDT",
			Side:      order.Buy,
			Price:     decimal.NewFromInt(50000),
			Quantity:  decimal.NewFromFloat(0.1),
			Filled:    decimal.Zero,
			Timestamp: int64(i),
		}
		orders[i] = o
		o.TIF = order.GTC // Set TIF for ProcessWithTIF
		m.ProcessWithTIF(o)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := &protocol.CancelOrderRequest{
			CommandID: "cancel-" + string(rune(i)),
			UserID:    100,
			Symbol:    "BTC-USDT",
			OrderID:   orders[i].OrderID,
			Timestamp: int64(i + b.N),
		}
		m.CancelOrder(req)
	}
}

// BenchmarkAmendOrder tests order amendment
func BenchmarkAmendOrder(b *testing.B) {
	orderBook := book.NewOrderBook("BTC-USDT")
	publisher := &BenchPublisher{}
	seqGen := event.NewSequenceGenerator(0)
	m := New(orderBook, seqGen, publisher)

	// Pre-add orders to amend
	orders := make([]*order.Order, b.N)
	for i := 0; i < b.N; i++ {
		o := &order.Order{
			ID:        uint64(i),
			CommandID: "cmd-" + string(rune(i)),
			UserID:    100,
			Symbol:    "BTC-USDT",
			Side:      order.Buy,
			Price:     decimal.NewFromInt(50000),
			Quantity:  decimal.NewFromFloat(1.0),
			Filled:    decimal.Zero,
			Timestamp: int64(i),
		}
		orders[i] = o
		m.ProcessWithTIF(o)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Amend size down (keeps priority)
		req := &protocol.AmendOrderRequest{
			BaseCommand: protocol.BaseCommand{
				CommandID: "amend-" + string(rune(i)),
				UserID:    100,
				MarketID:  "BTC-USDT",
				Timestamp: int64(i + b.N),
			},
			OrderID:  orders[i].OrderID,
			NewPrice: decimal.NewFromInt(50000), // Same price
			NewSize:  decimal.NewFromFloat(0.5), // Reduce size
		}
		m.ProcessAmend(req)
	}
}

// BenchmarkIcebergOrder tests iceberg order processing
func BenchmarkIcebergOrder(b *testing.B) {
	orderBook := book.NewOrderBook("BTC-USDT")
	publisher := &BenchPublisher{}
	seqGen := event.NewSequenceGenerator(0)
	m := New(orderBook, seqGen, publisher)

	// Add iceberg sell order
	iceberg := &order.Order{
		ID:        1,
		CommandID: "cmd-iceberg",
		UserID:    100,
		Symbol:    "BTC-USDT",
		Side:      order.Sell,
		Price:     decimal.NewFromInt(50000),
		Quantity:  decimal.NewFromInt(1000), // Large total
		Filled:    decimal.Zero,
		Timestamp: 1,
	}
	iceberg.SetupIceberg(decimal.NewFromFloat(1.0)) // Show 1.0 at a time
	iceberg.TIF = order.GTC
	m.ProcessWithTIF(iceberg)

	orders := make([]*order.Order, b.N)
	for i := 0; i < b.N; i++ {
		orders[i] = &order.Order{
			ID:        uint64(i + 2),
			CommandID: "cmd-buy-" + string(rune(i)),
			UserID:    200,
			Symbol:    "BTC-USDT",
			Side:      order.Buy,
			Price:     decimal.NewFromInt(50000),
			Quantity:  decimal.NewFromFloat(0.5), // Partial fills
			Filled:    decimal.Zero,
			Timestamp: int64(i + 2),
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.ProcessWithTIF(orders[i])
	}
}

// BenchmarkMatcher_RealisticWorkload simulates realistic trading patterns
func BenchmarkMatcher_RealisticWorkload(b *testing.B) {
	orderBook := book.NewOrderBook("BTC-USDT")
	publisher := &BenchPublisher{}
	seqGen := event.NewSequenceGenerator(0)
	m := New(orderBook, seqGen, publisher)

	// Pre-populate with orders
	for i := 0; i < 100; i++ {
		buyOrder := &order.Order{
			ID:        uint64(i * 2),
			CommandID: "cmd-buy-" + string(rune(i)),
			UserID:    uint64(100 + (i % 20)),
			Symbol:    "BTC-USDT",
			Side:      order.Buy,
			Price:     decimal.NewFromInt(int64(49900 - i)),
			Quantity:  decimal.NewFromFloat(0.1 * float64(1+(i%5))),
			Filled:    decimal.Zero,
			Timestamp: int64(i * 2),
		}
		m.ProcessWithTIF(buyOrder)

		sellOrder := &order.Order{
			ID:        uint64(i*2 + 1),
			CommandID: "cmd-sell-" + string(rune(i)),
			UserID:    uint64(200 + (i % 20)),
			Symbol:    "BTC-USDT",
			Side:      order.Sell,
			Price:     decimal.NewFromInt(int64(50100 + i)),
			Quantity:  decimal.NewFromFloat(0.1 * float64(1+(i%5))),
			Filled:    decimal.Zero,
			Timestamp: int64(i*2 + 1),
		}
		m.ProcessWithTIF(sellOrder)
	}

	orderID := uint64(1000)
	activeOrders := make([]*order.Order, 0, 100)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		op := i % 100

		switch {
		case op < 60: // 60% limit orders
			side := order.Buy
			basePrice := int64(49900)
			if i%2 == 0 {
				side = order.Sell
				basePrice = 50100
			}

			o := &order.Order{
				ID:        orderID,
				CommandID: "cmd-" + string(rune(int(orderID))),
				UserID:    uint64(100 + (i % 50)),
				Symbol:    "BTC-USDT",
				Side:      side,
				Price:     decimal.NewFromInt(basePrice + int64(i%100)),
				Quantity:  decimal.NewFromFloat(0.1 * float64(1+(i%10))),
				Filled:    decimal.Zero,
				Timestamp: int64(i + 1000),
			}
			m.ProcessWithTIF(o)
			activeOrders = append(activeOrders, o)
			orderID++

		case op < 75: // 15% market orders
			side := order.Buy
			if i%2 == 0 {
				side = order.Sell
			}

			o := &order.Order{
				ID:        orderID,
				OrderID:   "ORDER-" + string(rune(int(orderID))),
				CommandID: "cmd-market-" + string(rune(int(orderID))),
				UserID:    uint64(100 + (i % 50)),
				Symbol:    "BTC-USDT",
				Side:      side,
				Type:      order.Market,
				Quantity:  decimal.NewFromFloat(0.05),
				Filled:    decimal.Zero,
				Timestamp: int64(i + 1000),
			}
			m.ProcessWithTIF(o)
			orderID++

		case op < 90: // 15% cancels
			if len(activeOrders) > 0 {
				idx := i % len(activeOrders)
				o := activeOrders[idx]
				req := &protocol.CancelOrderRequest{
					CommandID: "cancel-" + string(rune(int(orderID))),
					UserID:    o.UserID,
					Symbol:    "BTC-USDT",
					OrderID:   o.OrderID,
					Timestamp: int64(i + 1000),
				}
				m.CancelOrder(req)
				// Remove from tracking
				activeOrders = append(activeOrders[:idx], activeOrders[idx+1:]...)
			}

		default: // 10% amends
			if len(activeOrders) > 0 {
				idx := i % len(activeOrders)
				o := activeOrders[idx]
				newQty := o.Quantity.Mul(decimal.NewFromFloat(0.8)) // Reduce by 20%
				req := &protocol.AmendOrderRequest{
					BaseCommand: protocol.BaseCommand{
						CommandID: "amend-" + string(rune(int(orderID))),
						UserID:    o.UserID,
						MarketID:  "BTC-USDT",
						Timestamp: int64(i + 1000),
					},
					OrderID:  o.OrderID,
					NewPrice: o.Price,
					NewSize:  newQty,
				}
				m.ProcessAmend(req)
			}
		}
	}
}
