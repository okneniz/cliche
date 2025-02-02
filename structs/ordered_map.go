package structs

import (
	"encoding/json"
)

// TODO : add unit tests

type OrderedMap[K comparable, V any] struct {
	keys   map[K]int
	values *TruncatedList[V]
	order  *TruncatedList[K]
}

func NewOrderedMap[K comparable, V any](capacity int) *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		keys:   make(map[K]int, capacity),
		values: NewTruncatedList[V](0),
		order:  NewTruncatedList[K](0),
	}
}

func (c *OrderedMap[K, V]) Get(key K) (V, bool) {
	idx, exists := c.keys[key]
	if !exists {
		var x V
		return x, false
	}

	return c.values.At(idx)
}

// TODO : what about rewriting same key and Truncation?
// test it
// may be store list of indexes in keys?
func (c *OrderedMap[K, V]) Put(key K, s V) {
	_, exists := c.keys[key]
	if exists {
		return
	}

	c.order.Append(key)
	c.values.Append(s)
	c.keys[key] = c.values.Size() - 1
}

func (c *OrderedMap[K, V]) Empty() bool {
	return c.Size() == 0
}

func (c *OrderedMap[K, V]) Size() int {
	return c.values.Size()
}

func (c *OrderedMap[K, V]) Truncate(pos int) {
	if pos < 0 || pos >= c.Size() {
		return
	}

	for i := c.order.Size() - 1; i >= pos; i-- {
		key, _ := c.order.At(i)

		if idx, exists := c.keys[key]; exists && idx >= pos {
			delete(c.keys, key)
		}
	}

	c.values.Truncate(pos)
	c.order.Truncate(pos)
}

func (c *OrderedMap[K, V]) String() string {
	js, err := json.Marshal(c.keys)
	if err != nil {
		return err.Error()
	}

	return string(js)
}

func (c *OrderedMap[K, V]) Map() map[K]V {
	m := make(map[K]V, len(c.keys))

	for k, v := range c.keys {
		if x, ok := c.values.At(v); ok {
			m[k] = x
		}
	}

	return m
}
