package protocol

import (
	"github.com/shopspring/decimal"
)

// OrderBookLogRef is a lightweight reference to an event log
// Used to avoid import cycles between protocol and event packages
type OrderBookLogRef interface {
	GetSeqID() uint64
	GetLogType() string
}

// PlaceOrderResult contains the result of a place order operation
type PlaceOrderResult struct {
	OrderID     string          `json:"order_id"`
	Accepted    bool            `json:"accepted"`
	Filled      decimal.Decimal `json:"filled"`       // Amount filled
	Remaining   decimal.Decimal `json:"remaining"`    // Amount remaining
	Trades      []interface{}   `json:"trades"`       // Trades generated (event.OrderBookLog)
	InBook      bool            `json:"in_book"`      // True if order is resting in book
	PartialFill bool            `json:"partial_fill"` // True if partially filled
}

// CancelOrderResult contains the result of a cancel order operation
type CancelOrderResult struct {
	OrderID       string          `json:"order_id"`
	Cancelled     bool            `json:"cancelled"`
	FilledBefore  decimal.Decimal `json:"filled_before"`  // Amount filled before cancel
	CancelledSize decimal.Decimal `json:"cancelled_size"` // Amount cancelled
}

// AmendOrderResult contains the result of an amend order operation
type AmendOrderResult struct {
	OrderID        string          `json:"order_id"`
	Amended        bool            `json:"amended"`
	NewPrice       decimal.Decimal `json:"new_price"`
	NewSize        decimal.Decimal `json:"new_size"`
	Trades         []interface{}   `json:"trades"`           // Trades generated on re-match (event.OrderBookLog)
	MatchedOnAmend bool            `json:"matched_on_amend"` // True if trades occurred
}
