package queue

import "github.com/adimiuprix/spot-engine/order"

type OrderQueue struct {
	rb *RingBuffer[*order.Order]
}

func NewOrderQueue(size uint64) *OrderQueue {
	return &OrderQueue{
		rb: New[*order.Order](size),
	}
}

// Push menambahkan order ke queue (alias dari Push untuk konsistensi)
func (q *OrderQueue) Push(o *order.Order) bool {
	return q.rb.Push(o)
}

// Pop mengambil dan menghapus order pertama dari queue
func (q *OrderQueue) Pop() (*order.Order, bool) {
	return q.rb.Pop()
}

// Front mengintip order pertama tanpa menghapusnya dari queue
func (q *OrderQueue) Front() *order.Order {
	o, ok := q.rb.Peek()
	if !ok {
		return nil
	}
	return o
}

// Remove menghapus order pertama dari queue (alias dari Pop)
func (q *OrderQueue) Remove() (*order.Order, bool) {
	return q.Pop()
}

// IsEmpty mengecek apakah queue kosong
func (q *OrderQueue) IsEmpty() bool {
	return q.rb.IsEmpty()
}

// Size mengembalikan jumlah order dalam queue
func (q *OrderQueue) Size() uint64 {
	return q.rb.Size()
}

// IsFull mengecek apakah queue penuh
func (q *OrderQueue) IsFull() bool {
	return q.rb.IsFull()
}
