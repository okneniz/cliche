package scanner

import (
	"fmt"
	"sort"
	"strings"

	"github.com/okneniz/cliche/span"
	"github.com/okneniz/cliche/structs"
)

type Match struct {
	subString   string
	span        span.Interface // todo : rename to bounds
	expressions structs.Set[string]
	groups      []Group
}

type Group struct {
	Index int
	Name  string
	Span  span.Interface
}

func (m *Match) SubString() string {
	return m.subString
}

func (m *Match) Span() span.Interface {
	return m.span
}

func (m *Match) Key() string {
	sort.Slice(m.groups, func(i, j int) bool {
		return m.groups[i].Index < m.groups[j].Index
	})

	s := m.span.String()
	s += "-"
	s += m.groupsToString()
	return s
}

func (m *Match) Expressions() []string {
	return m.expressions.Slice()
}

func (m *Match) Groups() []Group {
	return m.groups
}

func (m *Match) Clone() *Match {
	return &Match{
		subString:   m.subString,
		span:        m.span,
		expressions: m.expressions.Clone(),
		groups:      m.groups, // clone it too?
	}
}

func (m *Match) String() string {
	return fmt.Sprintf(
		"Match{%s, '%s', /%s/ [%s]}",
		m.span.String(),
		m.subString,
		strings.Join(m.expressions.Slice(), ", "),
		m.groupsToString(),
	)
}

func (m *Match) groupsToString() string {
	pairs := make([]string, 0, len(m.groups))
	for k, v := range m.groups {
		if v.Name == "" {
			pairs = append(pairs, fmt.Sprintf("%d : %s", k, v.Span))
		} else {
			pairs = append(pairs, fmt.Sprintf("%d (%s): %s", k, v.Name, v.Span))
		}
	}

	sort.SliceStable(pairs, func(i, j int) bool { return pairs[i] < pairs[j] })
	return strings.Join(pairs, ", ")
}
