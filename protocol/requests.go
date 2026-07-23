package protocol

import "github.com/shopspring/decimal"

// PlaceOrderRequest represents a request to place a new order
type PlaceOrderRequest struct {
	BaseCommand
	OrderID     string          `json:"order_id"`
	Side        string          `json:"side"`         // "buy" or "sell"
	OrderType   string          `json:"order_type"`   // "limit" or "market"
	Price       decimal.Decimal `json:"price"`        // Required for limit orders
	Size        decimal.Decimal `json:"size"`         // Total order quantity
	VisibleSize decimal.Decimal `json:"visible_size"` // For iceberg orders (0 = not iceberg)
	QuoteSize   decimal.Decimal `json:"quote_size"`   // For market orders in quote currency
}

// Validate checks if the place order request is valid
func (r *PlaceOrderRequest) Validate() error {
	if err := r.BaseCommand.Validate(); err != nil {
		return err
	}
	if r.OrderID == "" {
		return ErrInvalidOrderID
	}
	if r.Side != "buy" && r.Side != "sell" {
		return ErrInvalidSide
	}
	if r.OrderType != "limit" && r.OrderType != "market" {
		return ErrInvalidOrderType
	}
	if r.Size.LessThanOrEqual(decimal.Zero) && r.QuoteSize.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidSize
	}
	return nil
}

// CancelOrderRequest represents a request to cancel an existing order
type CancelOrderRequest struct {
	CommandID string // Unique command ID for idempotency
	UserID    uint64 // User making the cancellation
	Symbol    string // Market symbol (e.g., "BTCUSD")
	OrderID   string // Order to cancel
	Timestamp int64  // Command timestamp (for determinism)
}

// Validate checks if the cancel order request is valid
func (r *CancelOrderRequest) Validate() error {
	if r.CommandID == "" {
		return ErrInvalidCommandID
	}
	if r.Timestamp <= 0 {
		return ErrInvalidTimestamp
	}
	if r.Symbol == "" {
		return ErrInvalidMarketID
	}
	if r.OrderID == "" {
		return ErrInvalidOrderID
	}
	return nil
}

// AmendOrderRequest represents a request to modify an existing order
type AmendOrderRequest struct {
	BaseCommand
	OrderID  string          `json:"order_id"`
	NewPrice decimal.Decimal `json:"new_price"` // New price (string-encoded for precision)
	NewSize  decimal.Decimal `json:"new_size"`  // New total size (string-encoded for precision)
}

// Validate checks if the amend order request is valid
func (r *AmendOrderRequest) Validate() error {
	if err := r.BaseCommand.Validate(); err != nil {
		return err
	}
	if r.OrderID == "" {
		return ErrInvalidOrderID
	}
	if r.NewPrice.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidPrice
	}
	if r.NewSize.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidSize
	}
	return nil
}

// CreateMarketRequest represents a request to create a new market
type CreateMarketRequest struct {
	BaseCommand
	MinLotSize decimal.Decimal `json:"min_lot_size"` // Minimum trade unit
}

// Validate checks if the create market request is valid
func (r *CreateMarketRequest) Validate() error {
	if err := r.BaseCommand.Validate(); err != nil {
		return err
	}
	if r.MinLotSize.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidLotSize
	}
	return nil
}

// SuspendMarketRequest represents a request to suspend trading on a market
type SuspendMarketRequest struct {
	BaseCommand
	Reason string `json:"reason"` // Human-readable reason for suspension
}

// Validate checks if the suspend market request is valid
func (r *SuspendMarketRequest) Validate() error {
	return r.BaseCommand.Validate()
}

// ResumeMarketRequest represents a request to resume trading on a market
type ResumeMarketRequest struct {
	BaseCommand
}

// Validate checks if the resume market request is valid
func (r *ResumeMarketRequest) Validate() error {
	return r.BaseCommand.Validate()
}

// UpdateConfigRequest represents a request to update market configuration
type UpdateConfigRequest struct {
	BaseCommand
	MinLotSize decimal.Decimal `json:"min_lot_size"` // New minimum trade unit
}

// Validate checks if the update config request is valid
func (r *UpdateConfigRequest) Validate() error {
	if err := r.BaseCommand.Validate(); err != nil {
		return err
	}
	if r.MinLotSize.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidLotSize
	}
	return nil
}

// UserEventRequest represents a custom user event (for extensions)
type UserEventRequest struct {
	BaseCommand
	EventType string                 `json:"event_type"`
	Payload   map[string]interface{} `json:"payload"`
}

// Validate checks if the user event request is valid
func (r *UserEventRequest) Validate() error {
	if err := r.BaseCommand.Validate(); err != nil {
		return err
	}
	if r.EventType == "" {
		return ErrInvalidEventType
	}
	return nil
}
