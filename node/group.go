package node

import (
	"fmt"

	"github.com/okneniz/cliche/span"
)

type group struct {
	value Alternation
	index int
	*base
}

var _ Container = new(group)

func NewGroup(alt Alternation, index int) Node {
	return &group{
		value: alt,
		index: index,
		base:  newBase(fmt.Sprintf("(%s)", alt.GetKey())),
	}
}

func (n *group) GetValue() Node {
	return n.value
}

func (n *group) Visit(scanner Scanner, input Input, bounds span.Interface, match Callback) {
	for _, sp := range n.value.VisitAlternation(scanner, input, bounds) {
		pos := scanner.Position()
		groupsPos := scanner.GroupsPosition()

		scanner.MatchGroup(sp)
		match(n, sp)

		nextFrom := span.NextFrom(sp.To(), sp.Empty())
		next := span.Pair(nextFrom, bounds.To())

		n.VisitNested(scanner, input, next, match)

		scanner.Rewind(pos)
		scanner.RewindGroups(groupsPos)
	}
}

func (n *group) Size() (int, bool) {
	if size, fixedSize := n.value.Size(); fixedSize {
		if nestedSize, fixedSize := n.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}

func (n *group) Copy() Node {
	return NewGroup(n.value.CopyAlternation(), n.index)
}
