package cliche

import (
	"fmt"
	"strings"

	"github.com/okneniz/cliche/span"
)

type Output interface {
	Yield(n Node, from, to int, isLeaf, isEmpty bool)
	Matches() []*Match

	LastMatchSpan() (span.Interface, bool)
	LastPosOf(n Node) (int, bool)

	Groups() Captures
	NamedGroups() Captures

	Handler
}

type Handler interface {
	Position() int
	Rewind(pos int)
}

type Captures interface {
	From(name string, pos int)
	To(name string, pos int)
	Delete(name string)
}

// https://www.regular-expressions.info/engine.html
// This is a very important point to understand: a regex engine always returns the leftmost match, even if a “better” match could be found later.

type scanner struct {
	input TextBuffer

	// current matched expression
	expression *truncatedList[nodeMatch]

	// current captured groups
	groups *captures

	// current captured named groups
	namedGroups *captures

	// output
	matches map[Node]*matchesList
}

var _ Output = new(scanner)

func newFullScanner(input TextBuffer) *scanner {
	s := new(scanner)
	s.input = input

	// TODO : use node as key for unnamed groups to avoid generate string ID
	s.groups = newCaptures()
	s.namedGroups = newCaptures()

	// TODO : capacity = height of tree
	s.expression = newTruncatedList[nodeMatch](50)
	// TODO : clean matches after scan is really required?
	s.matches = make(map[Node]*matchesList)

	return s
}

func (s *scanner) String() string {
	ms := make([]string, 0, 10)
	for _, v := range s.Matches() {
		ms = append(ms, v.String())
	}

	return fmt.Sprintf(
		"Scanner(\n\texpression=%s,\n\tmatches=%s,\n\tgroups=%s,\n\tnamedGroups=%s\n)",
		s.expression.String(),
		"[\n"+strings.Join(ms, ",\n")+"]",
		s.groups.String(),
		s.namedGroups.String(),
	)
}

func (s *scanner) Yield(n Node, from, to int, leaf, empty bool) {
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

	if strMatch, ok := s.lastMatch(); ok {
		list, exists := s.matches[n]
		if !exists {
			list = newMatchesList()
			s.matches[n] = list
		}

		list.push(strMatch)
	}
}

func (s *scanner) lastMatch() (*Match, bool) {
	n, exists := s.expression.last()
	if !exists {
		return nil, false
	}

	begin, exists := s.firstSpan()
	if !exists {
		return nil, false
	}

	m := &Match{
		expressions: newSet().merge(n.node.GetExpressions()),
		groups:      s.groups.ToSlice(),
		namedGroups: s.namedGroups.ToMap(),
		span:        n.span,
	}

	if begin.From() >= s.input.Size() {
		// TODO : size - 1 or size?
		m.span = span.Empty(s.input.Size() - 1)
		return m, true
	}

	beginSubstring, exists := s.firstNotEmptySpan()
	if !exists {
		return m, true
	}

	endSubstring, exists := s.lastNotEmptySpan()
	if !exists {
		return m, true
	}

	subString, err := s.input.Substring(
		beginSubstring.From(),
		endSubstring.To(),
	)
	if err != nil {
		panic(err)
	}

	m.subString = subString
	m.span = span.New(
		beginSubstring.From(),
		endSubstring.To(),
	)

	return m, true
}

func (s *scanner) LastPosOf(n Node) (int, bool) {
	m, exists := s.matches[n]
	if !exists {
		return -1, false
	}

	match, exists := m.maximum()
	if !exists {
		return -1, false
	}

	return match.Span().To(), true
}

func (s *scanner) Matches() []*Match {
	size := 0
	for _, v := range s.matches {
		size += v.size()
	}

	result := make([]*Match, 0, size)
	c := make(map[string]int)

	for _, list := range s.matches {
		for _, v := range list.Slice() {
			key := v.Key() // remove expressions from keys?

			if idx, exists := c[key]; exists {
				result[idx].expressions.merge(v.expressions)
			} else {
				c[key] = len(result)
				result = append(result, v.Clone()) // TODO : how to remove copy / allocation?
			}
		}
	}

	return result
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

func (s *scanner) firstSpan() (span.Interface, bool) {
	if x, ok := s.expression.first(); ok {
		return x.span, true
	}

	return nil, false
}

func (s *scanner) LastMatchSpan() (span.Interface, bool) {
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

func (s *scanner) Groups() Captures {
	return s.groups
}

func (s *scanner) NamedGroups() Captures {
	return s.namedGroups
}
