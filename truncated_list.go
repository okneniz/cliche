package cliche

import (
	"fmt"
	"strings"
)

type truncatedList[T fmt.Stringer] struct {
	data []T
	size int
}

func newTruncatedList[T fmt.Stringer](cap int) *truncatedList[T] {
	return &truncatedList[T]{
		data: make([]T, 0, cap),
		size: 0,
	}
}

func (l *truncatedList[T]) append(item T) {
	if l.size >= len(l.data) {
		l.data = append(l.data, item)
	} else {
		l.data[l.size] = item
	}

	l.size++
}

func (l *truncatedList[T]) len() int {
	return l.size
}

func (l *truncatedList[T]) truncate(newSize int) {
	if newSize < 0 || newSize > l.size {
		err := OutOfBounds{
			Min:   0,
			Max:   l.size,
			Value: newSize,
		}

		panic(err)
	}

	l.size = newSize
}

func (l *truncatedList[T]) at(idx int) T {
	if idx < 0 || idx >= len(l.data) {
		err := OutOfBounds{
			Min:   0,
			Max:   l.size,
			Value: idx,
		}

		panic(err)
	}

	return l.data[idx]
}

func (l *truncatedList[T]) first() (T, bool) {
	if l.size == 0 {
		var x T
		return x, false
	}

	return l.data[0], true
}

func (l *truncatedList[T]) last() (T, bool) {
	if l.size == 0 {
		var x T
		return x, false
	}

	return l.data[l.size-1], true
}

func (l *truncatedList[T]) toSlice() []T {
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
		items[i] = l.data[i].String()
	}
	return fmt.Sprintf("[%s]", strings.Join(items, ", "))
}
