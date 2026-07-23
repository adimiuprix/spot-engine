package matcher

import (
	"github.com/adimiuprix/spot-engine/event"
	"github.com/adimiuprix/spot-engine/protocol"
)

// CancelResult contains the result of a cancel operation
type CancelResult struct {
	Success bool
	Reason  protocol.RejectReason
	Detail  string
	Log     *event.OrderBookLog
}

// CancelOrder cancels an order in the book
func (m *Matcher) CancelOrder(req *protocol.CancelOrderRequest) CancelResult {
	// Find order in book
	o := m.book.FindOrder(req.OrderID)
	if o == nil {
		// Order not found
		rejectLog := event.NewRejectLog(
			m.seqGen.Next(),
			req.CommandID,
			req.UserID,
			req.Symbol,
			req.Timestamp,
			protocol.RejectReasonOrderNotFound,
			"order not found in book",
			req.OrderID,
		)
		m.publisher.Publish(rejectLog)

		return CancelResult{
			Success: false,
			Reason:  protocol.RejectReasonOrderNotFound,
			Detail:  "order not found in book",
			Log:     rejectLog,
		}
	}

	// Validate ownership
	if o.UserID != req.UserID {
		// User doesn't own this order
		rejectLog := event.NewRejectLog(
			m.seqGen.Next(),
			req.CommandID,
			req.UserID,
			req.Symbol,
			req.Timestamp,
			protocol.RejectReasonInvalidOrderOwner,
			"user does not own this order",
			req.OrderID,
		)
		m.publisher.Publish(rejectLog)

		return CancelResult{
			Success: false,
			Reason:  protocol.RejectReasonInvalidOrderOwner,
			Detail:  "user does not own this order",
			Log:     rejectLog,
		}
	}

	// Remove from book
	removed := m.book.RemoveOrder(o)
	if !removed {
		// Failed to remove (should not happen if FindOrder succeeded)
		rejectLog := event.NewRejectLog(
			m.seqGen.Next(),
			req.CommandID,
			req.UserID,
			req.Symbol,
			req.Timestamp,
			protocol.RejectReasonInternalError,
			"failed to remove order from book",
			req.OrderID,
		)
		m.publisher.Publish(rejectLog)

		return CancelResult{
			Success: false,
			Reason:  protocol.RejectReasonInternalError,
			Detail:  "failed to remove order from book",
			Log:     rejectLog,
		}
	}

	// Emit cancel log
	cancelLog := event.NewCancelLog(
		m.seqGen.Next(),
		req.CommandID,
		req.UserID,
		req.Symbol,
		req.Timestamp,
		req.OrderID,
		sideToString(o.Side),
		o.Price,
		o.Remaining(),
	)
	m.publisher.Publish(cancelLog)

	return CancelResult{
		Success: true,
		Log:     cancelLog,
	}
}
