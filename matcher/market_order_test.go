package matcher

import (
	"testing"

	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

// TestMarketBuyFullFill tests market buy with sufficient liquidity
func TestMarketBuyFullFill(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Add sell orders
	sell := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1000,
	)
	matcher.book.Add(&sell)

	// Create market buy
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Market, order.GTC,
		decimal.Zero, // Market orders have no price
		decimal.NewFromInt(10),
		1001,
	)

	result := matcher.Process(&buy)

	// Should execute
	if len(result.Trades) == 0 {
		t.Fatal("Expected market order to execute")
	}

	// Buy should be filled
	if !buy.Filled.Equal(decimal.NewFromInt(10)) {
		t.Errorf("Expected buy filled 10, got %v", buy.Filled)
	}

	// Sell should be filled
	if !sell.IsFilled() {
		t.Error("Expected sell to be filled")
	}
}

// TestMarketBuyMultipleLevels tests market buy walking through levels
func TestMarketBuyMultipleLevels(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Add sell orders at different prices
	sell1 := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(5),
		1000,
	)
	sell2 := order.NewOrder(
		2, "SELL-2", "CMD-S2", 102, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(101),
		decimal.NewFromInt(5),
		1001,
	)
	matcher.book.Add(&sell1)
	matcher.book.Add(&sell2)

	// Create market buy for total quantity
	buy := order.NewOrder(
		3, "BUY-1", "CMD-B1", 103, "BTCUSD",
		order.Buy, order.Market, order.GTC,
		decimal.Zero,
		decimal.NewFromInt(10), // Total from both levels
		1002,
	)

	result := matcher.Process(&buy)

	// Should have 2 trades (walks through 2 levels)
	// Each trade: 1 trade log + 2 fill logs = 3 logs per trade
	if len(result.Trades) < 6 {
		t.Fatalf("Expected at least 6 logs (2 trades), got %d", len(result.Trades))
	}

	// Buy should be fully filled
	if !buy.Filled.Equal(decimal.NewFromInt(10)) {
		t.Errorf("Expected buy filled 10, got %v", buy.Filled)
	}

	// Both sells should be filled
	if !sell1.IsFilled() {
		t.Error("Expected sell1 to be filled")
	}
	if !sell2.IsFilled() {
		t.Error("Expected sell2 to be filled")
	}
}

// TestMarketBuyEmptyBook tests market buy with no liquidity
func TestMarketBuyEmptyBook(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Create market buy with empty book
	buy := order.NewOrder(
		1, "BUY-1", "CMD-B1", 101, "BTCUSD",
		order.Buy, order.Market, order.GTC,
		decimal.Zero,
		decimal.NewFromInt(10),
		1000,
	)

	result := matcher.Process(&buy)

	// Should not execute (no liquidity)
	// Base mode market order without liquidity just doesn't match
	// No reject log for base mode
	if len(result.Trades) != 0 {
		t.Errorf("Expected no trades with empty book, got %d", len(result.Trades))
	}

	// Order should not be filled
	if buy.Filled.GreaterThan(decimal.Zero) {
		t.Error("Expected order not to be filled")
	}
}

// TestMarketBuyPartialFill tests market buy with insufficient liquidity
func TestMarketBuyPartialFill(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Add small sell order
	sell := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(5), // Only 5 available
		1000,
	)
	matcher.book.Add(&sell)

	// Create market buy for more
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Market, order.GTC,
		decimal.Zero,
		decimal.NewFromInt(10), // Want 10
		1001,
	)

	result := matcher.Process(&buy)

	// Should execute partially
	if len(result.Trades) == 0 {
		t.Fatal("Expected partial execution")
	}

	// Buy should be partially filled
	if !buy.Filled.Equal(decimal.NewFromInt(5)) {
		t.Errorf("Expected buy filled 5, got %v", buy.Filled)
	}

	// Base mode market order: no reject for remaining
	// Just executes what it can and stops
}

// TestMarketSellFullFill tests market sell with sufficient liquidity
func TestMarketSellFullFill(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Add buy order
	buy := order.NewOrder(
		1, "BUY-1", "CMD-B1", 101, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1000,
	)
	matcher.book.Add(&buy)

	// Create market sell
	sell := order.NewOrder(
		2, "SELL-1", "CMD-S1", 102, "BTCUSD",
		order.Sell, order.Market, order.GTC,
		decimal.Zero,
		decimal.NewFromInt(10),
		1001,
	)

	result := matcher.Process(&sell)

	// Should execute
	if len(result.Trades) == 0 {
		t.Fatal("Expected market order to execute")
	}

	// Sell should be filled
	if !sell.Filled.Equal(decimal.NewFromInt(10)) {
		t.Errorf("Expected sell filled 10, got %v", sell.Filled)
	}

	// Buy should be filled
	if !buy.IsFilled() {
		t.Error("Expected buy to be filled")
	}
}

// TestMarketSellMultipleLevels tests market sell walking through levels
func TestMarketSellMultipleLevels(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Add buy orders at different prices
	buy1 := order.NewOrder(
		1, "BUY-1", "CMD-B1", 101, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(5),
		1000,
	)
	buy2 := order.NewOrder(
		2, "BUY-2", "CMD-B2", 102, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(99),
		decimal.NewFromInt(5),
		1001,
	)
	matcher.book.Add(&buy1)
	matcher.book.Add(&buy2)

	// Create market sell
	sell := order.NewOrder(
		3, "SELL-1", "CMD-S1", 103, "BTCUSD",
		order.Sell, order.Market, order.GTC,
		decimal.Zero,
		decimal.NewFromInt(10),
		1002,
	)

	result := matcher.Process(&sell)

	// Should have 2 trades
	if len(result.Trades) < 6 {
		t.Fatalf("Expected at least 6 logs (2 trades), got %d", len(result.Trades))
	}

	// Sell should be fully filled
	if !sell.Filled.Equal(decimal.NewFromInt(10)) {
		t.Errorf("Expected sell filled 10, got %v", sell.Filled)
	}

	// Both buys should be filled
	if !buy1.IsFilled() {
		t.Error("Expected buy1 to be filled")
	}
	if !buy2.IsFilled() {
		t.Error("Expected buy2 to be filled")
	}
}

// TestMarketBuyQuoteMode tests market buy with quote size (budget)
func TestMarketBuyQuoteMode(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Add sell order at price 100
	sell := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100), // Price 100
		decimal.NewFromInt(10),  // 10 units available
		1000,
	)
	matcher.book.Add(&sell)

	// Create market buy with quote size (spend $500)
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Market, order.GTC,
		decimal.Zero,
		decimal.Zero, // No base quantity
		1001,
	)
	buy.QuoteSize = decimal.NewFromInt(500) // Spend $500

	result := matcher.Process(&buy)

	// Should execute
	if len(result.Trades) == 0 {
		t.Fatal("Expected market order to execute")
	}

	// Should buy 5 units (500 / 100 = 5)
	expectedFilled := decimal.NewFromInt(5)
	if !buy.Filled.Equal(expectedFilled) {
		t.Errorf("Expected buy filled %v, got %v", expectedFilled, buy.Filled)
	}
}

// TestMarketSellQuoteMode tests market sell with quote size (target revenue)
func TestMarketSellQuoteMode(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Add buy order at price 100
	buy := order.NewOrder(
		1, "BUY-1", "CMD-B1", 101, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1000,
	)
	matcher.book.Add(&buy)

	// Create market sell with quote size (receive $300)
	sell := order.NewOrder(
		2, "SELL-1", "CMD-S1", 102, "BTCUSD",
		order.Sell, order.Market, order.GTC,
		decimal.Zero,
		decimal.Zero, // No base quantity
		1001,
	)
	sell.QuoteSize = decimal.NewFromInt(300) // Receive $300

	result := matcher.Process(&sell)

	// Should execute
	if len(result.Trades) == 0 {
		t.Fatal("Expected market order to execute")
	}

	// Should sell 3 units (300 / 100 = 3)
	expectedFilled := decimal.NewFromInt(3)
	if !sell.Filled.Equal(expectedFilled) {
		t.Errorf("Expected sell filled %v, got %v", expectedFilled, sell.Filled)
	}
}

// TestMarketOrderBelowLotSize tests market order with remainder below lot size
func TestMarketOrderBelowLotSize(t *testing.T) {
	matcher, pub := newTestMatcher()

	// Set lot size
	matcher.SetLotSize(decimal.NewFromFloat(0.1))

	// Add sell order
	sell := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromFloat(0.5), // 0.5 available
		1000,
	)
	matcher.book.Add(&sell)

	// Create market buy for slightly more (will have tiny remainder)
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Market, order.GTC,
		decimal.Zero,
		decimal.Zero,
		1001,
	)
	buy.QuoteSize = decimal.NewFromInt(60) // Want to spend $60 at price 100

	result := matcher.Process(&buy)

	// Should execute what it can
	if len(result.Trades) == 0 {
		t.Fatal("Expected execution")
	}

	// Should have reject log for remainder below lot size
	logs := pub.GetLogs()
	hasReject := false
	for _, log := range logs {
		if log.LogType == event.LogTypeReject {
			hasReject = true
			break
		}
	}
	if !hasReject {
		t.Error("Expected reject log for remainder below lot size")
	}
}

// TestMarketOrderExecutionPrice tests that market orders use maker price
func TestMarketOrderExecutionPrice(t *testing.T) {
	matcher, pub := newTestMatcher()

	// Add sell orders at different prices
	sell1 := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100), // Best price
		decimal.NewFromInt(5),
		1000,
	)
	sell2 := order.NewOrder(
		2, "SELL-2", "CMD-S2", 102, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(105), // Worse price
		decimal.NewFromInt(5),
		1001,
	)
	matcher.book.Add(&sell1)
	matcher.book.Add(&sell2)

	// Market buy
	buy := order.NewOrder(
		3, "BUY-1", "CMD-B1", 103, "BTCUSD",
		order.Buy, order.Market, order.GTC,
		decimal.Zero,
		decimal.NewFromInt(10), // Will hit both levels
		1002,
	)

	matcher.Process(&buy)

	// Check trade prices
	logs := pub.GetLogs()
	trade1Found := false
	trade2Found := false

	for _, log := range logs {
		if log.LogType == event.LogTypeTrade {
			if log.TradePrice.Equal(decimal.NewFromInt(100)) {
				trade1Found = true // First trade at 100
			}
			if log.TradePrice.Equal(decimal.NewFromInt(105)) {
				trade2Found = true // Second trade at 105
			}
		}
	}

	if !trade1Found {
		t.Error("Expected trade at price 100")
	}
	if !trade2Found {
		t.Error("Expected trade at price 105")
	}
}
