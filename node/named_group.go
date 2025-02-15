package node

import "fmt"

type namedGroup struct {
	Name  string      `json:"name,omitempty"`
	Value Alternation `json:"value,omitempty"`
	*base
}

func NewNamedGroup(name string, alt Alternation) Node {
	g := &namedGroup{
		Name:  name,
		Value: alt,
		base:  newBase(fmt.Sprintf("(?<%s>%s)", name, alt.GetKey())),
	}

	return g
}

func (n *namedGroup) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *namedGroup) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	n.Value.VisitAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) bool {
			pos := scanner.Position()
			groupsPos := scanner.NamedGroupsPosition()

			// TODO : why to? what about empty captures
			scanner.MatchNamedGroup(n.Name, from, vTo)
			scanner.Match(n, from, vTo, n.IsLeaf(), empty)
			onMatch(n, from, vTo, empty)
			n.base.VisitNested(scanner, input, vTo+1, to, onMatch)

			scanner.Rewind(pos)
			scanner.RewindNamedGroups(groupsPos)

			return false
		},
	)
}

func (n *namedGroup) Size() (int, bool) {
	if size, fixedSize := n.Value.Size(); fixedSize {
		if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}
