package cliche

import (
	"fmt"

	"github.com/okneniz/cliche/span"
)

type matchesList struct {
	list []*Match
}

func newMatchesList() *matchesList {
	b := new(matchesList)
	return b
}

func (b *matchesList) compare(m1, m2 span.Interface) int {
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

func (b *matchesList) push(m *Match) {
	if len(b.list) == 0 {
		b.list = append(b.list, m)
		return
	}

	s := m.Span()
	last := b.list[len(b.list)-1]
	lastSpan := last.Span()

	if lastSpan.IsInclude(s.From()) {
		z := b.compare(lastSpan, s)
		if z < 0 {
			b.list[len(b.list)-1] = m
			return
		}

		return
	}

	b.list = append(b.list, m)
}

func (b *matchesList) maximum() (*Match, bool) {
	if len(b.list) == 0 {
		return nil, false
	}

	return b.list[len(b.list)-1], true
}

func (b *matchesList) size() int {
	return len(b.list)
}

func (b matchesList) String() string {
	return fmt.Sprintln(b.list)
}

func (b *matchesList) Slice() []*Match {
	return b.list
}
