package protocol

import "github.com/shopspring/decimal"

// QueryType represents the type of query operation
type QueryType uint8

const (
	QueryUnknown  QueryType = 0
	QueryDepth    QueryType = 1 // Get order book depth
	QueryStats    QueryType = 2 // Get market statistics
	QuerySnapshot QueryType = 3 // Take snapshot
)

// Query represents a read-only operation
// Queries do not participate in replay or persistence
type Query struct {
	Type     QueryType `json:"type"`
	MarketID string    `json:"market_id"`
}

// DepthLevel represents a single price level in the order book
type DepthLevel struct {
	Price    decimal.Decimal `json:"price"`
	Quantity decimal.Decimal `json:"quantity"`
	Orders   int             `json:"orders"`
}

// Depth represents the order book depth snapshot
type Depth struct {
	MarketID  string       `json:"market_id"`
	Bids      []DepthLevel `json:"bids"`
	Asks      []DepthLevel `json:"asks"`
	Timestamp int64        `json:"timestamp"`
}

// MarketStats represents market statistics
type MarketStats struct {
	MarketID       string          `json:"market_id"`
	State          OrderBookState  `json:"state"`
	BidCount       int             `json:"bid_count"`
	AskCount       int             `json:"ask_count"`
	BestBid        decimal.Decimal `json:"best_bid"`
	BestAsk        decimal.Decimal `json:"best_ask"`
	LastTradePrice decimal.Decimal `json:"last_trade_price"`
	MinLotSize     decimal.Decimal `json:"min_lot_size"`
	Timestamp      int64           `json:"timestamp"`
}
