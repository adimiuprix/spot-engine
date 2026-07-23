package protocol

// OrderBookState represents the operational state of a market
type OrderBookState uint8

const (
	// StateRunning - Normal operation, all orders accepted
	StateRunning OrderBookState = iota

	// StateSuspended - Temporarily paused, no new orders accepted
	// Existing orders remain in book, can be cancelled
	StateSuspended

	// StateHalted - Emergency stop, no operations allowed
	// Trading completely frozen
	StateHalted
)

// String returns the string representation of the state
func (s OrderBookState) String() string {
	switch s {
	case StateRunning:
		return "running"
	case StateSuspended:
		return "suspended"
	case StateHalted:
		return "halted"
	default:
		return "unknown"
	}
}

// CanAcceptOrders returns true if the market can accept new orders
func (s OrderBookState) CanAcceptOrders() bool {
	return s == StateRunning
}

// CanCancelOrders returns true if orders can be cancelled
func (s OrderBookState) CanCancelOrders() bool {
	return s == StateRunning || s == StateSuspended
}

// CanAmendOrders returns true if orders can be amended
func (s OrderBookState) CanAmendOrders() bool {
	return s == StateRunning
}
