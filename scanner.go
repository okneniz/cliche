package cliche

import (
	"fmt"

	"github.com/okneniz/cliche/span"
)

type Scanner interface {
	Match(n Node, from, to int, isLeaf, isEmpty bool)
	Position() int
	Rewind(pos int)

	MatchGroup(from int, to int)
	GroupsPosition() int
	GetGroup(idx int) (span.Interface, bool)
	RewindGroups(pos int)

	MatchNamedGroup(name string, from int, to int)
	NamedGroupsPosition() int
	GetNamedGroup(name string) (span.Interface, bool)
	RewindNamedGroups(pos int)
}

type scanner struct {
	input       Input
	output      Output
	expression  *truncatedList[nodeMatch]
	groups      Captures
	namedGroups NamedCaptures
}

var _ Scanner = new(scanner)

func newFullScanner(
	input Input,
	output Output,
	groups Captures,
	namedGroups NamedCaptures,
) *scanner {
	s := new(scanner)
	s.input = input
	s.output = output
	s.groups = groups
	s.namedGroups = namedGroups

	// TODO : get all attributes by args with interfaces
	// TODO : capacity = height of tree
	s.expression = newTruncatedList[nodeMatch](50)

	return s
}

func (s *scanner) String() string {
	return fmt.Sprintf(
		"Scanner(\n\toutput=%s,\n\tgroups=%s\n)",
		s.output.String(),
		s.groups,
	)
}

func (s *scanner) Position() int {
	return s.expression.len()
}

func (s *scanner) Rewind(pos int) {
	s.expression.truncate(pos)
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

	s.output.Yield(
		n,
		subString,
		sp,
		s.groups.Slice(),
		s.namedGroups.Map(),
	)
}

func (s *scanner) currentMatchSpan() (span.Interface, bool) {
	begin, exists := s.firstSpan()
	if !exists {
		return nil, false
	}

	if begin.From() > s.input.Size() {
		return span.Empty(s.input.Size()), true
	}

	endSubstring, exists := s.lastNotEmptySpan()
	if !exists {
		return begin, true
	}

	return span.New(
		begin.From(),
		endSubstring.To(),
	), true
}

func (s *scanner) firstSpan() (span.Interface, bool) {
	if x, ok := s.expression.first(); ok {
		return x.span, true
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

func (s *scanner) MatchGroup(from int, to int) {
	s.groups.Append(span.New(from, to))
}

func (s *scanner) GroupsPosition() int {
	return s.groups.Size()
}

func (s *scanner) GetGroup(idx int) (span.Interface, bool) {
	// a, b := s.groups.At(idx - 1)
	// fmt.Println("GetGroup", idx, s.groups, a, b)
	return s.groups.At(idx - 1)
}

func (s *scanner) RewindGroups(pos int) {
	s.groups.Truncate(pos)
}

func (s *scanner) MatchNamedGroup(name string, from int, to int) {
	s.namedGroups.Put(name, span.New(from, to))
}

func (s *scanner) NamedGroupsPosition() int {
	return s.namedGroups.Size()
}

func (s *scanner) GetNamedGroup(name string) (span.Interface, bool) {
	return s.namedGroups.Get(name)
}

func (s *scanner) RewindNamedGroups(pos int) {
	s.namedGroups.Rewind(pos)
}
