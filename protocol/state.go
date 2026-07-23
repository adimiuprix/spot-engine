package protocol

// OrderBookState represents the lifecycle state of an order book
type OrderBookState string

const (
	StateRunning   OrderBookState = "running"   // Normal operation
	StateSuspended OrderBookState = "suspended" // Trading suspended, only cancel allowed
	StateHalted    OrderBookState = "halted"    // Emergency halt, no operations allowed
)

func (s OrderBookState) String() string {
	return string(s)
}

// CanPlaceOrder returns true if new orders can be placed in this state
func (s OrderBookState) CanPlaceOrder() bool {
	return s == StateRunning
}

// CanCancelOrder returns true if orders can be cancelled in this state
func (s OrderBookState) CanCancelOrder() bool {
	return s == StateRunning || s == StateSuspended
}

// CanAmendOrder returns true if orders can be amended in this state
func (s OrderBookState) CanAmendOrder() bool {
	return s == StateRunning
}
