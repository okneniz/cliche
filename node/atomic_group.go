package node

import "fmt"

// atomicGroup - An atomic group is a group that,
// when the regex engine exits from it,
// automatically throws away all backtracking positions
// remembered by any tokens inside the group.
// Atomic groups are non-capturing.
type atomicGroup struct {
	value Alternation
	*base
}

var _ Container = new(atomicGroup)

func NewAtomicGroup(alt Alternation) Node {
	return &atomicGroup{
		value: alt,
		base:  newBase(fmt.Sprintf("(?>%s)", alt.GetKey())),
	}
}

func (n *atomicGroup) GetValue() Node {
	return n.value
}

func (n *atomicGroup) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	pos := scanner.Position()
	groupsPos := scanner.GroupsPosition()

	n.value.VisitAlternation(
		scanner,
		input,
		from,
		to,
		func(variant Node, vFrom, vTo int, empty bool) bool {
			scanner.Match(n, from, vTo, n.IsLeaf(), empty)
			onMatch(n, from, vTo, empty)

			// throws away all backtracking positions
			scanner.RewindGroups(groupsPos)

			nextFrom := nextFor(vTo, empty)
			n.base.VisitNested(scanner, input, nextFrom, to, onMatch)

			return true // stop on first variant
		},
	)

	scanner.Rewind(pos)
}

// TODO : move to groupBase
func (n *atomicGroup) Size() (int, bool) {
	if size, fixedSize := n.value.Size(); fixedSize {
		if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}

func (n *atomicGroup) Copy() Node {
	return NewAtomicGroup(n.value.CopyAlternation())
}
