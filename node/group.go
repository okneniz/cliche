package node

import (
	"fmt"

	"github.com/okneniz/cliche/span"
)

type group struct {
	value Alternation
	*base
}

var _ Container = new(group)

func NewGroup(alt Alternation) Node {
	return &group{
		value: alt,
		base:  newBase(fmt.Sprintf("(%s)", alt.GetKey())),
	}
}

func (n *group) GetValue() Node {
	return n.value
}

func (n *group) Visit(scanner Scanner, input Input, b span.Interface, match Callback) {
	n.value.VisitAlternation(
		scanner,
		input,
		b,
		func(x Node, sp span.Interface) bool {
			pos := scanner.Position()
			groupsPos := scanner.GroupsPosition()

			scanner.MatchGroup(sp.From(), sp.To())

			match(n, sp)

			nextFrom := nextFor(sp.To(), sp.Empty())
			next := span.Pair(nextFrom, b.To())

			n.base.VisitNested(scanner, input, next, match)

			scanner.Rewind(pos)
			scanner.RewindGroups(groupsPos)

			return false
		},
	)
}

func (n *group) Size() (int, bool) {
	if size, fixedSize := n.value.Size(); fixedSize {
		if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}

func (n *group) Copy() Node {
	return NewGroup(n.value.CopyAlternation())
}
