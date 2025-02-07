package node

import "fmt"

type namedGroup struct {
	Name  string      `json:"name,omitempty"`
	Value Alternation `json:"value,omitempty"`
	*nestedNode
}

func NewNamedGroup(name string, expression Alternation) Node {
	g := &namedGroup{
		Name:       name,
		nestedNode: newNestedNode(),
		Value:      expression,
	}

	return g
}

func (n *namedGroup) GetKey() string {
	return fmt.Sprintf("(?<%s>%s)", n.Name, n.Value.GetKey())
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
		func(_ Node, vFrom, vTo int, empty bool) {
			pos := scanner.Position()
			groupsPos := scanner.NamedGroupsPosition()

			// TODO : why to? what about empty captures
			scanner.MatchNamedGroup(n.Name, from, vTo)
			scanner.Match(n, from, vTo, n.IsEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.VisitNested(scanner, input, vTo+1, to, onMatch)

			scanner.Rewind(pos)
			scanner.RewindNamedGroups(groupsPos)
		},
	)
}

func (n *namedGroup) Size() (int, bool) {
	if size, fixedSize := n.Value.Size(); fixedSize {
		if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}
