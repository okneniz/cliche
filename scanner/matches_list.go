package scanner

import (
	"fmt"

	"github.com/okneniz/cliche/quantity"
)

// TODO : add unit tests too

// https://www.regular-expressions.info/engine.html
// This is a very important point to understand:
// a regex engine always returns the leftmost match,
// even if a “better” match could be found later.

// https://www.regular-expressions.info/posix.html

type matchesList struct {
	list []*Match
}

func newMatchesList() *matchesList {
	b := new(matchesList)
	return b
}

func (b *matchesList) compare(m1, m2 quantity.Interface) int {
	switch {
	case m1.From() > m2.From():
		return -1
	case m1.From() < m2.From():
		return 1
	default:
		switch {
		case m1.Size() > m2.Size():
			return 1
		case m1.Size() < m2.Size():
			return -1
		default:
			return 0
		}
	}
}

func (b *matchesList) Push(m *Match) {
	if len(b.list) == 0 {
		b.list = append(b.list, m)
		return
	}

	s := m.Span()
	last := b.list[len(b.list)-1]
	lastSpan := last.Span()

	if lastSpan.Include(s.From()) {
		z := b.compare(lastSpan, s)
		if z < 0 {
			b.list[len(b.list)-1] = m
			return
		}

		return
	}

	b.list = append(b.list, m)
}

func (b *matchesList) Maximum() (*Match, bool) {
	if len(b.list) == 0 {
		return nil, false
	}

	return b.list[len(b.list)-1], true
}

func (b *matchesList) Size() int {
	return len(b.list)
}

func (b matchesList) String() string {
	return fmt.Sprintln(b.list)
}

func (b *matchesList) Slice() []*Match {
	return b.list
}
