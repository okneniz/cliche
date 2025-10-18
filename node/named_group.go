package node

import "fmt"

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

func (n *namedGroup) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	n.value.VisitAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) bool {
			pos := scanner.Position()
			groupsPos := scanner.NamedGroupsPosition()

			scanner.MatchNamedGroup(n.name, from, vTo)
			scanner.Match(n, from, vTo, n.IsLeaf(), empty)
			onMatch(n, from, vTo, empty)

			nextFrom := nextFor(vTo, empty)
			n.base.VisitNested(scanner, input, nextFrom, to, onMatch)

			scanner.Rewind(pos)
			scanner.RewindNamedGroups(groupsPos)

			return false
		},
	)
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
