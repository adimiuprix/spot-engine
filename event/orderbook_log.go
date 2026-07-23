package event

import (
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/shopspring/decimal"
)

// OrderBookLog is the deterministic event model emitted by the matching engine
// This structure participates in replay and must contain only replay-stable fields
type OrderBookLog struct {
	// Event identification
	SeqID     uint64    `json:"seq_id"`      // Sequential log ID
	LogType   LogType   `json:"log_type"`    // Category: trade, fill, cancel, reject, admin
	EventType EventType `json:"event_type"`  // Specific event subtype

	// Command context (deterministic)
	CommandID string `json:"command_id"`          // Original command ID
	UserID    uint64 `json:"user_id"`             // Actor identity
	MarketID  string `json:"market_id"`           // Target market
	Timestamp int64  `json:"timestamp"`           // Logical event time (from command)
	SeqNum    uint64 `json:"seq_num,omitempty"`   // Optional: upstream sequence number

	// Order context
	OrderID       string          `json:"order_id,omitempty"`
	Side          string          `json:"side,omitempty"`           // "buy" or "sell"
	OrderType     string          `json:"order_type,omitempty"`     // "limit" or "market"
	Price         decimal.Decimal `json:"price,omitempty"`
	Size          decimal.Decimal `json:"size,omitempty"`
	FilledSize    decimal.Decimal `json:"filled_size,omitempty"`
	RemainingSize decimal.Decimal `json:"remaining_size,omitempty"`

	// Trade context (for LogTypeTrade)
	TradeID       uint64          `json:"trade_id,omitempty"`
	MakerOrderID  string          `json:"maker_order_id,omitempty"`
	TakerOrderID  string          `json:"taker_order_id,omitempty"`
	TradePrice    decimal.Decimal `json:"trade_price,omitempty"`
	TradeQuantity decimal.Decimal `json:"trade_quantity,omitempty"`

	// Rejection context (for LogTypeReject)
	RejectReason protocol.RejectReason `json:"reject_reason,omitempty"`
	RejectDetail string                 `json:"reject_detail,omitempty"`

	// Admin context (for LogTypeAdmin)
	AdminReason string                 `json:"admin_reason,omitempty"`
	OldState    protocol.OrderBookState `json:"old_state,omitempty"`
	NewState    protocol.OrderBookState `json:"new_state,omitempty"`
	ConfigChanges map[string]interface{} `json:"config_changes,omitempty"`
}

// NewTradeLog creates a trade execution log
func NewTradeLog(
	seqID uint64,
	commandID string,
	userID uint64,
	marketID string,
	timestamp int64,
	tradeID uint64,
	makerOrderID string,
	takerOrderID string,
	price decimal.Decimal,
	quantity decimal.Decimal,
	side string,
) *OrderBookLog {
	return &OrderBookLog{
		SeqID:         seqID,
		LogType:       LogTypeTrade,
		EventType:     EventTypeTrade,
		CommandID:     commandID,
		UserID:        userID,
		MarketID:      marketID,
		Timestamp:     timestamp,
		TradeID:       tradeID,
		MakerOrderID:  makerOrderID,
		TakerOrderID:  takerOrderID,
		TradePrice:    price,
		TradeQuantity: quantity,
		Side:          side,
	}
}

// NewFillLog creates an order fill notification log
func NewFillLog(
	seqID uint64,
	commandID string,
	userID uint64,
	marketID string,
	timestamp int64,
	orderID string,
	side string,
	price decimal.Decimal,
	filledSize decimal.Decimal,
	remainingSize decimal.Decimal,
	isFullFill bool,
) *OrderBookLog {
	eventType := EventTypePartialFill
	if isFullFill {
		eventType = EventTypeFill
	}

	return &OrderBookLog{
		SeqID:         seqID,
		LogType:       LogTypeFill,
		EventType:     eventType,
		CommandID:     commandID,
		UserID:        userID,
		MarketID:      marketID,
		Timestamp:     timestamp,
		OrderID:       orderID,
		Side:          side,
		Price:         price,
		FilledSize:    filledSize,
		RemainingSize: remainingSize,
	}
}

// NewCancelLog creates an order cancellation log
func NewCancelLog(
	seqID uint64,
	commandID string,
	userID uint64,
	marketID string,
	timestamp int64,
	orderID string,
	side string,
	price decimal.Decimal,
	remainingSize decimal.Decimal,
) *OrderBookLog {
	return &OrderBookLog{
		SeqID:         seqID,
		LogType:       LogTypeCancel,
		EventType:     EventTypeCancel,
		CommandID:     commandID,
		UserID:        userID,
		MarketID:      marketID,
		Timestamp:     timestamp,
		OrderID:       orderID,
		Side:          side,
		Price:         price,
		RemainingSize: remainingSize,
	}
}

// NewRejectLog creates a rejection log
func NewRejectLog(
	seqID uint64,
	commandID string,
	userID uint64,
	marketID string,
	timestamp int64,
	reason protocol.RejectReason,
	detail string,
	orderID string,
) *OrderBookLog {
	return &OrderBookLog{
		SeqID:        seqID,
		LogType:      LogTypeReject,
		EventType:    EventTypeReject,
		CommandID:    commandID,
		UserID:       userID,
		MarketID:     marketID,
		Timestamp:    timestamp,
		OrderID:      orderID,
		RejectReason: reason,
		RejectDetail: detail,
	}
}

// NewAdminLog creates a management command success log
func NewAdminLog(
	seqID uint64,
	commandID string,
	userID uint64,
	marketID string,
	timestamp int64,
	eventType EventType,
	oldState protocol.OrderBookState,
	newState protocol.OrderBookState,
	reason string,
	configChanges map[string]interface{},
) *OrderBookLog {
	return &OrderBookLog{
		SeqID:         seqID,
		LogType:       LogTypeAdmin,
		EventType:     eventType,
		CommandID:     commandID,
		UserID:        userID,
		MarketID:      marketID,
		Timestamp:     timestamp,
		OldState:      oldState,
		NewState:      newState,
		AdminReason:   reason,
		ConfigChanges: configChanges,
	}
}
