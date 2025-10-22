package scanner

import (
	"fmt"

	"github.com/okneniz/cliche/buf"
	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/quantity"
	"github.com/okneniz/cliche/structs"
)

type (
	FullScanner struct {
		input       node.Input
		output      node.Output
		expression  *structs.TruncatedList[nodeMatch]
		groups      Captures
		namedGroups NamedCaptures
		holes       *structs.TruncatedList[quantity.Interface]
		roots       map[string]node.Node
		options     ScanOptionsHistory
	}

	Captures interface {
		Append(...quantity.Interface)
		Truncate(int)
		Size() int
		At(int) (quantity.Interface, bool)
		First() (quantity.Interface, bool)
		Last() (quantity.Interface, bool)
		Slice() []quantity.Interface
	}

	NamedCaptures interface {
		Get(string) (quantity.Interface, bool)
		Put(string, quantity.Interface)
		Truncate(int)
		Empty() bool
		Size() int
		Map() map[string]quantity.Interface
		String() string // TODO : remove and use map when it needed
	}

	ScanOptionsHistory interface {
		Get(node.ScanOption) (bool, bool)
		Put(node.ScanOption, bool)
		Truncate(int)
		Empty() bool
		Size() int
		Map() map[node.ScanOption]bool
		String() string // TODO : remove and use map when it needed
	}
)

var (
	_ node.Scanner       = new(FullScanner)
	_ node.Input         = buf.NewRunesBuffer("")
	_ node.Output        = NewOutput()
	_ Captures           = structs.NewTruncatedList[quantity.Interface](0)
	_ NamedCaptures      = structs.NewOrderedMap[string, quantity.Interface](0)
	_ ScanOptionsHistory = structs.NewOrderedMap[node.ScanOption, bool](0)
)

func NewFullScanner(
	input node.Input,
	output node.Output,
	roots map[string]node.Node, // add traverse method for tree
	options ...node.ScanOption,
) *FullScanner {
	s := new(FullScanner)
	s.input = input
	s.output = output
	s.roots = roots

	// TODO : capacity = max count of captured groups in expression
	s.groups = structs.NewTruncatedList[quantity.Interface](10)
	s.namedGroups = structs.NewOrderedMap[string, quantity.Interface](10)

	// TODO : capacity = height of tree (but what about quantifier)
	s.expression = structs.NewTruncatedList[nodeMatch](50)

	// TODO : capacity = max count of assertions / lookaheads / lookbehins in expression
	s.holes = structs.NewTruncatedList[quantity.Interface](3)

	// TODO : what about default options?

	s.options = structs.NewOrderedMap[node.ScanOption, bool](10)
	for _, x := range options {
		s.options.Put(x, true)
	}

	return s
}

func (s *FullScanner) String() string {
	return fmt.Sprintf(
		"Scanner(\n\toutput=%s,\n\texpression=%v,\n\tgroups=%v,\n\tholes=%v\n)",
		s.output.String(),
		s.expression.Slice(),
		s.groups.Slice(),
		s.holes.Slice(),
	)
}

func (s *FullScanner) OptionsInclude(opt node.ScanOption) bool {
	value, exists := s.options.Get(opt)
	if !exists {
		return false
	}

	return value
}

func (s *FullScanner) OptionsEnable(opt node.ScanOption) {
	s.options.Put(opt, true)
}

func (s *FullScanner) OptionsDisable(opt node.ScanOption) {
	s.options.Put(opt, false)
}

func (s *FullScanner) OptionsPosition() int {
	return s.options.Size()
}

func (s *FullScanner) RewindOptions(pos int) {
	s.options.Truncate(pos)
}

func (s *FullScanner) Position() int {
	return s.expression.Size()
}

func (s *FullScanner) Rewind(pos int) {
	// fmt.Println("rewind", pos)
	s.expression.Truncate(pos)
}

func (s *FullScanner) Scan(from, to int) {
	for _, root := range s.roots {
		nextFrom := from

		for nextFrom <= to {
			lastFrom := nextFrom
			root.Visit(s, s.input, nextFrom, to, func(n node.Node, from, to int, empty bool) {
				// fmt.Println("before", s)
				// fmt.Println("scanner match", fmt.Sprintf("%T", n), n.GetKey(), from, to, empty, n.GetExpressions().Slice())
				s.Match(n, from, to, empty)
				// fmt.Println("after", s)
				// fmt.Println("")
			})

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

func (s *FullScanner) Match(n node.Node, from, to int, empty bool) {
	x := nodeMatch{node: n}

	if empty {
		x.bounds = quantity.Empty(from)
	} else {
		x.bounds = quantity.Pair(from, to)
	}

	s.expression.Append(x)
	if !n.IsLeaf() {
		return
	}

	// check lastHole and collision in (lastNotEmptySpan method)
	sp, exists := s.capturedStringSpan()
	if !exists {
		return
	}

	sp = quantity.Get(sp, s.holes)
	subString := s.getSubString(sp)

	s.output.Yield(
		n,
		subString,
		sp,
		s.groups.Slice(),
		s.namedGroups.Map(),
	)
}

func (s *FullScanner) getSubString(sp quantity.Interface) string {
	if sp.Empty() {
		return ""
	}

	if sp.From() >= s.input.Size() || sp.To() >= s.input.Size() {
		// or panic?
		return ""
	}

	size := sp.Size()
	subString := make([]rune, 0, size)

	for idx := sp.From(); idx <= sp.To(); idx++ {
		r := s.input.ReadAt(idx)
		subString = append(subString, r)
	}

	return string(subString)
}

func (s *FullScanner) capturedStringSpan() (quantity.Interface, bool) {
	begin, exists := s.firstSpan()
	if !exists {
		return nil, false
	}

	if begin.From() > s.input.Size() {
		return quantity.Empty(s.input.Size()), true
	}

	end, exists := s.lastNotEmptySpan()
	if !exists {
		return begin, true
	}

	return quantity.Pair(
		begin.From(),
		end.To(),
	), true
}

func (s *FullScanner) firstSpan() (quantity.Interface, bool) {
	if x, ok := s.expression.First(); ok {
		// TODO : skip empty too?
		return x.bounds, true
	}

	return nil, false
}

func (s *FullScanner) lastNotEmptySpan() (quantity.Interface, bool) {
	for i := s.expression.Size() - 1; i >= 0; i-- {
		m, ok := s.expression.At(i)
		if !ok {
			return nil, false
		}

		if !m.bounds.Empty() {
			return m.bounds, true
		}
	}

	return nil, false
}

func (s *FullScanner) MatchGroup(from int, to int) {
	g := quantity.Get(quantity.Pair(from, to), s.holes)
	s.groups.Append(g)
}

func (s *FullScanner) GroupsPosition() int {
	return s.groups.Size()
}

func (s *FullScanner) GetGroup(idx int) (quantity.Interface, bool) {
	return s.groups.At(idx - 1)
}

func (s *FullScanner) RewindGroups(pos int) {
	s.groups.Truncate(pos)
}

func (s *FullScanner) MatchNamedGroup(name string, from int, to int) {
	g := quantity.Get(quantity.Pair(from, to), s.holes)
	s.namedGroups.Put(name, g)
}

func (s *FullScanner) NamedGroupsPosition() int {
	return s.namedGroups.Size()
}

func (s *FullScanner) GetNamedGroup(name string) (quantity.Interface, bool) {
	return s.namedGroups.Get(name)
}

func (s *FullScanner) RewindNamedGroups(pos int) {
	s.namedGroups.Truncate(pos)
}

func (s *FullScanner) MarkAsHole(from int, to int) {
	s.holes.Append(quantity.Pair(from, to))
}

func (s *FullScanner) HolesPosition() int {
	return s.holes.Size()
}

func (s *FullScanner) RewindHoles(pos int) {
	s.holes.Truncate(pos)
}
