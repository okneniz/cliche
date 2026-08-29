package node

import (
	"fmt"

	"github.com/okneniz/cliche/span"
)

type namedGroup struct {
	name  string
	value Alternation
	*base
}

var _ Container = new(namedGroup)

func NewNamedGroup(name string, alt Alternation) Node {
	g := &namedGroup{
		name:  name,
		value: alt,
		base:  newBase(fmt.Sprintf("(?<%s>%s)", name, alt.GetKey())),
	}

	return g
}

func (n *namedGroup) GetValue() Node {
	return n.value
}

func (n *namedGroup) Visit(scanner Scanner, input Input, bounds span.Interface, match Callback) {
	for _, sp := range n.value.VisitAlternation(scanner, input, bounds) {
		pos := scanner.Position()
		groupsPos := scanner.NamedGroupsPosition()

		scanner.MatchNamedGroup(n.name, sp)
		match(n, sp)

		nextFrom := nextFor(sp.To(), sp.Empty())
		next := span.Pair(nextFrom, bounds.To())
		n.base.VisitNested(scanner, input, next, match)

		scanner.Rewind(pos)
		scanner.RewindNamedGroups(groupsPos)
	}
}

func (n *namedGroup) Size() (int, bool) {
	if size, fixedSize := n.value.Size(); fixedSize {
		if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}

func (n *namedGroup) Copy() Node {
	return NewNamedGroup(n.name, n.value.CopyAlternation())
}
