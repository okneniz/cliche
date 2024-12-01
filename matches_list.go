package cliche

import (
	"fmt"
	"github.com/okneniz/cliche/span"
)

type matchesList struct {
	list []*stringMatch
}

func newMatchesList() *matchesList {
	b := new(matchesList)
	return b
}

// https://www.regular-expressions.info/engine.html
// This is a very important point to understand:
// a regex engine always returns the leftmost match,
// even if a “better” match could be found later.
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

func (b *matchesList) push(m *stringMatch) {
	if len(b.list) == 0 {
		b.list = append(b.list, m)
		return
	}

	s := m.Span()
	last := b.list[len(b.list)-1]
	lastSpan := last.Span()

	if b.include(lastSpan, s.From()) {
		z := b.compare(lastSpan, s)
		if z < 0 {
			b.list[len(b.list)-1] = m
			return
		}

		return
	}

	b.list = append(b.list, m)
}

// TODO : move this methods to span.Interface
func (b *matchesList) include(s span.Interface, x int) bool {
	// TODO:
	//
	// if s.Empty() {
	// 	return false
	// }

	if x < s.From() {
		return false
	}

	if x > s.To() {
		return false
	}

	return true
}

func (b *matchesList) maximum() (*stringMatch, bool) {
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

func (b *matchesList) Slice() []*stringMatch {
	return b.list
}
