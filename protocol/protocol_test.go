package protocol

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestPlaceOrderRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     PlaceOrderRequest
		wantErr bool
	}{
		{
			name: "valid limit buy order",
			req: PlaceOrderRequest{
				BaseCommand: BaseCommand{
					Type:      CmdPlaceOrder,
					CommandID: "cmd-1",
					UserID:    100,
					MarketID:  "BTC-USDT",
					Timestamp: time.Now().UnixNano(),
				},
				OrderID:   "order-1",
				Side:      "buy",
				OrderType: "limit",
				Price:     decimal.NewFromInt(50000),
				Size:      decimal.NewFromFloat(0.5),
			},
			wantErr: false,
		},
		{
			name: "missing CommandID",
			req: PlaceOrderRequest{
				BaseCommand: BaseCommand{
					Type:      CmdPlaceOrder,
					CommandID: "",
					UserID:    100,
					MarketID:  "BTC-USDT",
					Timestamp: time.Now().UnixNano(),
				},
				OrderID:   "order-1",
				Side:      "buy",
				OrderType: "limit",
				Price:     decimal.NewFromInt(50000),
				Size:      decimal.NewFromFloat(0.5),
			},
			wantErr: true,
		},
		{
			name: "invalid timestamp",
			req: PlaceOrderRequest{
				BaseCommand: BaseCommand{
					Type:      CmdPlaceOrder,
					CommandID: "cmd-1",
					UserID:    100,
					MarketID:  "BTC-USDT",
					Timestamp: 0,
				},
				OrderID:   "order-1",
				Side:      "buy",
				OrderType: "limit",
				Price:     decimal.NewFromInt(50000),
				Size:      decimal.NewFromFloat(0.5),
			},
			wantErr: true,
		},
		{
			name: "invalid side",
			req: PlaceOrderRequest{
				BaseCommand: BaseCommand{
					Type:      CmdPlaceOrder,
					CommandID: "cmd-1",
					UserID:    100,
					MarketID:  "BTC-USDT",
					Timestamp: time.Now().UnixNano(),
				},
				OrderID:   "order-1",
				Side:      "invalid",
				OrderType: "limit",
				Price:     decimal.NewFromInt(50000),
				Size:      decimal.NewFromFloat(0.5),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PlaceOrderRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSerialization(t *testing.T) {
	req := &PlaceOrderRequest{
		BaseCommand: BaseCommand{
			Type:      CmdPlaceOrder,
			CommandID: "cmd-1",
			UserID:    100,
			MarketID:  "BTC-USDT",
			Timestamp: time.Now().UnixNano(),
		},
		OrderID:   "order-1",
		Side:      "buy",
		OrderType: "limit",
		Price:     decimal.NewFromInt(50000),
		Size:      decimal.NewFromFloat(0.5),
	}

	// Marshal
	data, err := MarshalRequest(req)
	if err != nil {
		t.Fatalf("MarshalRequest failed: %v", err)
	}

	// Unmarshal
	result, err := UnmarshalRequest(data)
	if err != nil {
		t.Fatalf("UnmarshalRequest failed: %v", err)
	}

	// Type assertion
	unmarshaled, ok := result.(*PlaceOrderRequest)
	if !ok {
		t.Fatalf("UnmarshalRequest returned wrong type")
	}

	// Verify fields
	if unmarshaled.CommandID != req.CommandID {
		t.Errorf("CommandID mismatch: got %s, want %s", unmarshaled.CommandID, req.CommandID)
	}
	if unmarshaled.OrderID != req.OrderID {
		t.Errorf("OrderID mismatch: got %s, want %s", unmarshaled.OrderID, req.OrderID)
	}
	if !unmarshaled.Price.Equal(req.Price) {
		t.Errorf("Price mismatch: got %s, want %s", unmarshaled.Price, req.Price)
	}
}

func TestOrderBookState(t *testing.T) {
	tests := []struct {
		state           OrderBookState
		canPlace        bool
		canCancel       bool
		canAmend        bool
	}{
		{StateRunning, true, true, true},
		{StateSuspended, false, true, false},
		{StateHalted, false, false, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := tt.state.CanPlaceOrder(); got != tt.canPlace {
				t.Errorf("CanPlaceOrder() = %v, want %v", got, tt.canPlace)
			}
			if got := tt.state.CanCancelOrder(); got != tt.canCancel {
				t.Errorf("CanCancelOrder() = %v, want %v", got, tt.canCancel)
			}
			if got := tt.state.CanAmendOrder(); got != tt.canAmend {
				t.Errorf("CanAmendOrder() = %v, want %v", got, tt.canAmend)
			}
		})
	}
}
