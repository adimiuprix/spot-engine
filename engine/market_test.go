package engine

import (
	"testing"

	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/shopspring/decimal"
)

// TestNewMarket tests market creation
func TestNewMarket(t *testing.T) {
	seqGen := event.NewSequenceGenerator(0)
	pub := event.NewNoOpPublisher()

	market := NewMarket("BTCUSD", decimal.NewFromFloat(0.001), seqGen, pub)

	if market == nil {
		t.Fatal("Expected non-nil market")
	}

	if market.ID != "BTCUSD" {
		t.Errorf("Expected ID BTCUSD, got %s", market.ID)
	}

	if !market.MinLotSize.Equal(decimal.NewFromFloat(0.001)) {
		t.Errorf("Expected MinLotSize 0.001, got %v", market.MinLotSize)
	}

	if market.State != protocol.StateRunning {
		t.Errorf("Expected initial state Running, got %v", market.State)
	}

	if market.OrderBook == nil {
		t.Error("Expected non-nil OrderBook")
	}

	if market.Matcher == nil {
		t.Error("Expected non-nil Matcher")
	}
}

// TestMarket_GetSetState tests state management
func TestMarket_GetSetState(t *testing.T) {
	seqGen := event.NewSequenceGenerator(0)
	pub := event.NewNoOpPublisher()
	market := NewMarket("BTCUSD", decimal.NewFromFloat(0.001), seqGen, pub)

	// Initial state
	if market.GetState() != protocol.StateRunning {
		t.Errorf("Expected initial state Running, got %v", market.GetState())
	}

	// Set to Suspended
	market.SetState(protocol.StateSuspended)
	if market.GetState() != protocol.StateSuspended {
		t.Errorf("Expected state Suspended, got %v", market.GetState())
	}

	// Set to Halted
	market.SetState(protocol.StateHalted)
	if market.GetState() != protocol.StateHalted {
		t.Errorf("Expected state Halted, got %v", market.GetState())
	}

	// Set back to Running
	market.SetState(protocol.StateRunning)
	if market.GetState() != protocol.StateRunning {
		t.Errorf("Expected state Running, got %v", market.GetState())
	}
}

// TestMarket_CanPlaceOrder tests order placement permission
func TestMarket_CanPlaceOrder(t *testing.T) {
	seqGen := event.NewSequenceGenerator(0)
	pub := event.NewNoOpPublisher()
	market := NewMarket("BTCUSD", decimal.NewFromFloat(0.001), seqGen, pub)

	tests := []struct {
		state    protocol.OrderBookState
		canPlace bool
	}{
		{protocol.StateRunning, true},
		{protocol.StateSuspended, false},
		{protocol.StateHalted, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			market.SetState(tt.state)
			if got := market.CanPlaceOrder(); got != tt.canPlace {
				t.Errorf("CanPlaceOrder() = %v, want %v for state %v", got, tt.canPlace, tt.state)
			}
		})
	}
}

// TestMarket_CanCancelOrder tests order cancellation permission
func TestMarket_CanCancelOrder(t *testing.T) {
	seqGen := event.NewSequenceGenerator(0)
	pub := event.NewNoOpPublisher()
	market := NewMarket("BTCUSD", decimal.NewFromFloat(0.001), seqGen, pub)

	tests := []struct {
		state     protocol.OrderBookState
		canCancel bool
	}{
		{protocol.StateRunning, true},
		{protocol.StateSuspended, true},
		{protocol.StateHalted, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			market.SetState(tt.state)
			if got := market.CanCancelOrder(); got != tt.canCancel {
				t.Errorf("CanCancelOrder() = %v, want %v for state %v", got, tt.canCancel, tt.state)
			}
		})
	}
}

// TestMarket_CanAmendOrder tests order amendment permission
func TestMarket_CanAmendOrder(t *testing.T) {
	seqGen := event.NewSequenceGenerator(0)
	pub := event.NewNoOpPublisher()
	market := NewMarket("BTCUSD", decimal.NewFromFloat(0.001), seqGen, pub)

	tests := []struct {
		state    protocol.OrderBookState
		canAmend bool
	}{
		{protocol.StateRunning, true},
		{protocol.StateSuspended, false},
		{protocol.StateHalted, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			market.SetState(tt.state)
			if got := market.CanAmendOrder(); got != tt.canAmend {
				t.Errorf("CanAmendOrder() = %v, want %v for state %v", got, tt.canAmend, tt.state)
			}
		})
	}
}

// TestMarket_UpdateConfig tests config update
func TestMarket_UpdateConfig(t *testing.T) {
	seqGen := event.NewSequenceGenerator(0)
	pub := event.NewNoOpPublisher()
	market := NewMarket("BTCUSD", decimal.NewFromFloat(0.001), seqGen, pub)

	// Initial lot size
	if !market.MinLotSize.Equal(decimal.NewFromFloat(0.001)) {
		t.Errorf("Expected initial MinLotSize 0.001, got %v", market.MinLotSize)
	}

	// Update lot size
	newLotSize := decimal.NewFromFloat(0.01)
	market.UpdateConfig(newLotSize)

	// Verify update
	if !market.MinLotSize.Equal(newLotSize) {
		t.Errorf("Expected MinLotSize %v, got %v", newLotSize, market.MinLotSize)
	}

	// Verify OrderBook was also updated
	if !market.OrderBook.MinLotSize.Equal(newLotSize) {
		t.Errorf("Expected OrderBook MinLotSize %v, got %v", newLotSize, market.OrderBook.MinLotSize)
	}
}

// TestMarket_ConcurrentStateAccess tests thread-safe state access
func TestMarket_ConcurrentStateAccess(t *testing.T) {
	seqGen := event.NewSequenceGenerator(0)
	pub := event.NewNoOpPublisher()
	market := NewMarket("BTCUSD", decimal.NewFromFloat(0.001), seqGen, pub)

	// Run concurrent reads and writes
	done := make(chan bool)
	
	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			market.SetState(protocol.StateRunning)
			market.SetState(protocol.StateSuspended)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			_ = market.GetState()
			_ = market.CanPlaceOrder()
			_ = market.CanCancelOrder()
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// If we reach here without data race, test passes
	t.Log("Concurrent access test passed")
}
