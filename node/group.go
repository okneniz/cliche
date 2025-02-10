package node

import "fmt"

type group struct {
	Value Alternation `json:"value,omitempty"`
	*base
}

func NewGroup(alt Alternation) Node {
	return &group{
		Value: alt,
		base:  newBase(fmt.Sprintf("(%s)", alt.GetKey())),
	}
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
		func(_ Node, vFrom, vTo int, empty bool) bool {
			pos := scanner.Position()
			groupsPos := scanner.GroupsPosition()

			// TODO : why to? what about empty captures
			scanner.MatchGroup(from, vTo)
			scanner.Match(n, from, vTo, n.IsLeaf(), false)
			onMatch(n, from, vTo, empty)

			// https://www.regular-expressions.info/alternation.html
			//
			// TODO: The regex engine is eager. It stops searching as soon as it finds a valid match.
			//
			// maybe interupt if any nested node matched?

			n.base.VisitNested(scanner, input, vTo+1, to, onMatch)

			scanner.Rewind(pos)
			scanner.RewindGroups(groupsPos)

			return false
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
