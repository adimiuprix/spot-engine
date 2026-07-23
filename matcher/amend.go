package matcher

import (
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/order"
	"github.com/adimiuprix/spot-engine/protocol"
	"github.com/shopspring/decimal"
)

// AmendResult contains the result of an amend operation
type AmendResult struct {
	Success bool
	Reason  protocol.RejectReason
	Detail  string
	Trades  []*event.OrderBookLog
}

// ProcessAmend processes an amend order request
func (m *Matcher) ProcessAmend(req *protocol.AmendOrderRequest) AmendResult {
	// Find the order
	existingOrder := m.book.FindOrder(req.OrderID)
	if existingOrder == nil {
		return AmendResult{
			Success: false,
			Reason:  protocol.RejectReasonOrderNotFound,
			Detail:  "order not found in book",
		}
	}

	// Verify ownership
	if existingOrder.UserID != req.UserID {
		return AmendResult{
			Success: false,
			Reason:  protocol.RejectReasonInvalidOrderOwner,
			Detail:  "user does not own this order",
		}
	}

	// Check if already fully filled
	if existingOrder.IsFilled() {
		return AmendResult{
			Success: false,
			Reason:  protocol.RejectReasonOrderNotFound,
			Detail:  "order is fully filled",
		}
	}

	// Calculate changes
	priceChanged := !existingOrder.Price.Equal(req.NewPrice)
	sizeChanged := !existingOrder.Quantity.Equal(req.NewSize)

	// Validate new size
	if req.NewSize.LessThanOrEqual(existingOrder.Filled) {
		return AmendResult{
			Success: false,
			Reason:  protocol.RejectReasonInsufficientSize,
			Detail:  "new size must be greater than filled quantity",
		}
	}

	// Determine if priority is lost
	losePriority := priceChanged || req.NewSize.GreaterThan(existingOrder.Quantity)

	if losePriority {
		// Remove from book, update, and re-match
		return m.amendWithPriorityLoss(existingOrder, req)
	} else {
		// In-place update (size decrease only)
		return m.amendInPlace(existingOrder, req)
	}
}

// amendInPlace updates order in-place (keeps priority)
// Only for: same price + size decrease
func (m *Matcher) amendInPlace(o *order.Order, req *protocol.AmendOrderRequest) AmendResult {
	oldQuantity := o.Quantity
	
	// Update quantity
	o.Quantity = req.NewSize
	
	// Emit amend log as a cancel of the reduced portion
	if req.NewSize.LessThan(oldQuantity) {
		reduced := oldQuantity.Sub(req.NewSize)
		amendLog := event.NewCancelLog(
			m.seqGen.Next(),
			req.CommandID,
			req.UserID,
			req.MarketID,
			req.Timestamp,
			req.OrderID,
			sideToString(o.Side),
			o.Price,
			reduced,
		)
		m.publisher.Publish(amendLog)
	}

	return AmendResult{
		Success: true,
	}
}

// amendWithPriorityLoss removes order, updates, and re-matches as fresh order
// For: price change or size increase
func (m *Matcher) amendWithPriorityLoss(o *order.Order, req *protocol.AmendOrderRequest) AmendResult {
	// Remove from book
	m.book.RemoveOrder(o)

	// Emit cancel log for old order state
	oldRemaining := o.Remaining()
	cancelLog := event.NewCancelLog(
		m.seqGen.Next(),
		o.CommandID,
		o.UserID,
		o.Symbol,
		req.Timestamp,
		o.OrderID,
		sideToString(o.Side),
		o.Price,
		oldRemaining,
	)
	m.publisher.Publish(cancelLog)

	// Update order with new parameters
	o.Price = req.NewPrice
	o.Quantity = req.NewSize
	o.CommandID = req.CommandID // Update to amend command
	o.Timestamp = req.Timestamp // Update timestamp (loses time priority)

	// Re-match as fresh order (might trade immediately)
	result := m.Process(o)

	// If not fully filled, add back to book
	if !o.IsFilled() {
		m.book.Add(o)
	}

	return AmendResult{
		Success: true,
		Trades:  result.Trades,
	}
}
