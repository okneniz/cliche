package regular

import (
	"fmt"
)

type Handler interface { // TODO : should be generic type for different type of matches?
	Match(n node, from, to int, isLeaf, isEmpty bool)

	FirstMatch() *match
	FirstNotEmptyMatch() *match

	// TODO : how to remove it?
	// required only for quantifier
	LastMatch() *match // TODO : use (int, int) instead?
	LastNotEmptyMatch() *match

	Position() int
	Rewind(size int)

	AddNamedGroup(name string, index int)
	MatchNamedGroup(name string, index int)
	DeleteNamedGroup(name string)

	AddGroup(name string, index int)
	MatchGroup(name string, index int)
	DeleteGroup(name string)
}

type fullScanner struct {
	groups      *captures
	namedGroups *captures
	onMatch     Callback
	matches     truncatedList[match]
}

var _ Handler = new(fullScanner)

func newFullScanner(
	captures *captures,
	namedCaptures *captures,
	onMatch Callback,
) *fullScanner {
	s := new(fullScanner)
	s.groups = captures
	s.namedGroups = namedCaptures
	s.onMatch = onMatch
	s.matches = newTruncatedList[match](100)
	return s
}

func (s *fullScanner) String() string {
	return fmt.Sprintf(
		"Scanner(matches=%s, groups=%s, namedGroups=%s)",
		s.matches.String(),
		s.groups.String(),
		s.namedGroups.String(),
	)
}

func (s *fullScanner) Match(n node, from, to int, leaf, empty bool) {
	m := match{
		from:  from,
		to:    to,
		node:  n,
		empty: empty,
	}

	s.matches.append(m)

	if leaf {
		s.onMatch(n, from, to, empty)
	}
}

func (s *fullScanner) Position() int {
	return s.matches.len()
}

func (s *fullScanner) Rewind(size int) {
	if s.matches.len() < size {
		return
	}

	s.matches.truncate(size)
}

func (s *fullScanner) FirstMatch() *match {
	if s.matches.len() > 0 {
		return s.matches.first()
	}

	return nil
}

func (s *fullScanner) FirstNotEmptyMatch() *match {
	// TODO : cache it in special memory address in list structure?
	for i := 0; i < s.matches.len(); i++ {
		m := s.matches.at(i)
		if !m.Empty() {
			return &m
		}
	}

	return nil
}

func (s *fullScanner) LastMatch() *match {
	if s.matches.len() > 0 {
		return s.matches.last()
	}

	return nil
}

func (s *fullScanner) LastNotEmptyMatch() *match {
	// TODO : cache it in special memory address in list structure?
	for i := s.matches.len() - 1; i >= 0; i-- {
		m := s.matches.at(i)
		if !m.Empty() {
			return &m
		}
	}

	return nil
}

func (s *fullScanner) AddNamedGroup(name string, index int) {
	s.namedGroups.From(name, index)
}

func (s *fullScanner) MatchNamedGroup(name string, index int) {
	s.namedGroups.To(name, index)
}

func (s *fullScanner) DeleteNamedGroup(name string) {
	s.namedGroups.Delete(name)
}

func (s *fullScanner) AddGroup(name string, index int) {
	s.groups.From(name, index)
}

func (s *fullScanner) MatchGroup(name string, index int) {
	s.groups.To(name, index)
}

func (s *fullScanner) DeleteGroup(name string) {
	s.groups.Delete(name)
}

type Match interface {
	From() int
	To() int
	Size() int
	String() string
}

type match struct {
	from  int
	to    int
	node  node
	empty bool
}

var _ Match = &match{}

func (m match) From() int {
	return m.from
}

func (m match) To() int {
	return m.to
}

func (m match) Empty() bool {
	return m.empty
}

func (m match) String() string {
	return fmt.Sprintf("%s - [%d..%d] %v", m.node.getKey(), m.from, m.to, m.empty)
}

func (m match) Size() int {
	if m.empty {
		return 0
	}

	return m.to - m.from + 1
}

type FullMatch struct {
	expressions []string
	subString   string
	from        int
	to          int
	groups      []bounds
	namedGroups map[string]bounds
	empty       bool // required for empty matches like .? or .*
}

var _ Match = &FullMatch{}

func (m *FullMatch) From() int {
	return m.from
}

func (m *FullMatch) To() int {
	return m.to
}


func (m *FullMatch) Size() int {
	if m.empty {
		return 0
	}

	return len(m.subString)
}

func (m *FullMatch) String() string {
	return m.subString
}

func (m *FullMatch) Expressions() []string {
	return m.expressions
}

func (m *FullMatch) NamedGroups() map[string]bounds {
	return m.namedGroups
}

func (m *FullMatch) Groups() []bounds {
	return m.groups
}
