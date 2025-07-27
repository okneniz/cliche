package structs

import (
	"encoding/json"
)

type OrderedMap[K comparable, V any] struct {
	data  map[K]*TruncatedList[V]
	order *TruncatedList[K]
}

func NewOrderedMap[K comparable, V any](capacity int) *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		data:  make(map[K]*TruncatedList[V], capacity),
		order: NewTruncatedList[K](0),
	}
}

func (c *OrderedMap[K, V]) Get(key K) (V, bool) {
	values, exists := c.data[key]
	if !exists {
		var x V
		return x, false
	}

	return values.Last()
}

// TODO :
// RUBY : /(?<test>.)(?<test>.)(?<test>.)/.match("123").named_captures => {"test"=>"3"}

func (c *OrderedMap[K, V]) Put(key K, s V) {
	values, exists := c.data[key]
	if !exists {
		values = NewTruncatedList[V](1)
		c.data[key] = values
	}

	c.order.Append(key)
	values.Append(s)
}

func (c *OrderedMap[K, V]) Empty() bool {
	return c.Size() == 0
}

func (c *OrderedMap[K, V]) Size() int {
	return c.order.Size()
}

func (c *OrderedMap[K, V]) Truncate(pos int) {
	if pos < 0 || pos >= c.Size() {
		return
	}

	for i := c.order.Size() - 1; i >= pos; i-- {
		key, _ := c.order.At(i)

		values, exists := c.data[key]
		if !exists {
			continue
		}

		values.Truncate(values.Size() - 1)
	}

	c.order.Truncate(pos)
}

func (c *OrderedMap[K, V]) String() string {
	js, err := json.Marshal(c.Map())
	if err != nil {
		return err.Error()
	}

	return string(js)
}

func (c *OrderedMap[K, V]) Map() map[K]V {
	m := make(map[K]V, len(c.data))

	for key, values := range c.data {
		value, ok := values.Last()
		if ok {
			m[key] = value
		}
	}

	return m
}
