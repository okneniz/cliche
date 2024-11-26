package regular

import (
	"fmt"
	"sort"
	"strings"
)

type Handler interface {
	Match(n node, from, to int, isLeaf, isEmpty bool)
	Matches() []*stringMatch

	// TODO : how to remove it?
	// required only for quantifier
	//
	// maybe use LastPosOf in quantifier?
	LastMatch() *nodeMatch

	Position() int
	Rewind(pos int)

	LastPosOf(n node) (int, bool)

	AddNamedGroup(name string, pos int)
	MatchNamedGroup(name string, pos int)
	DeleteNamedGroup(name string)

	AddGroup(name string, pos int)
	MatchGroup(name string, pos int)
	DeleteGroup(name string)
}

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
	matches map[node]*matchesList
}

var _ Handler = new(Scanner)

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
	s.matches = make(map[node]*matchesList)

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

func (s *Scanner) Match(n node, from, to int, leaf, empty bool) {
	x := nodeMatch{
		node: n,
		span: span{
			from:  from,
			to:    to,
			empty: empty,
		},
	}

	s.expression.append(x)

	if leaf {
		begin := s.firstSpan()
		beginSubstring := s.firstNotEmptySpan()
		endSubstring := s.lastNotEmptySpan()

		m := &stringMatch{
			span: span{
				from: begin.From(),
				to:   begin.From(),
			},
			expressions: newDict().merge(n.getExpressions()), // вот это можно делать только в методе Matches,
			groups:      s.groups.ToSlice(),                  // а в matches list хранить только спаны, так быстрее
			namedGroups: s.namedGroups.ToMap(),
		}

		if m.span.from >= s.input.Size() { // fix for empty matches
			m.span.from = s.input.Size() - 1
		}

		m.span.empty = true

		if beginSubstring != nil && endSubstring != nil {
			subString, err := s.input.Substring(
				beginSubstring.From(),
				endSubstring.To(),
			)

			if err != nil {
				// TODO : how to handle error?
				fmt.Println("error", err)
			}

			m.subString = subString

			m.span.from = beginSubstring.From()
			m.span.to = endSubstring.To()
			m.span.empty = len(subString) == 0
		}

		list, exists := s.matches[n]
		if !exists {
			list = newMatchesList()
			s.matches[n] = list
		}

		list.push(m)
	}
}

func (s *Scanner) matchesToString() string {
	x := ""

	for k, v := range s.matches {
		x += k.getKey() + " - " + v.String() + "\n"
	}

	return x
}

func (s *Scanner) LastPosOf(n node) (int, bool) {
	m, exists := s.matches[n]
	if !exists {
		return -1, false
	}

	stringMatch, exists := m.maximum()
	if !exists {
		return -1, false
	}

	return stringMatch.To(), true
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

func (s *Scanner) firstSpan() *span {
	if s.expression.len() > 0 {
		return &s.expression.first().span
	}

	return nil
}

func (s *Scanner) LastMatch() *nodeMatch {
	if s.expression.len() > 0 {
		return s.expression.last()
	}

	return nil
}

func (s *Scanner) firstNotEmptySpan() *span {
	for i := 0; i < s.expression.len(); i++ {
		m := s.expression.at(i)
		if !m.span.empty {
			return &m.span
		}
	}

	return nil
}

func (s *Scanner) lastNotEmptySpan() *span {
	for i := s.expression.len() - 1; i >= 0; i-- {
		m := s.expression.at(i)
		if !m.span.empty {
			return &m.span
		}
	}

	return nil
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
	node node
	span span
}

func (m nodeMatch) String() string {
	return fmt.Sprintf("nodeMatch{%s: %s}", m.span, m.node.getKey())
}

type Match interface {
	From() int
	To() int
	Size() int
	String() string
}

type span struct {
	from  int
	to    int  // TODO : remove store size and remove "empty" flag?
	empty bool // required for empty matches like .? or .*
}

func (m span) From() int {
	return m.from
}

func (m span) To() int {
	return m.to
}

// required for empty matches like .? or .*
func (m span) Empty() bool {
	return m.empty
}

func (m span) Size() int {
	if m.empty {
		return 0
	}

	return m.to - m.from + 1
}

func (m span) String() string {
	if m.empty {
		return fmt.Sprintf("(%d)", m.from)
	}

	return fmt.Sprintf("[%d-%d]", m.from, m.to)
}

type stringMatch struct {
	subString   string
	span        span
	expressions dict
	groups      []span
	namedGroups map[string]span
}

func (m *stringMatch) From() int {
	return m.span.From()
}

func (m *stringMatch) To() int {
	return m.span.To()
}

func (m *stringMatch) GetSpan() span {
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

// TODO : в тестах проверять, что groups входят в span / bouns строки

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

func (m *stringMatch) NamedGroups() map[string]span {
	return m.namedGroups
}

func (m *stringMatch) Groups() []span {
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
