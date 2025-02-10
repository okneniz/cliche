package node

import "fmt"

// atomicGroup - An atomic group is a group that,
// when the regex engine exits from it,
// automatically throws away all backtracking positions
// remembered by any tokens inside the group.
// Atomic groups are non-capturing.
type atomicGroup struct {
	Value Alternation `json:"value,omitempty"`
	*base
}

func NewAtomicGroup(alt Alternation) Node {
	return &atomicGroup{
		Value: alt,
		base:  newBase(fmt.Sprintf("(?>%s)", alt.GetKey())),
	}
}

func (n *atomicGroup) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *atomicGroup) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	pos := scanner.Position()
	groupsPos := scanner.GroupsPosition()

	n.Value.VisitAlternation(
		scanner,
		input,
		from,
		to,
		func(variant Node, vFrom, vTo int, empty bool) bool {
			scanner.Match(n, from, vTo, n.IsLeaf(), empty)
			onMatch(n, from, vTo, empty)

			// throws away all backtracking positions
			scanner.RewindGroups(groupsPos)

			if empty {
				n.base.VisitNested(scanner, input, vTo, to, onMatch)
			} else {
				n.base.VisitNested(scanner, input, vTo+1, to, onMatch)
			}

			return true // stop on first variant
		},
	)

	scanner.Rewind(pos)
}

// TODO : move to groupBase
func (n *atomicGroup) Size() (int, bool) {
	if size, fixedSize := n.Value.Size(); fixedSize {
		if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}
