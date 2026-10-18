package scanner

import (
	"fmt"

	"github.com/okneniz/cliche/buf"
	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/span"
	"github.com/okneniz/cliche/structs"
)

type (
	FullScanner struct {
		input      node.Input
		output     *Output
		expression *structs.TruncatedList[nodeMatch]
		groups     *structs.OrderedMap[int, Group]
		groupNames *structs.OrderedMap[string, int]
		holes      *structs.TruncatedList[span.Interface]
		roots      map[string]node.Node
		options    ScanOptionsHistory
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
	_ ScanOptionsHistory = structs.NewOrderedMap[node.ScanOption, bool](0)
)

func NewFullScanner(
	input node.Input,
	output *Output,
	roots map[string]node.Node, // add traverse method for tree
	options ...node.ScanOption,
) *FullScanner {
	s := new(FullScanner)
	s.input = input
	s.output = output
	s.roots = roots

	// groups     structs.OrderedMap[int, capture]
	// groupNames structs.OrderedMap[string, int]

	// TODO : capacity = max count of captured groups in expression
	s.groups = structs.NewOrderedMap[int, Group](10) // TODO : rename abstraction?
	s.groupNames = structs.NewOrderedMap[string, int](10)

	// TODO : capacity = height of tree (but what about quantifier)
	s.expression = structs.NewTruncatedList[nodeMatch](50)

	// TODO : capacity = max count of assertions / lookaheads / lookbehins in expression
	s.holes = structs.NewTruncatedList[span.Interface](3)

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
		s.groups.Map(),
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
	s.expression.Truncate(pos)
}

func (s *FullScanner) Scan(from, to int) {
	for _, root := range s.roots {
		nextFrom := from

		for nextFrom <= to {
			lastFrom := nextFrom
			next := span.Pair(nextFrom, to)

			root.Visit(s, s.input, next, func(n node.Node, sp span.Interface) {
				// fmt.Println("before", s.expression)

				// fmt.Println(
				// 	"scanner match",
				// 	fmt.Sprintf("%T", n),
				// 	n.GetKey(), sp,
				// 	n.GetExpressions().Slice(),
				// )

				s.Match(n, sp)
				// fmt.Println("after", s.expression)
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

func (s *FullScanner) Match(n node.Node, sp span.Interface) {
	x := nodeMatch{
		node:   n,
		bounds: sp,
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

	sp = span.Get(sp, s.holes)
	subString := s.getSubString(sp)

	s.output.Yield(
		n,
		subString,
		sp,
		s.groups.Slice(),
	)
}

func (s *FullScanner) getSubString(sp span.Interface) string {
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

func (s *FullScanner) capturedStringSpan() (span.Interface, bool) {
	begin, exists := s.firstSpan()
	if !exists {
		return nil, false
	}

	if begin.From() > s.input.Size() {
		return span.Empty(s.input.Size()), true
	}

	end, exists := s.lastNotEmptySpan()
	if !exists {
		return begin, true
	}

	return span.Pair(
		begin.From(),
		end.To(),
	), true
}

func (s *FullScanner) firstSpan() (span.Interface, bool) {
	if x, ok := s.expression.First(); ok {
		// TODO : skip empty too?
		return x.bounds, true
	}

	return nil, false
}

func (s *FullScanner) lastNotEmptySpan() (span.Interface, bool) {
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

func (s *FullScanner) MatchGroup(idx int, name string, sp span.Interface) {
	g := Group{
		Index: idx,
		Name:  name,
		Span:  span.Get(sp, s.holes),
	}

	s.groups.Put(idx, g)

	if name != "" {
		s.groupNames.Put(name, idx)
	}
}

func (s *FullScanner) GroupsPosition() int {
	return s.groups.Size()
}

func (s *FullScanner) GetGroupByIndex(idx int) (span.Interface, bool) {
	if group, ok := s.groups.Get(idx); ok {
		return group.Span, true
	}

	return nil, false
}

func (s *FullScanner) GetGroupByName(name string) (span.Interface, bool) {
	if idx, ok := s.groupNames.Get(name); ok {
		return s.GetGroupByIndex(idx)
	}

	return nil, false
}

func (s *FullScanner) RewindGroups(pos int) {
	s.groups.Truncate(pos)
	s.groupNames.Truncate(pos)
}

func (s *FullScanner) MarkAsHole(sp span.Interface) {
	s.holes.Append(sp)
}

func (s *FullScanner) HolesPosition() int {
	return s.holes.Size()
}

func (s *FullScanner) RewindHoles(pos int) {
	s.holes.Truncate(pos)
}
