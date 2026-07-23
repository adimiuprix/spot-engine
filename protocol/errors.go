package protocol

import "errors"

// Request validation errors (returned before enqueue)
var (
	ErrInvalidCommandID  = errors.New("invalid_command_id: CommandID cannot be empty")
	ErrInvalidTimestamp  = errors.New("invalid_timestamp: Timestamp must be positive")
	ErrInvalidMarketID   = errors.New("invalid_market_id: MarketID cannot be empty")
	ErrInvalidOrderID    = errors.New("invalid_order_id: OrderID cannot be empty")
	ErrInvalidSide       = errors.New("invalid_side: Side must be 'buy' or 'sell'")
	ErrInvalidOrderType  = errors.New("invalid_order_type: OrderType must be 'limit' or 'market'")
	ErrInvalidSize       = errors.New("invalid_size: Size must be positive")
	ErrInvalidPrice      = errors.New("invalid_price: Price must be positive")
	ErrInvalidLotSize    = errors.New("invalid_lot_size: LotSize must be positive")
	ErrInvalidEventType  = errors.New("invalid_event_type: EventType cannot be empty")
	ErrQueueFull         = errors.New("queue_full: Ring buffer is full")
	ErrEngineShutdown    = errors.New("engine_shutdown: Engine is shutting down")
	ErrNotFound          = errors.New("not_found: Resource not found")
)

// RejectReason represents business-level rejection reasons (emitted as logs)
type RejectReason string

const (
	// Order-level rejections
	RejectReasonInvalidPayload     RejectReason = "invalid_payload"
	RejectReasonDuplicateOrderID   RejectReason = "duplicate_order_id"
	RejectReasonOrderNotFound      RejectReason = "order_not_found"
	RejectReasonInvalidOrderOwner  RejectReason = "invalid_order_owner"
	RejectReasonInsufficientSize   RejectReason = "insufficient_size"
	RejectReasonBelowMinLotSize    RejectReason = "below_min_lot_size"
	RejectReasonInvalidIcebergSize RejectReason = "invalid_iceberg_size"

	// Market-level rejections
	RejectReasonMarketNotFound      RejectReason = "market_not_found"
	RejectReasonMarketAlreadyExists RejectReason = "market_already_exists"
	RejectReasonMarketSuspended     RejectReason = "market_suspended"
	RejectReasonMarketHalted        RejectReason = "market_halted"

	// General rejections
	RejectReasonUnknownCommand RejectReason = "unknown_command"
	RejectReasonInternalError  RejectReason = "internal_error"
)

func (r RejectReason) String() string {
	return string(r)
}
