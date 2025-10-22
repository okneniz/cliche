package node

import "fmt"

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

func (n *group) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	n.value.VisitAlternation(
		scanner,
		input,
		from,
		to,
		func(variant Node, vFrom, vTo int, empty bool) bool {
			pos := scanner.Position()
			groupsPos := scanner.GroupsPosition()

			scanner.MatchGroup(from, vTo)
			onMatch(n, from, vTo, empty)

			nextFrom := nextFor(vTo, empty)
			n.base.VisitNested(scanner, input, nextFrom, to, onMatch)

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
