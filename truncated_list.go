package regular

import (
	"fmt"
	"strings"
)

type list[T fmt.Stringer] struct {
	data []T
	size int
}

func newList[T fmt.Stringer](cap int) *list[T] {
	l := new(list[T])
	l.data = make([]T, 0, cap)
	return l
}

func (l *list[T]) append(item T) {
	if l.size >= len(l.data) {
		l.data = append(l.data, item)
	} else {
		l.data[l.size] = item
	}

	l.size++
}

func (l *list[T]) len() int {
	return l.size
}

func (l *list[T]) truncate(newSize int) {
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

func (l *list[T]) at(idx int) T {
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

func (l *list[T]) first() *T {
	if l.size == 0 {
		return nil
	}

	return &l.data[0]
}

func (l *list[T]) last() *T {
	if l.size == 0 {
		return nil
	}

	return &l.data[l.size-1]
}

// TODO : check bounds in the tests
func (l *list[T]) toSlice() []T {
	if l.size == 0 {
		return nil
	}

	return l.data[0:l.size]
}

func (l *list[T]) String() string {
	if l.size == 0 {
		return "[]"
	}

	items := make([]string, l.size)
	for i := 0; i < l.size; i++ {
		items[i] = l.data[i].String()
	}
	return fmt.Sprintf("[%s]", strings.Join(items, ", "))
}
