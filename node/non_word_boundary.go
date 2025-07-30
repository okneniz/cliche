package node

import (
	"unicode"
)

type nonWordBoundary struct {
	*base
}

func NewNonWordBoundary() Node {
	return &nonWordBoundary{
		base: newBase("\\B"),
	}
}

func (n *nonWordBoundary) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.nested {
		x.Traverse(f)
	}
}

func (n *nonWordBoundary) Visit(
	scanner Scanner,
	input Input,
	from, to int,
	onMatch Callback,
) {
	isWordBoundary := (!n.isWord(input, from-1) && n.isWord(input, from)) ||
		(n.isWord(input, from-1) && !n.isWord(input, from))

	if !isWordBoundary {
		pos := scanner.Position()

		scanner.Match(n, from, from, n.IsLeaf(), true)
		n.base.VisitNested(scanner, input, from, to, onMatch)

		scanner.Rewind(pos)
	}
}

func (n *nonWordBoundary) isWord(input Input, pos int) bool {
	if pos < 0 || pos >= input.Size() {
		return false
	}

	x := input.ReadAt(pos)
	return x == '_' || unicode.IsLetter(x) || unicode.IsDigit(x)
}

func (n *nonWordBoundary) Size() (int, bool) {
	return 0, false
}

func (n *nonWordBoundary) Copy() Node {
	return NewNonWordBoundary()
}
