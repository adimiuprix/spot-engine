package queue

const defaultRingBufferSize = 32768

type RingBuffer[T any] struct {
	buffer []T
	head   uint64
	tail   uint64
	size   uint64
}

func New[T any](size uint64) *RingBuffer[T] {

	if size == 0 {

		size = defaultRingBufferSize

	}

	return &RingBuffer[T]{

		buffer: make([]T, size),

		size: size,
	}

}

func (r *RingBuffer[T]) Push(value T) bool {

	next := (r.head + 1) % r.size

	if next == r.tail {

		return false

	}

	r.buffer[r.head] = value

	r.head = next

	return true
}

func (r *RingBuffer[T]) Pop() (T, bool) {

	var zero T

	if r.head == r.tail {

		return zero, false

	}

	value := r.buffer[r.tail]

	r.tail = (r.tail + 1) % r.size

	return value, true
}
