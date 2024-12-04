package cliche

import (
	"fmt"
	"sort"
	"strings"

	"github.com/okneniz/cliche/span"
)

type Output interface {
	Yield(n Node, from, to int, isLeaf, isEmpty bool)
	Matches() []*stringMatch

	Position() int
	Rewind(pos int)

	LastMatchSpan() (span.Interface, bool)
	LastPosOf(n Node) (int, bool)

	AddNamedGroup(name string, pos int)
	MatchNamedGroup(name string, pos int)
	DeleteNamedGroup(name string)

	AddGroup(name string, pos int)
	MatchGroup(name string, pos int)
	DeleteGroup(name string)
}

// https://www.regular-expressions.info/engine.html
// This is a very important point to understand: a regex engine always returns the leftmost match, even if a “better” match could be found later.

type Scanner struct {
	input TextBuffer
	from  int
	to    int

	// current matched expression
	expression *truncatedList[nodeMatch]

	// current captured groups
	groups *captures

	// current captured named groups
	namedGroups *captures

	// output
	matches map[Node]*matchesList
}

var _ Output = new(Scanner)

func newFullScanner(input TextBuffer, from, to int) *Scanner {
	s := new(Scanner)
	s.input = input
	s.from = from
	s.to = to

	// TODO : use node as key for unnamed groups to avoid generate string ID
	s.groups = newCaptures()
	s.namedGroups = newCaptures()

	// TODO : capacity = height of tree
	s.expression = newTruncatedList[nodeMatch](50)
	// TODO : clean matches after scan is really required?
	s.matches = make(map[Node]*matchesList)

	return s
}

func (s *Scanner) String() string {
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

func (s *Scanner) Yield(n Node, from, to int, leaf, empty bool) {
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

func (s *Scanner) lastMatch() (*stringMatch, bool) {
	n, exists := s.expression.last()
	if !exists {
		return nil, false
	}

	begin, exists := s.firstSpan()
	if !exists {
		return nil, false
	}

	m := &stringMatch{
		expressions: newDict().merge(n.node.GetExpressions()),
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

func (s *Scanner) LastPosOf(n Node) (int, bool) {
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

func (s *Scanner) Matches() []*stringMatch {
	size := 0
	for _, v := range s.matches {
		size += v.size()
	}

	result := make([]*stringMatch, 0, size)
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

func (s *Scanner) Position() int {
	return s.expression.len()
}

func (s *Scanner) Rewind(pos int) {
	if s.expression.len() < pos {
		return
	}

	s.expression.truncate(pos)
}

func (s *Scanner) firstSpan() (span.Interface, bool) {
	if x, ok := s.expression.first(); ok {
		return x.span, true
	}

	return nil, false
}

func (s *Scanner) LastMatchSpan() (span.Interface, bool) {
	if x, ok := s.expression.last(); ok {
		return x.span, true
	}

	return nil, false
}

func (s *Scanner) firstNotEmptySpan() (span.Interface, bool) {
	for i := 0; i < s.expression.len(); i++ {
		m := s.expression.at(i)

		if !m.span.Empty() {
			return m.span, true
		}
	}

	return nil, false
}

func (s *Scanner) lastNotEmptySpan() (span.Interface, bool) {
	for i := s.expression.len() - 1; i >= 0; i-- {
		m := s.expression.at(i)
		if !m.span.Empty() {
			return m.span, true
		}
	}

	return nil, false
}

func (s *Scanner) AddNamedGroup(name string, index int) {
	s.namedGroups.From(name, index)
}

func (s *Scanner) MatchNamedGroup(name string, index int) {
	s.namedGroups.To(name, index)
}

func (s *Scanner) DeleteNamedGroup(name string) {
	s.namedGroups.Delete(name)
}

func (s *Scanner) AddGroup(name string, index int) {
	s.groups.From(name, index)
}

func (s *Scanner) MatchGroup(name string, index int) {
	s.groups.To(name, index)
}

func (s *Scanner) DeleteGroup(name string) {
	s.groups.Delete(name)
}

type nodeMatch struct {
	node Node
	span span.Interface
}

func (m nodeMatch) String() string {
	return fmt.Sprintf("nodeMatch{%s: %s}", m.span, m.node.GetKey())
}

type stringMatch struct {
	subString   string
	span        span.Interface
	expressions dict
	groups      []span.Interface
	namedGroups map[string]span.Interface
}

func (m *stringMatch) Span() span.Interface {
	return m.span
}

func (m *stringMatch) Key() string {
	s := m.span.String()
	s += "-"
	s += m.groupsToString()
	s += "-"
	s += m.namedGroupsToString()
	return s
}

func (m *stringMatch) String() string {
	return fmt.Sprintf(
		"stringMatch{%s, '%s', (%s) [%s] {%s}",
		m.span.String(),
		m.subString,
		strings.Join(m.expressions.Slice(), ", "),
		m.groupsToString(),
		m.namedGroupsToString(),
	)
}

func (m *stringMatch) groupsToString() string {
	s := make([]string, len(m.groups))
	for i, x := range m.groups {
		s[i] = x.String()
	}

	sort.SliceStable(s, func(i, j int) bool { return s[i] < s[j] })
	return strings.Join(s, ", ")
}

// TODO : в тестах проверять, что groups входят в span строки

func (m *stringMatch) namedGroupsToString() string {
	pairs := make([]string, 0, len(m.namedGroups))
	for k, v := range m.namedGroups {
		pairs = append(pairs, k+": "+v.String())
	}
	sort.SliceStable(pairs, func(i, j int) bool { return pairs[i] < pairs[j] })
	return strings.Join(pairs, ", ")
}

func (m *stringMatch) Expressions() []string {
	return m.expressions.Slice()
}

func (m *stringMatch) NamedGroups() map[string]span.Interface {
	return m.namedGroups
}

func (m *stringMatch) Groups() []span.Interface {
	return m.groups
}

func (m *stringMatch) Clone() *stringMatch {
	return &stringMatch{
		subString:   m.subString,
		span:        m.span,
		expressions: newDict().merge(m.expressions),
		groups:      m.groups,
		namedGroups: m.namedGroups,
	}
}
