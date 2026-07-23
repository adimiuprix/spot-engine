package queue

import "github.com/adimiuprix/spot-engine/order"

type OrderQueue struct {
	items []*order.Order
}

func (q *OrderQueue) Push(
	o *order.Order,
) {

	q.items =
		append(q.items, o)

}

func (q *OrderQueue) Front() *order.Order {

	if len(q.items) == 0 {
		return nil
	}

	return q.items[0]
}
