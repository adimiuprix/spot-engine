package matcher

import (
	"github.com/adimiuprix/spot-engine/book"
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/shopspring/decimal"
)

type Matcher struct {
	book      *book.OrderBook
	tradeID   uint64
	seqGen    *event.SequenceGenerator
	publisher event.PublishLog
	lotSize   decimal.Decimal // Minimum trade unit
}

type Result struct {
	Trades []*event.OrderBookLog
	Fills  []*event.OrderBookLog
}

func New(book *book.OrderBook, seqGen *event.SequenceGenerator, publisher event.PublishLog) *Matcher {
	return &Matcher{
		book:      book,
		tradeID:   1,
		seqGen:    seqGen,
		publisher: publisher,
		lotSize:   decimal.NewFromFloat(0.00000001), // Default 1e-8
	}
}

// SetLotSize sets the minimum trade unit
func (m *Matcher) SetLotSize(size decimal.Decimal) {
	m.lotSize = size
}

// Process routes an order to the appropriate handler based on order type
// This is the legacy entry point - new code should use ProcessWithTIF()
func (matcher *Matcher) Process(o *order.Order) Result {
	// Route based on order TYPE (Limit vs Market)
	switch o.Type {
	case order.Limit:
		return matcher.processLimit(o)
	case order.Market:
		return matcher.processMarket(o)
	default:
		// Default to limit order behavior
		return matcher.processLimit(o)
	}
}

// processReplenishments processes iceberg order replenishments
func (matcher *Matcher) processReplenishments(side order.Side) {
	tree := matcher.book.BidTree
	if side == order.Sell {
		tree = matcher.book.AskTree
	}

	// Iterate through all price levels
	tree.Ascend(func(level *book.PriceLevel) bool {
		// Process replenishments (moves replenished orders to tail)
		replenished := level.ProcessReplenishments()

		// Emit replenishment logs
		for _, o := range replenished {
			replenishLog := event.NewFillLog(
				matcher.seqGen.Next(),
				o.CommandID,
				o.UserID,
				o.Symbol,
				o.Timestamp,
				o.OrderID,
				sideToString(o.Side),
				o.Price,
				o.Filled,
				o.Remaining(),
				false, // Not fully filled since it replenished
			)
			matcher.publisher.Publish(replenishLog)
		}

		return true // Continue iteration
	})
}

func sideToString(side order.Side) string {
	if side == order.Buy {
		return "buy"
	}
	return "sell"
}

// GetTradeID returns the current trade ID
func (m *Matcher) GetTradeID() uint64 {
	return m.tradeID
}

// SetTradeID sets the trade ID (used during snapshot restore)
func (m *Matcher) SetTradeID(id uint64) {
	m.tradeID = id
}

func (m *Matcher) execute(buy *order.Order, sell *order.Order) []*event.OrderBookLog {
	logs := make([]*event.OrderBookLog, 0, 3) // trade + 2 fills

	buyRemaining := buy.Remaining()
	sellRemaining := sell.Remaining()

	// Determine trade quantity (minimum of both remainings)
	var qty decimal.Decimal
	if buyRemaining.LessThan(sellRemaining) {
		qty = buyRemaining
	} else {
		qty = sellRemaining
	}

	// Price is always the resting order's price (maker price)
	price := sell.Price

	// Update filled quantities
	buy.Filled = buy.Filled.Add(qty)
	sell.Filled = sell.Filled.Add(qty)

	// Use taker's timestamp (the incoming order)
	timestamp := buy.Timestamp

	// Create trade log
	tradeLog := event.NewTradeLog(
		m.seqGen.Next(),
		buy.CommandID,
		buy.UserID,
		buy.Symbol,
		timestamp,
		m.tradeID,
		sell.OrderID, // Maker (resting order)
		buy.OrderID,  // Taker (incoming order)
		price,
		qty,
		"buy", // Taker side
	)
	logs = append(logs, tradeLog)
	m.publisher.Publish(tradeLog)

	// Create fill log for maker (sell side)
	makerFillLog := event.NewFillLog(
		m.seqGen.Next(),
		sell.CommandID,
		sell.UserID,
		sell.Symbol,
		timestamp,
		sell.OrderID,
		"sell",
		price,
		sell.Filled,
		sell.Remaining(),
		sell.IsFilled(),
	)
	logs = append(logs, makerFillLog)
	m.publisher.Publish(makerFillLog)

	// Create fill log for taker (buy side)
	takerFillLog := event.NewFillLog(
		m.seqGen.Next(),
		buy.CommandID,
		buy.UserID,
		buy.Symbol,
		timestamp,
		buy.OrderID,
		"buy",
		price,
		buy.Filled,
		buy.Remaining(),
		buy.IsFilled(),
	)
	logs = append(logs, takerFillLog)
	m.publisher.Publish(takerFillLog)

	m.tradeID++

	return logs
}
