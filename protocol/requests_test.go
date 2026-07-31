package protocol

import (
	"testing"

	"github.com/shopspring/decimal"
)

// TestCancelOrderRequestValidation tests cancel request validation
func TestCancelOrderRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     CancelOrderRequest
		wantErr bool
	}{
		{
			name: "valid cancel request",
			req: CancelOrderRequest{
				CommandID: "cmd-cancel-1",
				UserID:    101,
				Symbol:    "BTCUSD",
				OrderID:   "order-1",
				Timestamp: 1000,
			},
			wantErr: false,
		},
		{
			name: "missing CommandID",
			req: CancelOrderRequest{
				CommandID: "",
				UserID:    101,
				Symbol:    "BTCUSD",
				OrderID:   "order-1",
				Timestamp: 1000,
			},
			wantErr: true,
		},
		{
			name: "invalid timestamp",
			req: CancelOrderRequest{
				CommandID: "cmd-cancel-1",
				UserID:    101,
				Symbol:    "BTCUSD",
				OrderID:   "order-1",
				Timestamp: 0,
			},
			wantErr: true,
		},
		{
			name: "missing Symbol",
			req: CancelOrderRequest{
				CommandID: "cmd-cancel-1",
				UserID:    101,
				Symbol:    "",
				OrderID:   "order-1",
				Timestamp: 1000,
			},
			wantErr: true,
		},
		{
			name: "missing OrderID",
			req: CancelOrderRequest{
				CommandID: "cmd-cancel-1",
				UserID:    101,
				Symbol:    "BTCUSD",
				OrderID:   "",
				Timestamp: 1000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CancelOrderRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAmendOrderRequestValidation tests amend request validation
func TestAmendOrderRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     AmendOrderRequest
		wantErr bool
	}{
		{
			name: "valid amend request",
			req: AmendOrderRequest{
				BaseCommand: BaseCommand{
					CommandID: "cmd-amend-1",
					UserID:    101,
					MarketID:  "BTCUSD",
					Timestamp: 1000,
				},
				OrderID:  "order-1",
				NewPrice: decimal.NewFromInt(100),
				NewSize:  decimal.NewFromInt(10),
			},
			wantErr: false,
		},
		{
			name: "missing OrderID",
			req: AmendOrderRequest{
				BaseCommand: BaseCommand{
					CommandID: "cmd-amend-1",
					UserID:    101,
					MarketID:  "BTCUSD",
					Timestamp: 1000,
				},
				OrderID:  "",
				NewPrice: decimal.NewFromInt(100),
				NewSize:  decimal.NewFromInt(10),
			},
			wantErr: true,
		},
		{
			name: "invalid NewPrice (zero)",
			req: AmendOrderRequest{
				BaseCommand: BaseCommand{
					CommandID: "cmd-amend-1",
					UserID:    101,
					MarketID:  "BTCUSD",
					Timestamp: 1000,
				},
				OrderID:  "order-1",
				NewPrice: decimal.Zero,
				NewSize:  decimal.NewFromInt(10),
			},
			wantErr: true,
		},
		{
			name: "invalid NewSize (negative)",
			req: AmendOrderRequest{
				BaseCommand: BaseCommand{
					CommandID: "cmd-amend-1",
					UserID:    101,
					MarketID:  "BTCUSD",
					Timestamp: 1000,
				},
				OrderID:  "order-1",
				NewPrice: decimal.NewFromInt(100),
				NewSize:  decimal.NewFromInt(-5),
			},
			wantErr: true,
		},
		{
			name: "missing BaseCommand fields",
			req: AmendOrderRequest{
				BaseCommand: BaseCommand{
					CommandID: "",
					UserID:    101,
					MarketID:  "BTCUSD",
					Timestamp: 1000,
				},
				OrderID:  "order-1",
				NewPrice: decimal.NewFromInt(100),
				NewSize:  decimal.NewFromInt(10),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("AmendOrderRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestBaseCommandValidation tests BaseCommand validation
func TestBaseCommandValidation(t *testing.T) {
	tests := []struct {
		name    string
		cmd     BaseCommand
		wantErr bool
	}{
		{
			name: "valid BaseCommand",
			cmd: BaseCommand{
				CommandID: "cmd-1",
				UserID:    101,
				MarketID:  "BTCUSD",
				Timestamp: 1000,
			},
			wantErr: false,
		},
		{
			name: "missing CommandID",
			cmd: BaseCommand{
				CommandID: "",
				UserID:    101,
				MarketID:  "BTCUSD",
				Timestamp: 1000,
			},
			wantErr: true,
		},
		{
			name: "invalid Timestamp",
			cmd: BaseCommand{
				CommandID: "cmd-1",
				UserID:    101,
				MarketID:  "BTCUSD",
				Timestamp: 0,
			},
			wantErr: true,
		},
		{
			name: "missing MarketID",
			cmd: BaseCommand{
				CommandID: "cmd-1",
				UserID:    101,
				MarketID:  "",
				Timestamp: 1000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("BaseCommand.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPlaceOrderRequestValidation_MarketOrders tests market order validation
func TestPlaceOrderRequestValidation_MarketOrders(t *testing.T) {
	tests := []struct {
		name    string
		req     PlaceOrderRequest
		wantErr bool
	}{
		{
			name: "valid market order with Size",
			req: PlaceOrderRequest{
				BaseCommand: BaseCommand{
					CommandID: "cmd-1",
					UserID:    101,
					MarketID:  "BTCUSD",
					Timestamp: 1000,
				},
				OrderID:   "order-1",
				Side:      "buy",
				OrderType: "market",
				Size:      decimal.NewFromInt(10),
			},
			wantErr: false,
		},
		{
			name: "valid market order with QuoteSize",
			req: PlaceOrderRequest{
				BaseCommand: BaseCommand{
					CommandID: "cmd-1",
					UserID:    101,
					MarketID:  "BTCUSD",
					Timestamp: 1000,
				},
				OrderID:   "order-1",
				Side:      "buy",
				OrderType: "market",
				QuoteSize: decimal.NewFromInt(1000),
			},
			wantErr: false,
		},
		{
			name: "invalid market order (no Size or QuoteSize)",
			req: PlaceOrderRequest{
				BaseCommand: BaseCommand{
					CommandID: "cmd-1",
					UserID:    101,
					MarketID:  "BTCUSD",
					Timestamp: 1000,
				},
				OrderID:   "order-1",
				Side:      "buy",
				OrderType: "market",
				Size:      decimal.Zero,
				QuoteSize: decimal.Zero,
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

// TestRejectReasonString tests RejectReason string conversion
func TestRejectReasonString(t *testing.T) {
	tests := []struct {
		reason RejectReason
		want   string
	}{
		{RejectReasonInvalidPayload, "invalid_payload"},
		{RejectReasonOrderNotFound, "order_not_found"},
		{RejectReasonNoLiquidity, "no_liquidity"},
		{RejectReasonMarketSuspended, "market_suspended"},
	}

	for _, tt := range tests {
		t.Run(string(tt.reason), func(t *testing.T) {
			if got := tt.reason.String(); got != tt.want {
				t.Errorf("RejectReason.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
