package scanner

import (
	"fmt"
	"strings"

	"github.com/okneniz/cliche/span"
)

type Captures interface {
	At(int) (span.Interface, bool)
	Append(span.Interface)
	Truncate(size int)
	Empty() bool
	Size() int
	Slice() []span.Interface
	String() string
}

// TODO : add unit tests too

type captures struct {
	spans []span.Interface
}

var _ Captures = new(captures)

func newCaptures(capacity int) *captures {
	return &captures{
		spans: make([]span.Interface, 0, capacity),
	}
}

func (c *captures) Append(s span.Interface) {
	c.spans = append(c.spans, s)
}

func (c *captures) At(idx int) (span.Interface, bool) {
	if idx < 0 || idx >= c.Size() {
		return nil, false
	}

	return c.spans[idx], true
}

func (c *captures) Truncate(size int) {
	if size >= c.Size() {
		return
	}

	// TODO : use truncated list
	c.spans = c.spans[:size]
}

func (c *captures) Empty() bool {
	return len(c.spans) == 0
}

func (c *captures) Size() int {
	return len(c.spans)
}

func (c *captures) String() string {
	ms := make([]string, len(c.spans))
	for i, v := range c.spans {
		ms[i] = v.String()
	}

	return fmt.Sprintf("[%s]", strings.Join(ms, ",\n"))
}

func (c *captures) Slice() []span.Interface {
	s := make([]span.Interface, len(c.spans))
	copy(s, c.spans)
	return s
}
