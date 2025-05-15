package node

import (
	"unicode"
)

type wordBoundary struct {
	*base
}

func NewWordBoundary() Node {
	return &wordBoundary{
		base: newBase("\\b"),
	}
}

func (n *wordBoundary) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.nested {
		x.Traverse(f)
	}
}

// https://www.regular-expressions.info/wordboundaries.html
//
// Before the first character in the string, if the first character is a word character.
// After the last character in the string, if the last character is a word character.
// Between two characters in the string, where one is a word character and the other is not a word character.

func (n *wordBoundary) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	isWordBoundary := (!n.isWord(input, from-1) && n.isWord(input, from)) ||
		(n.isWord(input, from-1) && !n.isWord(input, from))

	if isWordBoundary {
		scanner.Match(n, from, from, n.IsLeaf(), true)
		n.base.VisitNested(scanner, input, from, to, onMatch)
	}
}

func (n *wordBoundary) isWord(input Input, pos int) bool {
	if pos < 0 || pos >= input.Size() {
		return false
	}

	x := input.ReadAt(pos)
	return x == '_' || unicode.IsLetter(x) || unicode.IsDigit(x)
}

func (n *wordBoundary) Size() (int, bool) {
	return 0, false
}
