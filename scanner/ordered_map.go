package scanner

import (
	"encoding/json"

	"github.com/okneniz/cliche/span"
)

// TODO : add unit tests

type OrderedMap struct {
	spans []span.Interface
	order []string
	names map[string]int
}

func newOrderedMap(capacity int) *OrderedMap {
	return &OrderedMap{
		spans: make([]span.Interface, 0, capacity),
		order: make([]string, 0, capacity),
		names: make(map[string]int, capacity),
	}
}

func (c *OrderedMap) Get(name string) (span.Interface, bool) {
	idx, exists := c.names[name]
	if !exists {
		return nil, false
	}

	return c.spans[idx], true
}

func (c *OrderedMap) Put(name string, s span.Interface) {
	_, exists := c.names[name]
	if exists {
		return
	}

	c.order = append(c.order, name)
	c.spans = append(c.spans, s)
	c.names[name] = len(c.spans) - 1
}

func (c *OrderedMap) Empty() bool {
	return len(c.spans) == 0
}

func (c *OrderedMap) Size() int {
	return len(c.spans)
}

// rename to truncate
func (c *OrderedMap) Truncate(pos int) {
	if pos < 0 || pos >= c.Size() {
		return
	}

	for i := len(c.order) - 1; i >= pos; i-- {
		name := c.order[i]

		if idx, exists := c.names[name]; exists && idx >= pos {
			delete(c.names, name)
		}
	}

	// TODO : use truncated list
	c.spans = c.spans[:pos]
	c.order = c.order[:pos]
}

func (c *OrderedMap) String() string {
	js, err := json.Marshal(c.names)
	if err != nil {
		return err.Error()
	}

	return string(js)
}

func (c *OrderedMap) Map() map[string]span.Interface {
	m := make(map[string]span.Interface, len(c.names))

	for k, v := range c.names {
		x := c.spans[v] // TODO : check bounds
		m[k] = x
	}

	return m
}
