package scanner

import (
	"fmt"
	"sort"
	"strings"

	"github.com/okneniz/cliche/quantity"
	"github.com/okneniz/cliche/structs"
)

type Match struct {
	subString   string
	span        quantity.Interface // todo : rename to bounds
	expressions structs.Set[string]
	groups      []quantity.Interface
	namedGroups map[string]quantity.Interface
}

func (m *Match) SubString() string {
	return m.subString
}

func (m *Match) Span() quantity.Interface {
	return m.span
}

func (m *Match) Key() string {
	s := m.span.String()
	s += "-"
	s += m.groupsToString()
	s += "-"
	s += m.namedGroupsToString()
	return s
}

func (m *Match) Expressions() []string {
	return m.expressions.Slice()
}

func (m *Match) NamedGroups() map[string]quantity.Interface {
	return m.namedGroups
}

func (m *Match) Groups() []quantity.Interface {
	return m.groups
}

func (m *Match) Clone() *Match {
	return &Match{
		subString:   m.subString,
		span:        m.span,
		expressions: m.expressions.Clone(),
		groups:      m.groups,      // clone it too?
		namedGroups: m.namedGroups, // clone it too?
	}
}

func (m *Match) String() string {
	return fmt.Sprintf(
		"Match{%s, '%s', /%s/ [%s] {%s}}",
		m.span.String(),
		m.subString,
		strings.Join(m.expressions.Slice(), ", "),
		m.groupsToString(),
		m.namedGroupsToString(),
	)
}

func (m *Match) groupsToString() string {
	s := make([]string, len(m.groups))
	for i, x := range m.groups {
		s[i] = x.String()
	}

	sort.SliceStable(s, func(i, j int) bool { return s[i] < s[j] })
	return strings.Join(s, ", ")
}

func (m *Match) namedGroupsToString() string {
	pairs := make([]string, 0, len(m.namedGroups))
	for k, v := range m.namedGroups {
		pairs = append(pairs, k+": "+v.String())
	}
	sort.SliceStable(pairs, func(i, j int) bool { return pairs[i] < pairs[j] })
	return strings.Join(pairs, ", ")
}
