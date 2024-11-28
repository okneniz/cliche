package span

import (
	"fmt"
)

// TODO : move to span package?

type Interface interface {
	From() int
	To() int
	Empty() bool
	Size() int
	String() string
}

type span struct {
	from int
	to   int
}

var (
	_ Interface = span{0, 0}
)

func New(from int, to int) Interface {
	if from > to {
		panic(fmt.Sprintf("invalid SPAN %d %d", from, to))
	}
	return span{
		from: from,
		to:   to,
	}
}

func (m span) From() int {
	return m.from
}

func (m span) To() int {
	return m.to
}

// required for empty matches like .? or .*
func (m span) Empty() bool {
	return false
}

func (m span) Size() int {
	return m.to - m.from
}

func (m span) String() string {
	return fmt.Sprintf("[%d-%d]", m.From(), m.To())
}

type empty int

var (
	_ Interface = empty(0)
)

func Empty(pos int) Interface {
	return empty(pos)
}

func (m empty) From() int {
	return int(m)
}

func (m empty) To() int {
	return int(m)
}

// required for empty matches like .? or .*
func (m empty) Empty() bool {
	return true
}

func (m empty) Size() int {
	return 0
}

func (m empty) String() string {
	return fmt.Sprintf("(%d)", m.From())
}
