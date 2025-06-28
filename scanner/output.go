package scanner

import (
	"fmt"
	"strings"

	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/quantity"
)

// TODO : add unit tests too

type Output struct {
	matches map[node.Node]*matchesList
}

var _ node.Output = NewOutput()

func NewOutput() *Output {
	out := new(Output)
	out.matches = make(map[node.Node]*matchesList)
	return out
}

func (out *Output) String() string {
	ms := make([]string, 0, 10)
	for _, v := range out.Slice() {
		ms = append(ms, v.String())
	}

	return fmt.Sprintf(
		"Output(matches=[%s])",
		strings.Join(ms, ", "),
	)
}

// TODO : remove subString from params
func (out *Output) Yield(
	n node.Node,
	subString string,
	sp quantity.Interface,
	groups []quantity.Interface,
	namedGroups map[string]quantity.Interface,
) {
	m := &Match{
		subString: subString,
		span:      sp,
		// TODO : how to avoid copies / allocations?
		// maybe just save pointer to current expression
		// and swap nodes expressions by new collection only
		// when new expressions add (or any other mutations)?
		expressions: n.GetExpressions().Clone(),
		groups:      groups,
		namedGroups: namedGroups,
	}

	list, exists := out.matches[n]
	if !exists {
		list = newMatchesList()
		out.matches[n] = list
	}

	list.Push(m)
}

func (out *Output) LastPosOf(n node.Node) (int, bool) {
	m, exists := out.matches[n]
	if !exists {
		return -1, false
	}

	match, exists := m.Maximum()
	if !exists {
		return -1, false
	}

	return match.Span().To(), true
}

func (out *Output) Slice() []*Match {
	size := 0
	for _, v := range out.matches {
		size += v.Size()
	}

	result := make([]*Match, 0, size)
	c := make(map[string]int)

	for _, list := range out.matches {
		for _, v := range list.Slice() {
			key := v.Key()

			if idx, exists := c[key]; exists {
				v.expressions.AddTo(result[idx].expressions)
			} else {
				c[key] = len(result)
				// TODO : how to remove copy / allocations?
				result = append(result, v.Clone())
			}
		}
	}

	return result
}
