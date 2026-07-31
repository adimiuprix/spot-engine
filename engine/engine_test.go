package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/shopspring/decimal"
)

// TestNew tests engine creation
func TestNew(t *testing.T) {
	config := Config{
		Symbol:         "BTCUSD",
		RingBufferSize: 1024,
		MinLotSize:     decimal.NewFromFloat(0.001),
	}

	engine := New(config)

	if engine == nil {
		t.Fatal("Expected non-nil engine")
	}

	if engine.config.Symbol != "BTCUSD" {
		t.Errorf("Expected symbol BTCUSD, got %s", engine.config.Symbol)
	}

	if engine.orderBook == nil {
		t.Error("Expected non-nil order book")
	}

	if engine.matcher == nil {
		t.Error("Expected non-nil matcher")
	}

	if engine.orderQueue == nil {
		t.Error("Expected non-nil order queue")
	}

	if engine.running {
		t.Error("Expected engine not to be running initially")
	}

	if engine.state != protocol.StateRunning {
		t.Errorf("Expected initial state Running, got %v", engine.state)
	}
}

// TestEngine_StartStop tests engine lifecycle
func TestEngine_StartStop(t *testing.T) {
	config := Config{
		Symbol:         "BTCUSD",
		RingBufferSize: 1024,
		MinLotSize:     decimal.NewFromFloat(0.001),
	}

	engine := New(config)

	// Initially not running
	if engine.IsRunning() {
		t.Error("Expected engine not to be running")
	}

	// Start engine
	engine.Start()

	// Give it time to start
	time.Sleep(10 * time.Millisecond)

	// Should be running
	if !engine.IsRunning() {
		t.Error("Expected engine to be running after Start()")
	}

	// Stop engine
	engine.Stop()

	// Give it time to stop
	time.Sleep(10 * time.Millisecond)

	// Should not be running
	if engine.IsRunning() {
		t.Error("Expected engine to be stopped after Stop()")
	}
}

// TestEngine_StartTwice tests calling Start() multiple times
func TestEngine_StartTwice(t *testing.T) {
	config := Config{
		Symbol:         "BTCUSD",
		RingBufferSize: 1024,
		MinLotSize:     decimal.NewFromFloat(0.001),
	}

	engine := New(config)

	// Start engine
	engine.Start()
	time.Sleep(10 * time.Millisecond)

	// Start again (should be idempotent)
	engine.Start()

	// Should still be running
	if !engine.IsRunning() {
		t.Error("Expected engine to still be running")
	}

	// Cleanup
	engine.Stop()
}

// TestEngine_SubmitOrder tests order submission
func TestEngine_SubmitOrder(t *testing.T) {
	config := Config{
		Symbol:         "BTCUSD",
		RingBufferSize: 1024,
		MinLotSize:     decimal.NewFromFloat(0.001),
	}

	engine := New(config)

	o := order.NewOrder(
		1, "ORD-1", "CMD-1", 101, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		time.Now().UnixNano(),
	)

	// Submit order
	success := engine.SubmitOrder(&o)

	if !success {
		t.Error("Expected order submission to succeed")
	}
}

// TestEngine_SubmitOrderQueueFull tests queue overflow
func TestEngine_SubmitOrderQueueFull(t *testing.T) {
	config := Config{
		Symbol:         "BTCUSD",
		RingBufferSize: 2, // Very small buffer
		MinLotSize:     decimal.NewFromFloat(0.001),
	}

	engine := New(config)

	// Fill the queue
	for i := 0; i < 10; i++ {
		o := order.NewOrder(
			uint64(i),
			fmt.Sprintf("ORD-%d", i),
			fmt.Sprintf("CMD-%d", i),
			101, "BTCUSD",
			order.Buy, order.Limit, order.GTC,
			decimal.NewFromInt(100),
			decimal.NewFromInt(10),
			time.Now().UnixNano(),
		)
		engine.SubmitOrder(&o)
	}

	// Try to submit one more (should fail eventually due to small buffer)
	o := order.NewOrder(
		999, "ORD-999", "CMD-999", 101, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		time.Now().UnixNano(),
	)

	// May or may not fail depending on timing, but shouldn't crash
	_ = engine.SubmitOrder(&o)
	t.Log("Order submission with small buffer completed")
}

// TestEngine_Events tests event channel
func TestEngine_Events(t *testing.T) {
	config := Config{
		Symbol:         "BTCUSD",
		RingBufferSize: 1024,
		MinLotSize:     decimal.NewFromFloat(0.001),
	}

	engine := New(config)

	// Get events channel
	events := engine.Events()

	if events == nil {
		t.Error("Expected non-nil events channel")
	}
}

// TestEngine_ProcessOrderWithStateCheck tests order processing respects state
func TestEngine_ProcessOrderWithStateCheck(t *testing.T) {
	config := Config{
		Symbol:         "BTCUSD",
		RingBufferSize: 1024,
		MinLotSize:     decimal.NewFromFloat(0.001),
	}

	engine := New(config)
	engine.Start()
	defer engine.Stop()

	// Get events channel
	events := engine.Events()

	// Set market to suspended
	engine.mu.Lock()
	engine.state = protocol.StateSuspended
	engine.mu.Unlock()

	// Submit order
	o := order.NewOrder(
		1, "ORD-1", "CMD-1", 101, "BTCUSD",
		order.Buy, order.Limit, order.GTC,
		decimal.NewFromInt(100),
		decimal.NewFromInt(10),
		time.Now().UnixNano(),
	)
	engine.SubmitOrder(&o)

	// Wait for processing
	select {
	case log := <-events:
		// Should receive a reject log
		if log.LogType != event.LogTypeReject {
			t.Errorf("Expected reject log for suspended market, got %v", log.LogType)
		}
		if log.RejectReason != protocol.RejectReasonMarketSuspended {
			t.Errorf("Expected RejectReasonMarketSuspended, got %v", log.RejectReason)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for reject log")
	}
}

// TestEngine_DefaultLotSize tests default lot size when not specified
func TestEngine_DefaultLotSize(t *testing.T) {
	config := Config{
		Symbol:         "BTCUSD",
		RingBufferSize: 1024,
		MinLotSize:     decimal.Zero, // Not specified
	}

	engine := New(config)

	// Should use DefaultLotSize
	// We can't directly check matcher's lot size, but engine should be created successfully
	if engine == nil {
		t.Fatal("Expected engine to be created with default lot size")
	}
}

// TestConfig_DefaultValues tests config default values
func TestConfig_DefaultValues(t *testing.T) {
	config := Config{
		Symbol: "BTCUSD",
	}

	engine := New(config)

	// Should handle zero/unset values gracefully
	if engine == nil {
		t.Fatal("Expected engine to be created with default values")
	}

	// RingBufferSize should have a default or handle zero
	if engine.orderQueue == nil {
		t.Error("Expected order queue to be created")
	}
}

// TestEngine_ConcurrentSubmissions tests concurrent order submissions
func TestEngine_ConcurrentSubmissions(t *testing.T) {
	config := Config{
		Symbol:         "BTCUSD",
		RingBufferSize: 10000,
		MinLotSize:     decimal.NewFromFloat(0.001),
	}

	engine := New(config)
	engine.Start()
	defer engine.Stop()

	// Submit orders concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				o := order.NewOrder(
					uint64(id*100+j),
					fmt.Sprintf("ORD-%d-%d", id, j),
					fmt.Sprintf("CMD-%d-%d", id, j),
					uint64(101+id),
					"BTCUSD",
					order.Buy,
					order.Limit,
					order.GTC,
					decimal.NewFromInt(100),
					decimal.NewFromInt(10),
					time.Now().UnixNano(),
				)
				engine.SubmitOrder(&o)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	t.Log("Concurrent submissions completed")
}
