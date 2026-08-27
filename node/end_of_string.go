package node

import "github.com/okneniz/cliche/span"

type endOfString struct {
	*base
}

func NewEndOfString() Node {
	return &endOfString{
		base: newBase("\\z"),
	}
}

func (n *endOfString) Visit(scanner Scanner, input Input, from, to int, match Callback) {
	if from == input.Size() {
		pos := scanner.Position()
		match(n, span.Empty(from))
		n.base.VisitNested(scanner, input, from, to, match)
		scanner.Rewind(pos)
	}
}

func (n *endOfString) Size() (int, bool) {
	if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
		return nestedSize, true
	}

	return 0, false
}

func (n *endOfString) Copy() Node {
	return NewEndOfString()
}
