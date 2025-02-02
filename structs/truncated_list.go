package structs

type TruncatedList[T any] struct {
	data []T
	size int
}

func NewTruncatedList[T any](cap int) *TruncatedList[T] {
	return &TruncatedList[T]{
		data: make([]T, 0, cap),
		size: 0,
	}
}

func (l *TruncatedList[T]) Append(items ...T) {
	for _, x := range items {
		l.append(x)
	}
}

func (l *TruncatedList[T]) append(item T) {
	if l.size >= len(l.data) {
		l.data = append(l.data, item)
	} else {
		l.data[l.size] = item
	}

	l.size++
}

func (l *TruncatedList[T]) Truncate(newSize int) {
	if newSize < 0 || newSize >= l.size {
		return
	}

	l.size = newSize
}

func (l *TruncatedList[T]) Size() int {
	return l.size
}

func (l *TruncatedList[T]) At(idx int) (T, bool) {
	if idx < 0 || idx >= len(l.data) {
		var zero T
		return zero, false
	}

	return l.data[idx], true
}

func (l *TruncatedList[T]) First() (T, bool) {
	return l.At(0)
}

func (l *TruncatedList[T]) Last() (T, bool) {
	return l.At(l.Size() - 1)
}

func (l *TruncatedList[T]) Slice() []T {
	s := make([]T, l.size)
	for i := 0; i < l.size; i++ {
		s[i] = l.data[i]
	}
	return s
}
