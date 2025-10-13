package node

import (
	"fmt"
	"unicode"
)

// back reference \1, \2 or \9
type referenceNode struct {
	index int
	*base
}

func NodeForReference(index int) Node {
	return &referenceNode{
		index: index,
		base:  newBase(fmt.Sprintf("\\%d", index)),
	}
}

func (n *referenceNode) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	matchSpan, exists := scanner.GetGroup(n.index)

	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Regular_expressions/Backreference
	//
	// If the referenced capturing group is unmatched (for example, because it belongs to an unmatched alternative in a disjunction),
	// or the group hasn't matched yet (for example, because it lies to the right of the backreference),
	// the backreference always succeeds (as if it matches the empty string).

	pos := scanner.Position()

	if !exists || matchSpan.Empty() {
		scanner.Match(n, from, from, n.IsLeaf(), true)
		onMatch(n, from, from, true)
		n.base.VisitNested(scanner, input, from, to, onMatch)

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

			matched := false

			if scanner.OptionsInclude(ScanOptionCaseInsensetive) {
				matched = unicode.ToUpper(expected) == unicode.ToUpper(actual)
			} else {
				matched = expected == actual
			}

			if !matched {
				scanner.Rewind(pos)
				return
			}

			current++
		}

		scanner.Match(n, from, current-1, n.IsLeaf(), false)
		onMatch(n, from, current-1, false)

		n.base.VisitNested(scanner, input, current, to, onMatch)
		scanner.Rewind(pos)
	}
}

func (n *referenceNode) Size() (int, bool) {
	return 0, false
}

func (n *referenceNode) Copy() Node {
	return NodeForReference(n.index)
}
