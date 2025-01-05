package cliche

import (
	"encoding/json"

	"github.com/okneniz/cliche/span"
)

type NamedCaptures interface {
	Get(string) (span.Interface, bool)
	Put(string, span.Interface)
	Rewind(int)
	Empty() bool
	Size() int
	Map() map[string]span.Interface
	String() string
}

type namedCaptures struct {
	spans []span.Interface
	order []string
	names map[string]int
}

var _ NamedCaptures = new(namedCaptures)

func newNamedCaptures(capacity int) *namedCaptures {
	return &namedCaptures{
		spans: make([]span.Interface, 0, capacity),
		order: make([]string, 0, capacity),
		names: make(map[string]int, capacity),
	}
}

func (c *namedCaptures) Get(name string) (span.Interface, bool) {
	idx, exists := c.names[name]
	if !exists {
		return nil, false
	}

	return c.spans[idx], true
}

func (c *namedCaptures) Put(name string, s span.Interface) {
	_, exists := c.names[name]
	if exists {
		return
	}

	c.order = append(c.order, name)
	c.spans = append(c.spans, s)
	c.names[name] = len(c.spans) - 1
}

func (c *namedCaptures) Empty() bool {
	return len(c.spans) == 0
}

func (c *namedCaptures) Size() int {
	return len(c.spans)
}

func (c *namedCaptures) Rewind(pos int) {
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

func (c *namedCaptures) String() string {
	js, err := json.Marshal(c.names)
	if err != nil {
		return err.Error()
	}

	return string(js)
}

func (c *namedCaptures) Map() map[string]span.Interface {
	m := make(map[string]span.Interface, len(c.names))

	for k, v := range c.names {
		x := c.spans[v] // TODO : check bounds
		m[k] = x
	}

	return m
}
