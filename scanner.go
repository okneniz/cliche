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

	MarkAsHole(from int, to int)
	HolesPosition() int
	RewindHoles(pos int)
}

type scanner struct {
	input       Input
	output      Output
	expression  TruncatedList[nodeMatch]
	groups      Captures
	namedGroups NamedCaptures
	holes       TruncatedList[span.Interface]
}

var _ Scanner = new(scanner)

func newFullScanner(
	input Input,
	output Output,
	groups Captures, // or hide it?
	namedGroups NamedCaptures, // or hide it?
) *scanner {
	s := new(scanner)
	s.input = input
	s.output = output
	s.groups = groups
	s.namedGroups = namedGroups

	// TODO : capacity = height of tree
	s.expression = newTruncatedList[nodeMatch](50)

	// TODO : capacity = max count of assertions / lookaheads / lookbehins in expression
	s.holes = newTruncatedList[span.Interface](3)

	return s
}

func (s *scanner) String() string {
	return fmt.Sprintf(
		"Scanner(\n\toutput=%s,\n\tgroups=%s,\n\tholes=%s\n)",
		s.output.String(),
		s.groups,
		s.holes.String(),
	)
}

func (s *scanner) Position() int {
	return s.expression.Size()
}

func (s *scanner) Rewind(pos int) {
	s.expression.Truncate(pos)
}

func (s *scanner) Match(n Node, from, to int, leaf, empty bool) {
	x := nodeMatch{node: n}

	if empty {
		x.span = span.Empty(from)
	} else {
		x.span = span.New(from, to)
	}

	s.expression.Append(x)
	if !leaf {
		return
	}

	sp, exists := s.currentMatchSpan() // check lastHole and collision in (lastNotEmptySpan method)
	if !exists {
		return
	}

	groups := s.groups.Slice()
	namedGroups := s.namedGroups.Map()

	// what about empty last hole, just skip empty holes?
	if lastHole, ok := s.holes.Last(); ok {
		if sp.To() == lastHole.To() {
			sp = span.New(sp.From(), lastHole.From()-1)
		}

		for i, _ := range groups {
			if groups[i].To() == lastHole.To() && groups[i].From() != lastHole.From() {
				groups[i] = span.New(groups[i].From(), lastHole.From()-1)
				break
			}
		}

		for k, v := range namedGroups {
			if v.To() == lastHole.To() && v.From() != lastHole.From() {
				namedGroups[k] = span.New(v.From(), lastHole.From()-1)
				break
			}
		}

	}

	s.output.Yield(
		n,
		s.getSubString(sp),
		sp,
		groups,
		namedGroups,
	)
}

func (s *scanner) getSubString(sp span.Interface) string {
	if sp.Empty() {
		return ""
	}

	if sp.From() >= s.input.Size() || sp.To() >= s.input.Size() {
		return "" // or panic?
	}

	// if s.holes.Size() > 0 {
	// 	return s.getSubStringWithHoles(sp)
	// }

	size := sp.Size()
	subString := make([]rune, 0, size)

	for idx := sp.From(); idx <= sp.To(); idx++ {
		r := s.input.ReadAt(idx)
		subString = append(subString, r)
	}

	return string(subString)
}

// func (s *scanner) getSubStringWithHoles(sp span.Interface) string {
// 	size := sp.Size() // remove holes to better allocations?
// 	subString := make([]rune, 0, size)

// 	holeIdx := 0
// 	hole := s.nextHole(holeIdx)

// 	for idx := sp.From(); idx <= sp.To(); idx++ {
// 		for hole.To() < idx {
// 			holeIdx++
// 			hole = s.nextHole(holeIdx)
// 		}

// 		if hole.Include(idx) {
// 			continue
// 		}

// 		r := s.input.ReadAt(idx)
// 		subString = append(subString, r)
// 	}

// 	return string(subString)
// }

// func (s *scanner) nextHole(idx int) span.Interface {
// 	hole, ok := s.holes.At(idx)
// 	if !ok {
// 		hole = span.Empty(s.input.Size() + 1)
// 	}

// 	return hole
// }

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
	if x, ok := s.expression.First(); ok {
		// TODO : skip empty too?
		return x.span, true
	}

	return nil, false
}

func (s *scanner) lastNotEmptySpan() (span.Interface, bool) {
	for i := s.expression.Size() - 1; i >= 0; i-- {
		m, ok := s.expression.At(i)
		if !ok {
			return nil, false
		}

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

func (s *scanner) MarkAsHole(from int, to int) {
	s.holes.Append(span.New(from, to))
}

func (s *scanner) HolesPosition() int {
	return s.holes.Size()
}

func (s *scanner) RewindHoles(pos int) {
	s.holes.Truncate(pos)
}
