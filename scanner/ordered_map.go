package scanner

import (
	"encoding/json"
)

// TODO : add unit tests

type OrderedMap[K comparable, V any] struct {
	keys   map[K]int
	values []V
	order  []K
}

func newOrderedMap[K comparable, V any](capacity int) *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		keys:   make(map[K]int, capacity),
		values: make([]V, 0, capacity),
		order:  make([]K, 0, capacity),
	}
}

func (c *OrderedMap[K, V]) Get(key K) (V, bool) {
	idx, exists := c.keys[key]
	if !exists {
		var x V
		return x, false
	}

	return c.values[idx], true
}

func (c *OrderedMap[K, V]) Put(key K, s V) {
	_, exists := c.keys[key]
	if exists {
		return
	}

	c.order = append(c.order, key)
	c.values = append(c.values, s)
	c.keys[key] = len(c.values) - 1
}

func (c *OrderedMap[K, V]) Empty() bool {
	return len(c.values) == 0
}

func (c *OrderedMap[K, V]) Size() int {
	return len(c.values)
}

func (c *OrderedMap[K, V]) Truncate(pos int) {
	if pos < 0 || pos >= c.Size() {
		return
	}

	for i := len(c.order) - 1; i >= pos; i-- {
		key := c.order[i]

		if idx, exists := c.keys[key]; exists && idx >= pos {
			delete(c.keys, key)
		}
	}

	// TODO : use truncated list
	c.values = c.values[:pos]
	c.order = c.order[:pos]
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
		x := c.values[v] // TODO : check bounds
		m[k] = x
	}

	return m
}
