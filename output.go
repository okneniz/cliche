package cliche

import (
	"fmt"
	"strings"

	"github.com/okneniz/cliche/span"
)

type Output interface {
	Yield(
		n Node,
		subString string,
		sp span.Interface,
		groups []span.Interface,
		namedGroups map[string]span.Interface,
	)

	LastPosOf(n Node) (int, bool)
	Slice() []*Match
	String() string
}

type output struct {
	matches map[Node]*matchesList
}

var _ Output = new(output)

func newOutput() *output {
	s := new(output)
	s.matches = make(map[Node]*matchesList)
	return s
}

func (s *output) String() string {
	ms := make([]string, 0, 10)
	for _, v := range s.Slice() {
		ms = append(ms, v.String())
	}

	return fmt.Sprintf(
		"Output(matches=[%s])",
		strings.Join(ms, ", "),
	)
}

// TODO : remove subString from params
func (s *output) Yield(
	n Node,
	subString string,
	sp span.Interface,
	groups []span.Interface,
	namedGroups map[string]span.Interface,
) {
	m := &Match{
		subString: subString,
		span:      sp,
		// TODO : how to avoid copies / allocations?
		expressions: newSet().merge(n.GetExpressions()),
		groups:      groups,
		namedGroups: namedGroups,
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

// TODO : []*Match or []Match
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
				// TODO : how to remove copy / allocations?
				result = append(result, v.Clone())
			}
		}
	}

	return result
}
