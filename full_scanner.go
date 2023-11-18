package regular

import (
	"fmt"
)

type fullScanner struct {
	groups      *captures
	namedGroups *captures
	onMatch     Callback
	matches     list[match]
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
	s.matches = *newList[match](100) // pointer?
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
		fmt.Println("wtf", n, from, to, leaf, empty)
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

type FullMatch struct {
	expressions []string
	subString   string
	from        int
	to          int
	groups      []bounds
	namedGroups map[string]bounds
	empty       bool // required for empty matches like .? or .*
}

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
