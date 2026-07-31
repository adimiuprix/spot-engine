package matcher

import (
	"testing"

	"github.com/adimiuprix/spot-engine/book"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

// MockPublisher is a simple publisher for testing
type MockPublisher struct {
	logs []*event.OrderBookLog
}

func (m *MockPublisher) Publish(log *event.OrderBookLog) error {
	m.logs = append(m.logs, log)
	return nil
}

func (m *MockPublisher) Close() error {
	return nil
}

func (m *MockPublisher) GetLogs() []*event.OrderBookLog {
	return m.logs
}

func (m *MockPublisher) Clear() {
	m.logs = nil
}

// Helper function to create a test matcher
func newTestMatcher() (*Matcher, *MockPublisher) {
	ob := book.NewOrderBook("BTCUSD")
	seqGen := event.NewSequenceGenerator(0) // Start from 0
	pub := &MockPublisher{}
	matcher := New(ob, seqGen, pub)
	return matcher, pub
}

// TestNew tests matcher creation
func TestNew(t *testing.T) {
	matcher, _ := newTestMatcher()

	if matcher == nil {
		t.Fatal("Expected non-nil matcher")
	}

	if matcher.book == nil {
		t.Error("Expected non-nil order book")
	}

	if matcher.tradeID != 1 {
		t.Errorf("Expected initial tradeID 1, got %d", matcher.tradeID)
	}

	// Check default lot size
	expectedLotSize := decimal.NewFromFloat(0.00000001)
	if !matcher.lotSize.Equal(expectedLotSize) {
		t.Errorf("Expected lot size %v, got %v", expectedLotSize, matcher.lotSize)
	}
}

// TestSetLotSize tests setting lot size
func TestSetLotSize(t *testing.T) {
	matcher, _ := newTestMatcher()

	newLotSize := decimal.NewFromFloat(0.001)
	matcher.SetLotSize(newLotSize)

	if !matcher.lotSize.Equal(newLotSize) {
		t.Errorf("Expected lot size %v, got %v", newLotSize, matcher.lotSize)
	}
}

// TestGetSetTradeID tests trade ID management
func TestGetSetTradeID(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Initial trade ID
	if matcher.GetTradeID() != 1 {
		t.Errorf("Expected initial tradeID 1, got %d", matcher.GetTradeID())
	}

	// Set new trade ID
	matcher.SetTradeID(100)
	if matcher.GetTradeID() != 100 {
		t.Errorf("Expected tradeID 100, got %d", matcher.GetTradeID())
	}
}

// TestMatchLimitBuyFullFill tests full fill of a limit buy order
func TestMatchLimitBuyFullFill(t *testing.T) {
	matcher, pub := newTestMatcher()

	// Add a sell order to the book
	sell := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100), // Price
		decimal.NewFromInt(10),  // Quantity
		1000,
	)
	matcher.book.Add(&sell)

	// Create a buy order that will fully match
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100), // Same price
		decimal.NewFromInt(10),  // Same quantity
		1001,
	)

	// Process the buy order
	result := matcher.Process(&buy)

	// Should have trade logs (1 trade + 2 fills = 3 logs)
	if len(result.Trades) != 3 {
		t.Fatalf("Expected 3 logs (trade + 2 fills), got %d", len(result.Trades))
	}

	// Verify trade log
	tradeLog := result.Trades[0]
	if tradeLog.LogType != event.LogTypeTrade {
		t.Errorf("Expected trade log type, got %v", tradeLog.LogType)
	}

	// Both orders should be fully filled
	if !buy.IsFilled() {
		t.Error("Expected buy order to be fully filled")
	}
	if !sell.IsFilled() {
		t.Error("Expected sell order to be fully filled")
	}

	// Trade ID should increment
	if matcher.GetTradeID() != 2 {
		t.Errorf("Expected tradeID 2 after trade, got %d", matcher.GetTradeID())
	}

	// Publisher should have received logs
	if len(pub.GetLogs()) != 3 {
		t.Errorf("Expected 3 published logs, got %d", len(pub.GetLogs()))
	}
}

// TestMatchLimitBuyPartialFill tests partial fill of a limit buy order
func TestMatchLimitBuyPartialFill(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Add a small sell order
	sell := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(5), // Only 5 available
		1000,
	)
	matcher.book.Add(&sell)

	// Create a buy order for more quantity
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10), // Want 10
		1001,
	)

	result := matcher.Process(&buy)

	// Should have trade logs
	if len(result.Trades) != 3 {
		t.Fatalf("Expected 3 logs, got %d", len(result.Trades))
	}

	// Sell should be fully filled
	if !sell.IsFilled() {
		t.Error("Expected sell order to be fully filled")
	}

	// Buy should be partially filled (5 out of 10)
	if buy.IsFilled() {
		t.Error("Expected buy order to be partially filled, not fully")
	}
	if !buy.Filled.Equal(decimal.NewFromInt(5)) {
		t.Errorf("Expected buy filled 5, got %v", buy.Filled)
	}
	if !buy.Remaining().Equal(decimal.NewFromInt(5)) {
		t.Errorf("Expected buy remaining 5, got %v", buy.Remaining())
	}
}

// TestMatchLimitSellFullFill tests full fill of a limit sell order
func TestMatchLimitSellFullFill(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Add a buy order to the book
	buy := order.NewOrder(
		1, "BUY-1", "CMD-B1", 101, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1000,
	)
	matcher.book.Add(&buy)

	// Create a sell order
	sell := order.NewOrder(
		2, "SELL-1", "CMD-S1", 102, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1001,
	)

	result := matcher.Process(&sell)

	// Should have trade logs
	if len(result.Trades) != 3 {
		t.Fatalf("Expected 3 logs, got %d", len(result.Trades))
	}

	// Both orders should be fully filled
	if !buy.IsFilled() {
		t.Error("Expected buy order to be fully filled")
	}
	if !sell.IsFilled() {
		t.Error("Expected sell order to be fully filled")
	}
}

// TestMatchLimitBuyNoCross tests buy order that doesn't cross spread
func TestMatchLimitBuyNoCross(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Add a sell order at 105
	sell := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(105),
		decimal.NewFromInt(10),
		1000,
	)
	matcher.book.Add(&sell)

	// Create a buy order at 100 (below ask)
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1001,
	)

	result := matcher.Process(&buy)

	// Should not match - no trades
	if len(result.Trades) != 0 {
		t.Errorf("Expected no trades, got %d", len(result.Trades))
	}

	// Buy order should not be filled
	if buy.IsFilled() {
		t.Error("Expected buy order not to be filled")
	}
}

// TestMatchLimitSellNoCross tests sell order that doesn't cross spread
func TestMatchLimitSellNoCross(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Add a buy order at 95
	buy := order.NewOrder(
		1, "BUY-1", "CMD-B1", 101, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(95),
		decimal.NewFromInt(10),
		1000,
	)
	matcher.book.Add(&buy)

	// Create a sell order at 100 (above bid)
	sell := order.NewOrder(
		2, "SELL-1", "CMD-S1", 102, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1001,
	)

	result := matcher.Process(&sell)

	// Should not match
	if len(result.Trades) != 0 {
		t.Errorf("Expected no trades, got %d", len(result.Trades))
	}
}

// TestMatchLimitBuyMultipleLevels tests buy matching across multiple price levels
func TestMatchLimitBuyMultipleLevels(t *testing.T) {
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

	// Create a buy order that crosses both levels
	buy := order.NewOrder(
		3, "BUY-1", "CMD-B1", 103, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(101), // Price covers both levels
		decimal.NewFromInt(10),  // Quantity for both
		1002,
	)

	result := matcher.Process(&buy)

	// Should have 2 trades (2 * 3 logs = 6 logs)
	if len(result.Trades) != 6 {
		t.Fatalf("Expected 6 logs (2 trades), got %d", len(result.Trades))
	}

	// Buy should be fully filled
	if !buy.IsFilled() {
		t.Error("Expected buy order to be fully filled")
	}

	// Both sells should be fully filled
	if !sell1.IsFilled() {
		t.Error("Expected sell1 to be fully filled")
	}
	if !sell2.IsFilled() {
		t.Error("Expected sell2 to be fully filled")
	}
}

// TestMatchPricePriority tests that best price is matched first
func TestMatchPricePriority(t *testing.T) {
	matcher, pub := newTestMatcher()

	// Add sell orders at different prices
	sell1 := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(101), // Higher price
		decimal.NewFromInt(10),
		1000,
	)
	sell2 := order.NewOrder(
		2, "SELL-2", "CMD-S2", 102, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100), // Lower price (best)
		decimal.NewFromInt(10),
		1001,
	)
	matcher.book.Add(&sell1)
	matcher.book.Add(&sell2)

	// Create a buy order
	buy := order.NewOrder(
		3, "BUY-1", "CMD-B1", 103, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(101),
		decimal.NewFromInt(5), // Only partially fill
		1002,
	)

	matcher.Process(&buy)

	// Should match with sell2 (lower price) first
	tradeLog := pub.GetLogs()[0]
	if tradeLog.LogType != event.LogTypeTrade {
		t.Fatal("Expected trade log")
	}

	// Maker should be SELL-2 (best price)
	if tradeLog.MakerOrderID != "SELL-2" {
		t.Errorf("Expected maker to be SELL-2, got %s", tradeLog.MakerOrderID)
	}

	// Price should be 100 (maker price)
	if !tradeLog.TradePrice.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Expected trade price 100, got %v", tradeLog.TradePrice)
	}
}

// TestMatchTimePriority tests FIFO at same price level
func TestMatchTimePriority(t *testing.T) {
	matcher, pub := newTestMatcher()

	// Add sell orders at same price
	sell1 := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1000, // Earlier timestamp
	)
	sell2 := order.NewOrder(
		2, "SELL-2", "CMD-S2", 102, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1001, // Later timestamp
	)
	matcher.book.Add(&sell1)
	matcher.book.Add(&sell2)

	// Create a buy order
	buy := order.NewOrder(
		3, "BUY-1", "CMD-B1", 103, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(5), // Partial fill
		1002,
	)

	matcher.Process(&buy)

	// Should match with sell1 first (FIFO)
	tradeLog := pub.GetLogs()[0]
	if tradeLog.MakerOrderID != "SELL-1" {
		t.Errorf("Expected maker to be SELL-1 (FIFO), got %s", tradeLog.MakerOrderID)
	}

	// Sell1 should be partially filled
	if !sell1.Filled.Equal(decimal.NewFromInt(5)) {
		t.Errorf("Expected sell1 filled 5, got %v", sell1.Filled)
	}

	// Sell2 should not be touched
	if sell2.Filled.GreaterThan(decimal.Zero) {
		t.Error("Expected sell2 not to be filled (FIFO priority)")
	}
}

// TestMatchEmptyBook tests matching against empty book
func TestMatchEmptyBook(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Create a buy order with empty book
	buy := order.NewOrder(
		1, "BUY-1", "CMD-B1", 101, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1000,
	)

	result := matcher.Process(&buy)

	// Should not match
	if len(result.Trades) != 0 {
		t.Errorf("Expected no trades with empty book, got %d", len(result.Trades))
	}

	// Order should not be filled
	if buy.Filled.GreaterThan(decimal.Zero) {
		t.Error("Expected order not to be filled")
	}
}

// TestExecuteTradeDetails tests trade execution details
func TestExecuteTradeDetails(t *testing.T) {
	matcher, pub := newTestMatcher()

	// Add sell order
	sell := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1000,
	)
	matcher.book.Add(&sell)

	// Add buy order
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1001,
	)

	matcher.Process(&buy)

	logs := pub.GetLogs()

	// Verify trade log details
	tradeLog := logs[0]
	if tradeLog.LogType != event.LogTypeTrade {
		t.Error("Expected first log to be trade")
	}
	if !tradeLog.TradePrice.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Expected trade price 100, got %v", tradeLog.TradePrice)
	}
	if !tradeLog.TradeQuantity.Equal(decimal.NewFromInt(10)) {
		t.Errorf("Expected trade quantity 10, got %v", tradeLog.TradeQuantity)
	}
	if tradeLog.Side != "buy" {
		t.Errorf("Expected taker side 'buy', got %s", tradeLog.Side)
	}

	// Verify fill logs
	makerFillLog := logs[1]
	if makerFillLog.LogType != event.LogTypeFill {
		t.Error("Expected second log to be fill (maker)")
	}

	takerFillLog := logs[2]
	if takerFillLog.LogType != event.LogTypeFill {
		t.Error("Expected third log to be fill (taker)")
	}
}
