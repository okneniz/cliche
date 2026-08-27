package node

import "github.com/okneniz/cliche/span"

type startOfString struct {
	*base
}

func NewStartOfString() Node {
	return &startOfString{
		base: newBase("\\A"),
	}
}

func (n *startOfString) Visit(scanner Scanner, input Input, sp span.Interface, match Callback) {
	if sp.From() != 0 {
		return
	}

	pos := scanner.Position()
	match(n, span.Empty(sp.From()))
	n.base.VisitNested(scanner, input, sp, match)
	scanner.Rewind(pos)
}

func (n *startOfString) Size() (int, bool) {
	return 0, true
}

func (n *startOfString) Copy() Node {
	return NewStartOfString()
}
