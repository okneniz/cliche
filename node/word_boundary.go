package node

import (
	"unicode"

	"github.com/okneniz/cliche/span"
)

type wordBoundary struct {
	*base
}

func NewWordBoundary() Node {
	return &wordBoundary{
		base: newBase("\\b"),
	}
}

// https://www.regular-expressions.info/wordboundaries.html
//
// Before the first character in the string, if the first character is a word character.
// After the last character in the string, if the last character is a word character.
// Between two characters in the string, where one is a word character and the other is not a word character.

func (n *wordBoundary) Visit(scanner Scanner, input Input, sp span.Interface, match Callback) {
	isWordBoundary := (!n.isWord(input, sp.From()-1) && n.isWord(input, sp.From())) ||
		(n.isWord(input, sp.From()-1) && !n.isWord(input, sp.From()))

	if isWordBoundary {
		pos := scanner.Position()

		match(n, span.Empty(sp.From()))
		n.base.VisitNested(scanner, input, sp, match)
		scanner.Rewind(pos)
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

func (n *wordBoundary) Copy() Node {
	return NewWordBoundary()
}
