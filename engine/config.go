package engine

import "github.com/shopspring/decimal"

// DefaultLotSize is the minimum trade unit (1e-8)
// This prevents infinite loops in market order quote mode
var DefaultLotSize = decimal.NewFromFloat(0.00000001) // 0.00000001

type Config struct {
	Symbol string

	RingBufferSize uint64

	// MinLotSize is the minimum executable trade unit
	// Market orders with matchSize < MinLotSize will be rejected
	// Default: 1e-8 (0.00000001)
	MinLotSize decimal.Decimal
}
