package node

import (
	"fmt"
	"unicode"

	"github.com/okneniz/cliche/span"
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

func (n *referenceNode) Visit(scanner Scanner, input Input, bounds span.Interface, match Callback) {
	if bounds.From() >= input.Size() {
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
		match(n, span.Empty(bounds.From()))
		n.base.VisitNested(scanner, input, bounds, match)
		scanner.Rewind(pos)
	} else {
		// TODO : what about empty matches?

		current := bounds.From()

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

		// TODO : why -1 ? looks strange
		match(n, span.Pair(bounds.From(), current-1))
		next := span.Pair(current, bounds.To())
		n.base.VisitNested(scanner, input, next, match)
		scanner.Rewind(pos)
	}
}

func (n *referenceNode) Size() (int, bool) {
	return 0, false
}

func (n *referenceNode) Copy() Node {
	return NodeForReference(n.index)
}
