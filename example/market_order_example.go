package main

import (
	"fmt"
	"time"

	"github.com/adimiuprix/spot-engine/book"
	"github.com/adimiuprix/spot-engine/engine"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("=== Market Order Example ===\n")

	// Create engine with LotSize configuration
	config := engine.Config{
		Symbol:         "BTCUSD",
		RingBufferSize: 1024,
		MinLotSize:     decimal.NewFromFloat(0.001), // 0.001 BTC minimum
	}
	eng := engine.New(config)
	eng.Start()
	defer eng.Stop()

	// Monitor events
	go monitorEvents(eng.Events())

	// Wait for engine to start
	time.Sleep(100 * time.Millisecond)

	// ========================================
	// Setup: Place resting limit orders
	// ========================================
	fmt.Println("--- Setup: Placing resting limit orders ---")

	// Sell orders (asks) at different prices
	sellOrders := []*order.Order{
		createLimitOrder("SELL-1", 101, 1, "BTCUSD", order.Sell, 50000, 0.5),
		createLimitOrder("SELL-2", 102, 2, "BTCUSD", order.Sell, 50100, 1.0),
		createLimitOrder("SELL-3", 103, 3, "BTCUSD", order.Sell, 50200, 1.5),
		createLimitOrder("SELL-4", 104, 4, "BTCUSD", order.Sell, 50300, 2.0),
	}

	// Buy orders (bids) at different prices
	buyOrders := []*order.Order{
		createLimitOrder("BUY-1", 105, 5, "BTCUSD", order.Buy, 49900, 0.5),
		createLimitOrder("BUY-2", 106, 6, "BTCUSD", order.Buy, 49800, 1.0),
		createLimitOrder("BUY-3", 107, 7, "BTCUSD", order.Buy, 49700, 1.5),
		createLimitOrder("BUY-4", 108, 8, "BTCUSD", order.Buy, 49600, 2.0),
	}

	for _, o := range sellOrders {
		eng.SubmitOrder(o)
	}
	for _, o := range buyOrders {
		eng.SubmitOrder(o)
	}

	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n--- Initial Orderbook ---")
	printOrderBook(eng)

	// ========================================
	// Test 1: Market Buy - Base Mode (Size-based)
	// ========================================
	fmt.Println("\n\n=== Test 1: Market Buy - Base Mode ===")
	fmt.Println("Buy 1.5 BTC at market price")
	fmt.Println("Expected: Match 0.5@50000 + 1.0@50100")

	marketBuyBase := createMarketOrder("MKT-BUY-1", 201, 101, "BTCUSD", order.Buy, 1.5, 0)
	eng.SubmitOrder(marketBuyBase)

	time.Sleep(200 * time.Millisecond)

	// ========================================
	// Test 2: Market Buy - Quote Mode (Budget-based)
	// ========================================
	fmt.Println("\n\n=== Test 2: Market Buy - Quote Mode ===")
	fmt.Println("Spend 75,000 USDT to buy BTC")
	fmt.Println("Expected: Calculate qty from budget, match multiple levels")

	marketBuyQuote := createMarketOrderQuote("MKT-BUY-2", 202, 102, "BTCUSD", order.Buy, 75000)
	eng.SubmitOrder(marketBuyQuote)

	time.Sleep(200 * time.Millisecond)

	// ========================================
	// Test 3: Market Sell - Base Mode
	// ========================================
	fmt.Println("\n\n=== Test 3: Market Sell - Base Mode ===")
	fmt.Println("Sell 1.5 BTC at market price")
	fmt.Println("Expected: Match 0.5@49900 + 1.0@49800")

	marketSellBase := createMarketOrder("MKT-SELL-1", 203, 103, "BTCUSD", order.Sell, 1.5, 0)
	eng.SubmitOrder(marketSellBase)

	time.Sleep(200 * time.Millisecond)

	// ========================================
	// Test 4: Market Sell - Quote Mode
	// ========================================
	fmt.Println("\n\n=== Test 4: Market Sell - Quote Mode ===")
	fmt.Println("Sell BTC to receive 75,000 USDT")
	fmt.Println("Expected: Calculate qty needed, match multiple levels")

	marketSellQuote := createMarketOrderQuote("MKT-SELL-2", 204, 104, "BTCUSD", order.Sell, 75000)
	eng.SubmitOrder(marketSellQuote)

	time.Sleep(200 * time.Millisecond)

	// ========================================
	// Test 5: Market Order - No Liquidity
	// ========================================
	fmt.Println("\n\n=== Test 5: Market Order - No Liquidity ===")
	fmt.Println("Try to buy 100 BTC (more than available)")
	fmt.Println("Expected: Match what's available, reject remaining")

	marketBuyHuge := createMarketOrder("MKT-BUY-3", 205, 105, "BTCUSD", order.Buy, 100, 0)
	eng.SubmitOrder(marketBuyHuge)

	time.Sleep(200 * time.Millisecond)

	// ========================================
	// Test 6: Market Order - LotSize Protection
	// ========================================
	fmt.Println("\n\n=== Test 6: Market Order - LotSize Protection ===")
	fmt.Println("Spend 0.00001 USDT (micro amount)")
	fmt.Println("Expected: Reject due to below MinLotSize")

	// Add one more ask for this test
	microAsk := createLimitOrder("SELL-MICRO", 301, 301, "BTCUSD", order.Sell, 50000, 1.0)
	eng.SubmitOrder(microAsk)
	time.Sleep(100 * time.Millisecond)

	marketBuyMicro := createMarketOrderQuote("MKT-BUY-MICRO", 206, 106, "BTCUSD", order.Buy, 0.00001)
	eng.SubmitOrder(marketBuyMicro)

	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n\n--- Final Orderbook ---")
	printOrderBook(eng)

	// Wait for all events
	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n=== Market Order Example Complete ===")
}

// Helper: Create limit order
func createLimitOrder(orderID string, id uint64, userID uint64, symbol string, side order.Side, price float64, qty float64) *order.Order {
	return &order.Order{
		ID:        id,
		OrderID:   orderID,
		CommandID: fmt.Sprintf("cmd-%d", id),
		UserID:    userID,
		Symbol:    symbol,
		Side:      side,
		Type:      order.Limit,
		TIF:       order.GTC,
		Price:     decimal.NewFromFloat(price),
		Quantity:  decimal.NewFromFloat(qty),
		Filled:    decimal.Zero,
		Timestamp: time.Now().UnixNano(),
	}
}

// Helper: Create market order (base mode - size-based)
func createMarketOrder(orderID string, id uint64, userID uint64, symbol string, side order.Side, qty float64, price float64) *order.Order {
	return &order.Order{
		ID:        id,
		OrderID:   orderID,
		CommandID: fmt.Sprintf("cmd-%d", id),
		UserID:    userID,
		Symbol:    symbol,
		Side:      side,
		Type:      order.Market,
		TIF:       order.GTC,                   // TIF ignored for market orders
		Price:     decimal.NewFromFloat(price), // Price ignored for market orders
		Quantity:  decimal.NewFromFloat(qty),
		QuoteSize: decimal.Zero, // Base mode
		Filled:    decimal.Zero,
		Timestamp: time.Now().UnixNano(),
	}
}

// Helper: Create market order (quote mode - budget-based)
func createMarketOrderQuote(orderID string, id uint64, userID uint64, symbol string, side order.Side, quoteAmount float64) *order.Order {
	return &order.Order{
		ID:        id,
		OrderID:   orderID,
		CommandID: fmt.Sprintf("cmd-%d", id),
		UserID:    userID,
		Symbol:    symbol,
		Side:      side,
		Type:      order.Market,
		TIF:       order.GTC,    // TIF ignored for market orders
		Price:     decimal.Zero, // Price ignored for market orders
		Quantity:  decimal.Zero, // Quote mode - quantity calculated
		QuoteSize: decimal.NewFromFloat(quoteAmount),
		Filled:    decimal.Zero,
		Timestamp: time.Now().UnixNano(),
	}
}

func monitorEvents(events <-chan *event.OrderBookLog) {
	for log := range events {
		switch log.LogType {
		case event.LogTypeTrade:
			fmt.Printf("  [TRADE] %s BTC @ %s (taker: %s, maker: %s)\n",
				log.TradeQuantity.String(),
				log.TradePrice.String(),
				log.TakerOrderID,
				log.MakerOrderID,
			)

		case event.LogTypeCancel:
			fmt.Printf("  [CANCEL] Order %s: %s BTC @ %s\n",
				log.OrderID,
				log.RemainingSize.String(),
				log.Price.String(),
			)

		case event.LogTypeReject:
			fmt.Printf("  [REJECT] Order %s: %s - %s\n",
				log.OrderID,
				log.RejectReason,
				log.RejectDetail,
			)

		case event.LogTypeFill:
			status := "partial"
			if log.EventType == event.EventTypeFill {
				status = "full"
			}
			fmt.Printf("  [FILL] %s: Order %s filled %s BTC, remaining %s (%s)\n",
				log.Side,
				log.OrderID,
				log.FilledSize.String(),
				log.RemainingSize.String(),
				status,
			)
		}
	}
}

func printOrderBook(eng *engine.Engine) {
	bk := eng.GetOrderBook()

	fmt.Println("Asks (Sell Orders):")
	count := 0
	bk.AskTree.Ascend(func(level *book.PriceLevel) bool {
		if level.Volume.GreaterThan(decimal.Zero) && count < 5 {
			fmt.Printf("  %s: %s BTC (%d orders)\n",
				level.Price.String(),
				level.Volume.String(),
				level.OrderCount,
			)
			count++
		}
		return count < 5
	})

	fmt.Println("Bids (Buy Orders):")
	count = 0
	bk.BidTree.Ascend(func(level *book.PriceLevel) bool {
		if level.Volume.GreaterThan(decimal.Zero) && count < 5 {
			fmt.Printf("  %s: %s BTC (%d orders)\n",
				level.Price.String(),
				level.Volume.String(),
				level.OrderCount,
			)
			count++
		}
		return count < 5
	})
}
