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

func (n *endOfString) Visit(scanner Scanner, input Input, sp span.Interface, match Callback) {
	if sp.From() == input.Size() {
		pos := scanner.Position()
		match(n, span.Empty(sp.From()))
		n.base.VisitNested(scanner, input, sp, match)
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
