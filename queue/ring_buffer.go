package queue

const defaultRingBufferSize = 32768

type RingBuffer[T any] struct {
	buffer []T
	Head   uint64
	Tail   uint64
	size   uint64
}

func New[T any](size uint64) *RingBuffer[T] {
	if size == 0 {
		size = defaultRingBufferSize
	}

	return &RingBuffer[T]{
		buffer: make([]T, size),
		size:   size,
	}
}

func (r *RingBuffer[T]) Push(value T) bool {
	next := (r.Head + 1) % r.size
	if next == r.Tail {
		return false
	}

	r.buffer[r.Head] = value
	r.Head = next
	return true
}

func (r *RingBuffer[T]) Pop() (T, bool) {
	var zero T
	if r.Head == r.Tail {
		return zero, false
	}

	value := r.buffer[r.Tail]
	r.Tail = (r.Tail + 1) % r.size
	return value, true
}

func (r *RingBuffer[T]) Peek() (T, bool) {
	var zero T
	if r.Head == r.Tail {
		return zero, false
	}
	return r.buffer[r.Tail], true
}

func (r *RingBuffer[T]) Size() uint64 {
	if r.Head >= r.Tail {
		return r.Head - r.Tail
	}
	return r.size - r.Tail + r.Head
}

func (r *RingBuffer[T]) IsEmpty() bool {
	return r.Head == r.Tail
}

func (r *RingBuffer[T]) IsFull() bool {
	return (r.Head+1)%r.size == r.Tail
}
