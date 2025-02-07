package scanner

import (
	"fmt"

	"github.com/okneniz/cliche/buf"
	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/span"
	"github.com/okneniz/cliche/structs"
)

type FullScanner struct {
	input       node.Input
	output      node.Output
	expression  *structs.TruncatedList[nodeMatch]
	groups      Captures
	namedGroups NamedCaptures
	holes       *structs.TruncatedList[span.Interface]
	roots       map[string]node.Node
}

type Captures interface {
	Append(...span.Interface)
	Truncate(int)
	Size() int
	At(int) (span.Interface, bool)
	First() (span.Interface, bool)
	Last() (span.Interface, bool)
	Slice() []span.Interface
}

type NamedCaptures interface {
	Get(string) (span.Interface, bool)
	Put(string, span.Interface)
	Truncate(int)
	Empty() bool
	Size() int
	Map() map[string]span.Interface
	String() string // TODO : remove and use map when it needed
}

var (
	_ node.Scanner  = new(FullScanner)
	_ node.Input    = buf.NewRunesBuffer("")
	_ node.Output   = NewOutput()
	_ Captures      = structs.NewTruncatedList[span.Interface](0)
	_ NamedCaptures = structs.NewOrderedMap[string, span.Interface](0)
)

func NewFullScanner(
	input node.Input,
	output node.Output,
	roots map[string]node.Node, // add traverse method for tree
) *FullScanner {
	s := new(FullScanner)
	s.input = input
	s.output = output
	s.roots = roots

	// TODO : capacity = max count of captured groups in expression
	s.groups = structs.NewTruncatedList[span.Interface](10)
	s.namedGroups = structs.NewOrderedMap[string, span.Interface](10)

	// TODO : capacity = height of tree (but what about quantifier)
	s.expression = structs.NewTruncatedList[nodeMatch](50)

	// TODO : capacity = max count of assertions / lookaheads / lookbehins in expression
	s.holes = structs.NewTruncatedList[span.Interface](3)

	return s
}

func (s *FullScanner) String() string {
	return fmt.Sprintf(
		"Scanner(\n\toutput=%s,\n\tgroups=%s,\n\tholes=%v\n)",
		s.output.String(),
		s.groups,
		s.holes,
	)
}

func (s *FullScanner) Position() int {
	return s.expression.Size()
}

func (s *FullScanner) Rewind(pos int) {
	s.expression.Truncate(pos)
}

func (s *FullScanner) Scan(from, to int) {
	skip := func(_ node.Node, _, _ int, _ bool) {}

	// TODO : rewrite to traverse
	for _, root := range s.roots {
		nextFrom := from

		for nextFrom <= to {
			lastFrom := nextFrom
			root.Visit(s, s.input, nextFrom, to, skip)

			if pos, ok := s.output.LastPosOf(root); ok && pos >= nextFrom {
				nextFrom = pos
			}

			if lastFrom == nextFrom {
				nextFrom++
			}

			s.Rewind(0)
		}
	}
}

func (s *FullScanner) Match(n node.Node, from, to int, leaf, empty bool) {
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

	// check lastHole and collision in (lastNotEmptySpan method)
	sp, exists := s.currentMatchSpan()
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

		for i := range groups {
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

func (s *FullScanner) getSubString(sp span.Interface) string {
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

// func (s *FullScanner) getSubStringWithHoles(sp span.Interface) string {
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

// func (s *FullScanner) nextHole(idx int) span.Interface {
// 	hole, ok := s.holes.At(idx)
// 	if !ok {
// 		hole = span.Empty(s.input.Size() + 1)
// 	}

// 	return hole
// }

func (s *FullScanner) currentMatchSpan() (span.Interface, bool) {
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

func (s *FullScanner) firstSpan() (span.Interface, bool) {
	if x, ok := s.expression.First(); ok {
		// TODO : skip empty too?
		return x.span, true
	}

	return nil, false
}

func (s *FullScanner) lastNotEmptySpan() (span.Interface, bool) {
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

func (s *FullScanner) MatchGroup(from int, to int) {
	s.groups.Append(span.New(from, to))
}

func (s *FullScanner) GroupsPosition() int {
	return s.groups.Size()
}

func (s *FullScanner) GetGroup(idx int) (span.Interface, bool) {
	return s.groups.At(idx - 1)
}

func (s *FullScanner) RewindGroups(pos int) {
	s.groups.Truncate(pos)
}

func (s *FullScanner) MatchNamedGroup(name string, from int, to int) {
	s.namedGroups.Put(name, span.New(from, to))
}

func (s *FullScanner) NamedGroupsPosition() int {
	return s.namedGroups.Size()
}

func (s *FullScanner) GetNamedGroup(name string) (span.Interface, bool) {
	return s.namedGroups.Get(name)
}

func (s *FullScanner) RewindNamedGroups(pos int) {
	s.namedGroups.Truncate(pos)
}

func (s *FullScanner) MarkAsHole(from int, to int) {
	s.holes.Append(span.New(from, to))
}

func (s *FullScanner) HolesPosition() int {
	return s.holes.Size()
}

func (s *FullScanner) RewindHoles(pos int) {
	s.holes.Truncate(pos)
}
