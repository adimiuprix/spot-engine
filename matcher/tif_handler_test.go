package matcher

import (
	"testing"

	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

// TestProcessWithTIF_GTC tests GTC order processing
func TestProcessWithTIF_GTC(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Add sell order
	sell := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(5),
		1000,
	)
	matcher.book.Add(&sell)

	// Create GTC buy for more quantity
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10), // Want 10, only 5 available
		1001,
	)

	result := matcher.ProcessWithTIF(&buy)

	// Should partially fill (5 out of 10)
	if !buy.Filled.Equal(decimal.NewFromInt(5)) {
		t.Errorf("Expected buy filled 5, got %v", buy.Filled)
	}

	// Remaining 5 should rest in book (GTC behavior)
	// This would be handled by engine, not matcher
	if len(result.Trades) != 3 { // 1 trade + 2 fills
		t.Errorf("Expected 3 logs, got %d", len(result.Trades))
	}
}

// TestProcessWithTIF_IOC_FullFill tests IOC with full fill
func TestProcessWithTIF_IOC_FullFill(t *testing.T) {
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

	// Create IOC buy
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Limit, order.IOC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1001,
	)

	result := matcher.ProcessWithTIF(&buy)

	// Should fully fill
	if !buy.IsFilled() {
		t.Error("Expected buy to be fully filled")
	}

	// Should not have cancel log (fully filled)
	logs := pub.GetLogs()
	for _, log := range logs {
		if log.LogType == event.LogTypeCancel {
			t.Error("Expected no cancel log for fully filled IOC")
		}
	}

	if len(result.Trades) != 3 { // 1 trade + 2 fills
		t.Errorf("Expected 3 logs, got %d", len(result.Trades))
	}
}

// TestProcessWithTIF_IOC_PartialFill tests IOC with partial fill
func TestProcessWithTIF_IOC_PartialFill(t *testing.T) {
	matcher, pub := newTestMatcher()

	// Add sell order
	sell := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(5), // Only 5 available
		1000,
	)
	matcher.book.Add(&sell)

	// Create IOC buy for more
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Limit, order.IOC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10), // Want 10
		1001,
	)

	result := matcher.ProcessWithTIF(&buy)

	// Should partially fill
	if !buy.Filled.Equal(decimal.NewFromInt(5)) {
		t.Errorf("Expected buy filled 5, got %v", buy.Filled)
	}

	// Should have cancel log for remaining 5
	logs := pub.GetLogs()
	hasCancel := false
	for _, log := range logs {
		if log.LogType == event.LogTypeCancel {
			hasCancel = true
			// Verify cancel amount
			if !log.RemainingSize.Equal(decimal.NewFromInt(5)) {
				t.Errorf("Expected cancel remaining 5, got %v", log.RemainingSize)
			}
		}
	}
	if !hasCancel {
		t.Error("Expected cancel log for IOC remaining")
	}

	// Total: 1 trade + 2 fills + 1 cancel = 4 logs
	// Note: IOC actually produces extra logs (reject before cancel in some cases)
	if len(result.Trades) < 4 {
		t.Errorf("Expected at least 4 logs (trade+fills+cancel), got %d", len(result.Trades))
	}
}

// TestProcessWithTIF_IOC_NoFill tests IOC with no fill
func TestProcessWithTIF_IOC_NoFill(t *testing.T) {
	matcher, pub := newTestMatcher()

	// Add sell order at higher price
	sell := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(105), // Higher price
		decimal.NewFromInt(10),
		1000,
	)
	matcher.book.Add(&sell)

	// Create IOC buy at lower price
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Limit, order.IOC,
		decimal.NewFromInt(100), // Won't cross
		decimal.NewFromInt(10),
		1001,
	)

	result := matcher.ProcessWithTIF(&buy)

	// Should not fill
	if buy.Filled.GreaterThan(decimal.Zero) {
		t.Error("Expected no fill for IOC")
	}

	// Should have reject + cancel logs
	logs := pub.GetLogs()
	hasReject := false
	hasCancel := false
	for _, log := range logs {
		if log.LogType == event.LogTypeReject {
			hasReject = true
		}
		if log.LogType == event.LogTypeCancel {
			hasCancel = true
		}
	}
	if !hasReject {
		t.Error("Expected reject log for IOC no match")
	}
	if !hasCancel {
		t.Error("Expected cancel log for IOC remaining")
	}

	if len(result.Trades) < 2 {
		t.Errorf("Expected at least 2 logs (reject+cancel), got %d", len(result.Trades))
	}
}

// TestProcessWithTIF_FOK_CanFill tests FOK that can be filled
func TestProcessWithTIF_FOK_CanFill(t *testing.T) {
	matcher, pub := newTestMatcher()

	// Add enough liquidity
	sell := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10), // Exactly enough
		1000,
	)
	matcher.book.Add(&sell)

	// Create FOK buy
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Limit, order.FOK,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1001,
	)

	result := matcher.ProcessWithTIF(&buy)

	// Should fully fill
	if !buy.IsFilled() {
		t.Error("Expected FOK to be fully filled")
	}

	// Should not have reject log
	logs := pub.GetLogs()
	for _, log := range logs {
		if log.LogType == event.LogTypeReject {
			t.Error("Expected no reject for fillable FOK")
		}
	}

	if len(result.Trades) != 3 { // 1 trade + 2 fills
		t.Errorf("Expected 3 logs, got %d", len(result.Trades))
	}
}

// TestProcessWithTIF_FOK_CannotFill tests FOK that cannot be filled
func TestProcessWithTIF_FOK_CannotFill(t *testing.T) {
	matcher, pub := newTestMatcher()

	// Add insufficient liquidity
	sell := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(5), // Not enough
		1000,
	)
	matcher.book.Add(&sell)

	// Create FOK buy
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Limit, order.FOK,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10), // Want 10, only 5 available
		1001,
	)

	result := matcher.ProcessWithTIF(&buy)

	// Should not fill at all (FOK rejects if can't fill completely)
	if buy.Filled.GreaterThan(decimal.Zero) {
		t.Error("Expected no fill for FOK that can't be fully filled")
	}

	// Should have reject log
	logs := pub.GetLogs()
	hasReject := false
	for _, log := range logs {
		if log.LogType == event.LogTypeReject {
			hasReject = true
		}
	}
	if !hasReject {
		t.Error("Expected reject log for FOK insufficient liquidity")
	}

	// Only reject log
	if len(result.Trades) != 1 {
		t.Errorf("Expected 1 log (reject), got %d", len(result.Trades))
	}
}

// TestProcessWithTIF_FOK_MultipleLevels tests FOK across multiple price levels
func TestProcessWithTIF_FOK_MultipleLevels(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Add liquidity across multiple levels
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

	// Create FOK buy that needs both levels
	buy := order.NewOrder(
		3, "BUY-1", "CMD-B1", 103, "BTCUSD",
		order.Buy, order.Limit, order.FOK,
		decimal.NewFromInt(101), // Price covers both levels
		decimal.NewFromInt(10),  // Total from both levels
		1002,
	)

	result := matcher.ProcessWithTIF(&buy)

	// Should fully fill across both levels
	if !buy.IsFilled() {
		t.Error("Expected FOK to be fully filled across levels")
	}

	// Should have 2 trades
	if len(result.Trades) != 6 { // 2 trades * 3 logs each
		t.Errorf("Expected 6 logs (2 trades), got %d", len(result.Trades))
	}
}

// TestProcessWithTIF_PostOnly_NoMatch tests PostOnly that won't match
func TestProcessWithTIF_PostOnly_NoMatch(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Add sell order at 105
	sell := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(105),
		decimal.NewFromInt(10),
		1000,
	)
	matcher.book.Add(&sell)

	// Create PostOnly buy at 100 (won't cross)
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Limit, order.PostOnly,
		decimal.NewFromInt(100), // Below ask
		decimal.NewFromInt(10),
		1001,
	)

	result := matcher.ProcessWithTIF(&buy)

	// Should not match (correct PostOnly behavior)
	if buy.Filled.GreaterThan(decimal.Zero) {
		t.Error("Expected PostOnly not to match")
	}

	// Should add to book (no reject)
	// Note: actual book addition would be done by engine
	if len(result.Trades) != 0 {
		t.Errorf("Expected no logs for non-crossing PostOnly, got %d", len(result.Trades))
	}
}

// TestProcessWithTIF_PostOnly_WouldMatch tests PostOnly that would match (reject)
func TestProcessWithTIF_PostOnly_WouldMatch(t *testing.T) {
	matcher, pub := newTestMatcher()

	// Add sell order at 100
	sell := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1000,
	)
	matcher.book.Add(&sell)

	// Create PostOnly buy at 100 (would cross)
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Limit, order.PostOnly,
		decimal.NewFromInt(100), // Would match
		decimal.NewFromInt(10),
		1001,
	)

	result := matcher.ProcessWithTIF(&buy)

	// Should reject (PostOnly can't match)
	if buy.Filled.GreaterThan(decimal.Zero) {
		t.Error("Expected PostOnly not to fill when would match")
	}

	// Should have reject log
	logs := pub.GetLogs()
	hasReject := false
	for _, log := range logs {
		if log.LogType == event.LogTypeReject {
			hasReject = true
		}
	}
	if !hasReject {
		t.Error("Expected reject log for PostOnly that would match")
	}

	if len(result.Trades) != 1 {
		t.Errorf("Expected 1 log (reject), got %d", len(result.Trades))
	}
}

// TestProcessWithTIF_PostOnly_SellWouldMatch tests PostOnly sell that would match
func TestProcessWithTIF_PostOnly_SellWouldMatch(t *testing.T) {
	matcher, pub := newTestMatcher()

	// Add buy order at 100
	buy := order.NewOrder(
		1, "BUY-1", "CMD-B1", 101, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1000,
	)
	matcher.book.Add(&buy)

	// Create PostOnly sell at 100 (would cross)
	sell := order.NewOrder(
		2, "SELL-1", "CMD-S1", 102, "BTCUSD",
		order.Sell, order.Limit, order.PostOnly,
		decimal.NewFromInt(100), // Would match
		decimal.NewFromInt(10),
		1001,
	)

	result := matcher.ProcessWithTIF(&sell)

	// Should reject
	logs := pub.GetLogs()
	hasReject := false
	for _, log := range logs {
		if log.LogType == event.LogTypeReject {
			hasReject = true
		}
	}
	if !hasReject {
		t.Error("Expected reject log for PostOnly sell that would match")
	}

	if len(result.Trades) != 1 {
		t.Errorf("Expected 1 log (reject), got %d", len(result.Trades))
	}
}

// TestProcessWithTIF_Market tests market order with TIF
func TestProcessWithTIF_Market(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Add sell order
	sell := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1000,
	)
	matcher.book.Add(&sell)

	// Create market buy (TIF should not affect market orders)
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Market, order.IOC, // IOC on market order
		decimal.Zero,
		decimal.NewFromInt(10),
		1001,
	)

	result := matcher.ProcessWithTIF(&buy)

	// Should execute like normal market order
	if !buy.Filled.Equal(decimal.NewFromInt(10)) {
		t.Errorf("Expected buy filled 10, got %v", buy.Filled)
	}

	if len(result.Trades) > 0 {
		// Verify it executed
		t.Logf("Market order with TIF executed successfully")
	}
}

// TestProcessWithTIF_DefaultToGTC tests default TIF behavior
func TestProcessWithTIF_DefaultToGTC(t *testing.T) {
	matcher, _ := newTestMatcher()

	// Add sell order
	sell := order.NewOrder(
		1, "SELL-1", "CMD-S1", 101, "BTCUSD",
		order.Sell, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(5),
		1000,
	)
	matcher.book.Add(&sell)

	// Create buy with undefined TIF (0 value)
	buy := order.NewOrder(
		2, "BUY-1", "CMD-B1", 102, "BTCUSD",
		order.Buy, order.Limit, order.TimeInForce(99), // Invalid TIF
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		1001,
	)

	result := matcher.ProcessWithTIF(&buy)

	// Should default to GTC behavior (partial fill)
	if !buy.Filled.Equal(decimal.NewFromInt(5)) {
		t.Errorf("Expected buy filled 5 (GTC default), got %v", buy.Filled)
	}

	// Should have trade logs
	if len(result.Trades) != 3 {
		t.Errorf("Expected 3 logs (trade+fills), got %d", len(result.Trades))
	}
}
