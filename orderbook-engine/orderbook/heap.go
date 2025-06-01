package orderbook

type Heap[T comparable] struct {
	data []T
	less func(a, b T) bool
}

func New[T comparable](less func(a, b T) bool) *Heap[T] {
	return &Heap[T]{data: []T{}, less: less}
}

func (h *Heap[T]) Len() int {
	return len(h.data)
}

func (h *Heap[T]) Peek() T {
	if len(h.data) == 0 {
		var zero T
		return zero
	}
	return h.data[0]
}

func (h *Heap[T]) Push(x T) {
	h.data = append(h.data, x)
	h.bubbleUp(len(h.data) - 1)
}

func (h *Heap[T]) Pop() (T, bool) {
	if len(h.data) == 0 {
		var zero T
		return zero, false
	}
	top := h.data[0]
	last := len(h.data) - 1
	h.data[0] = h.data[last]
	h.data = h.data[:last]
	h.bubbleDown(0)
	return top, true
}

func (h *Heap[T]) Remove(x T) (T, bool) {
	index := -1
	for i, val := range h.data {
		if val == x {
			index = i
			break
		}
	}
	if index == -1 {
		var zero T
		return zero, false
	}
	removed := h.data[index]
	last := len(h.data) - 1
	h.swap(index, last)
	h.data = h.data[:last]
	if index < len(h.data) {
		h.bubbleUp(index)
		h.bubbleDown(index)
	}
	return removed, true
}

func (h *Heap[T]) Items() []T {
	return h.data
}

func (h *Heap[T]) bubbleUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if h.less(h.data[i], h.data[parent]) {
			h.swap(i, parent)
			i = parent
		} else {
			break
		}
	}
}

func (h *Heap[T]) bubbleDown(i int) {
	n := len(h.data)
	for {
		left, right := 2*i + 1, 2*i + 2
		smallest := i

		if left < n && h.less(h.data[left], h.data[smallest]) {
			smallest = left
		}
		if right < n && h.less(h.data[right], h.data[smallest]) {
			smallest = right
		}
		if smallest == i {
			break
		}
		h.swap(i, smallest)
		i = smallest
	}
}

func (h *Heap[T]) swap(i, j int) {
	h.data[i], h.data[j] = h.data[j], h.data[i]
}
