package node

import "fmt"

// named back reference \k<name>

type nameReferenceNode struct {
	key  string
	name string
	*nestedNode
}

func NodeForNameReference(name string) *nameReferenceNode {
	return &nameReferenceNode{
		key:        fmt.Sprintf("\\k<%s>", name),
		name:       name,
		nestedNode: newNestedNode(),
	}
}

func (n *nameReferenceNode) GetKey() string {
	return n.key
}

func (n *nameReferenceNode) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *nameReferenceNode) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	matchSpan, exists := scanner.GetNamedGroup(n.name)

	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Regular_expressions/Backreference
	//
	// If the referenced capturing group is unmatched (for example, because it belongs to an unmatched alternative in a disjunction),
	// or the group hasn't matched yet (for example, because it lies to the right of the backreference),
	// the backreference always succeeds (as if it matches the empty string).

	pos := scanner.Position()

	if !exists || matchSpan.Empty() {
		scanner.Match(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.VisitNested(scanner, input, from, to, onMatch)

		scanner.Rewind(pos)
	} else {
		// TODO : what about empty matches?
		current := from

		// match the same string
		for prev := matchSpan.From(); prev <= matchSpan.To(); prev++ {
			if current >= input.Size() {
				scanner.Rewind(pos)
				return
			}

			expected := input.ReadAt(prev)
			actual := input.ReadAt(current)

			if expected != actual {
				scanner.Rewind(pos)
				return
			}

			current++
		}

		scanner.Match(n, from, current-1, n.IsEnd(), false)
		onMatch(n, from, current-1, false)

		n.nestedNode.VisitNested(scanner, input, current, to, onMatch)
		scanner.Rewind(pos)
	}
}

func (n *nameReferenceNode) Size() (int, bool) {
	return 0, false
}
