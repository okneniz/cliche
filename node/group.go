package node

import "fmt"

type group struct {
	Value Alternation `json:"value,omitempty"`
	*nestedNode
}

func NewGroup(expression Alternation) *group {
	g := &group{
		nestedNode: newNestedNode(),
		Value:      expression,
	}

	return g
}

func (n *group) GetKey() string {
	return fmt.Sprintf("(%s)", n.Value.GetKey())
}

func (n *group) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *group) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	n.Value.VisitAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			pos := scanner.Position()
			groupsPos := scanner.GroupsPosition()

			// TODO : why to? what about empty captures
			scanner.MatchGroup(from, vTo)
			scanner.Match(n, from, vTo, n.IsEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.VisitNested(scanner, input, vTo+1, to, onMatch)

			scanner.Rewind(pos)
			scanner.RewindGroups(groupsPos)
		},
	)
}

func (n *group) Size() (int, bool) {
	if size, fixedSize := n.Value.Size(); fixedSize {
		if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}
