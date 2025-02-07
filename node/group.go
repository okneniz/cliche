package node

import "fmt"

type group struct {
	Value Alternation `json:"value,omitempty"`
	*base
}

func NewGroup(expression Alternation) Node {
	g := &group{
		base:  newBase(),
		Value: expression,
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
			scanner.Match(n, from, vTo, n.IsLeaf(), false)
			onMatch(n, from, vTo, empty)
			n.base.VisitNested(scanner, input, vTo+1, to, onMatch)

			scanner.Rewind(pos)
			scanner.RewindGroups(groupsPos)
		},
	)
}

func (n *group) Size() (int, bool) {
	if size, fixedSize := n.Value.Size(); fixedSize {
		if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}
