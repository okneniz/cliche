package node

import (
	"unicode"

	"github.com/okneniz/cliche/span"
)

type nonWordBoundary struct {
	*base
}

func NewNonWordBoundary() Node {
	return &nonWordBoundary{
		base: newBase("\\B"),
	}
}

func (n *nonWordBoundary) Visit(
	scanner Scanner,
	input Input,
	sp span.Interface,
	match Callback,
) {
	isWordBoundary := (!n.isWord(input, sp.From()-1) && n.isWord(input, sp.From())) ||
		(n.isWord(input, sp.From()-1) && !n.isWord(input, sp.From()))

	if !isWordBoundary {
		pos := scanner.Position()
		match(n, span.Empty(sp.From()))
		n.base.VisitNested(scanner, input, sp, match)
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
