package cliche

import (
	"fmt"

	"github.com/okneniz/cliche/span"
)

type Output interface {
	Yield(n Node, s span.Interface, subString string)
	Groups() Captures
	NamedGroups() Captures
	Slice() []*Match
	String() string
	LastPosOf(n Node) (int, bool)
}

type Scanner interface {
	Match(n Node, from, to int, isLeaf, isEmpty bool)
	Position() int
	Rewind(pos int)
	Groups() Captures
	NamedGroups() Captures
}

type Captures interface {
	From(name string, pos int)
	To(name string, pos int)
	Delete(name string)
}

// https://www.regular-expressions.info/engine.html
// This is a very important point to understand:
// a regex engine always returns the leftmost match,
// even if a “better” match could be found later.

type scanner struct {
	input  TextBuffer
	output Output

	// current matched expression
	expression *truncatedList[nodeMatch]
}

var _ Scanner = new(scanner)

func newFullScanner(input TextBuffer, output Output) *scanner {
	s := new(scanner)
	s.input = input
	s.output = output

	// TODO : capacity = height of tree
	s.expression = newTruncatedList[nodeMatch](50)
	return s
}

func (s *scanner) String() string {
	return fmt.Sprintf(
		"Scanner(\n\toutput=%s\n)",
		s.output.String(),
	)
}

func (s *scanner) Position() int {
	return s.expression.len()
}

func (s *scanner) Rewind(pos int) {
	if s.expression.len() < pos {
		return
	}

	s.expression.truncate(pos)
}

func (s *scanner) Groups() Captures {
	return s.output.Groups()
}

func (s *scanner) NamedGroups() Captures {
	return s.output.NamedGroups()
}

func (s *scanner) Match(n Node, from, to int, leaf, empty bool) {
	x := nodeMatch{node: n}

	if empty {
		x.span = span.Empty(from)
	} else {
		x.span = span.New(from, to)
	}

	s.expression.append(x)
	if !leaf {
		return
	}

	sp, exists := s.currentMatchSpan()
	if !exists {
		return
	}

	var (
		subString string
		err       error
	)

	if !sp.Empty() {
		subString, err = s.input.Substring(sp.From(), sp.To())
		if err != nil {
			panic(err)
		}
	}

	s.output.Yield(n, sp, subString)
}

func (s *scanner) currentMatchSpan() (span.Interface, bool) {
	begin, exists := s.firstSpan()
	if !exists {
		return nil, false
	}

	end, exists := s.lastSpan()
	if !exists {
		return nil, false
	}

	if begin.From() >= s.input.Size() {
		// TODO : size - 1 or size?
		return span.Empty(s.input.Size() - 1), true
	}

	// TODO : кажется begin хватит тут
	beginSubstring, exists := s.firstNotEmptySpan()
	if !exists {
		return end, true
	}

	endSubstring, exists := s.lastNotEmptySpan()
	if !exists {
		return end, true
	}

	return span.New(
		beginSubstring.From(),
		endSubstring.To(),
	), true
}

func (s *scanner) firstSpan() (span.Interface, bool) {
	if x, ok := s.expression.first(); ok {
		return x.span, true
	}

	return nil, false
}

func (s *scanner) lastSpan() (span.Interface, bool) {
	if x, ok := s.expression.last(); ok {
		return x.span, true
	}

	return nil, false
}

func (s *scanner) firstNotEmptySpan() (span.Interface, bool) {
	for i := 0; i < s.expression.len(); i++ {
		m := s.expression.at(i)

		if !m.span.Empty() {
			return m.span, true
		}
	}

	return nil, false
}

func (s *scanner) lastNotEmptySpan() (span.Interface, bool) {
	for i := s.expression.len() - 1; i >= 0; i-- {
		m := s.expression.at(i)
		if !m.span.Empty() {
			return m.span, true
		}
	}

	return nil, false
}
