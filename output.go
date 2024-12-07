package cliche

import (
	"fmt"
	"strings"

	"github.com/okneniz/cliche/span"
)

// https://www.regular-expressions.info/engine.html
// This is a very important point to understand:
// a regex engine always returns the leftmost match,
// even if a “better” match could be found later.

type output struct {
	// current captured groups
	groups *captures

	// current captured named groups
	namedGroups *captures

	// output
	matches map[Node]*matchesList
}

var _ Output = new(output)

func newOutput() *output {
	s := new(output)
	s.groups = newCaptures()
	s.namedGroups = newCaptures()
	s.matches = make(map[Node]*matchesList)
	return s
}

func (s *output) Groups() Captures {
	return s.groups
}

func (s *output) NamedGroups() Captures {
	return s.namedGroups
}

func (s *output) String() string {
	ms := make([]string, 0, 10)
	for _, v := range s.Slice() {
		ms = append(ms, v.String())
	}

	return fmt.Sprintf(
		"Output(\n\tmatches=%s,\n\tgroups=%s,\n\tnamedGroups=%s\n)",
		"[\n"+strings.Join(ms, ",\n")+"]",
		s.groups.String(),
		s.namedGroups.String(),
	)
}

func (s *output) Yield(n Node, x span.Interface, subString string) {
	m := &Match{
		subString:   subString,
		expressions: newSet().merge(n.GetExpressions()),
		groups:      s.groups.ToSlice(),
		namedGroups: s.namedGroups.ToMap(),
		span:        x,
	}

	list, exists := s.matches[n]
	if !exists {
		list = newMatchesList()
		s.matches[n] = list
	}

	list.push(m)
}

func (s *output) LastPosOf(n Node) (int, bool) {
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

func (s *output) Slice() []*Match {
	size := 0
	for _, v := range s.matches {
		size += v.size()
	}

	result := make([]*Match, 0, size)
	c := make(map[string]int)

	for _, list := range s.matches {
		for _, v := range list.Slice() {
			key := v.Key()

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
