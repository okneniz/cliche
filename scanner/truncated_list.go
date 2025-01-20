package scanner

import (
	"fmt"
	"strings"
)

type TruncatedList[T any] interface {
	Append(T)
	At(int) (T, bool)
	First() (T, bool)
	Last() (T, bool)
	Size() int
	Truncate(int)
	Slice() []T
	String() string
}

type truncatedList[T any] struct {
	data []T
	size int
}

var _ TruncatedList[string] = new(truncatedList[string])

func newTruncatedList[T fmt.Stringer](cap int) *truncatedList[T] {
	return &truncatedList[T]{
		data: make([]T, 0, cap),
		size: 0,
	}
}

func (l *truncatedList[T]) Append(item T) {
	if l.size >= len(l.data) {
		l.data = append(l.data, item)
	} else {
		l.data[l.size] = item
	}

	l.size++
}

func (l *truncatedList[T]) At(idx int) (T, bool) {
	if idx < 0 || idx >= len(l.data) {
		var zero T
		return zero, false
	}

	return l.data[idx], true
}

func (l *truncatedList[T]) First() (T, bool) {
	if l.size == 0 {
		var x T
		return x, false
	}

	return l.data[0], true
}

func (l *truncatedList[T]) Last() (T, bool) {
	if l.size == 0 {
		var x T
		return x, false
	}

	return l.data[l.size-1], true
}

func (l *truncatedList[T]) Size() int {
	return l.size
}

func (l *truncatedList[T]) Truncate(newSize int) {
	if newSize < 0 || newSize > l.size {
		return
	}

	l.size = newSize
}

func (l *truncatedList[T]) Slice() []T {
	if l.size == 0 {
		return nil
	}

	return l.data[0:l.size]
}

func (l *truncatedList[T]) String() string {
	if l.size == 0 {
		return "[]"
	}

	items := make([]string, l.size) // TODO : use buffer instead
	for i := 0; i < l.size; i++ {
		items[i] = fmt.Sprintf("%v", l.data[i])
	}

	return fmt.Sprintf("[%s]", strings.Join(items, ", "))
}
