package node

import "github.com/okneniz/cliche/span"

type comment struct {
	*base
	text string
}

func NewComment(text string) Node {
	return &comment{
		base: newBase("comment"),
		text: text,
	}
}

func (n *comment) Visit(scanner Scanner, input Input, from, to int, match Callback) {
	pos := scanner.Position()

	match(n, span.Empty(from))
	n.base.VisitNested(scanner, input, from, to, match)
	scanner.Rewind(pos)
}

func (n *comment) Size() (int, bool) {
	return 0, false
}

func (n *comment) Copy() Node {
	return NewComment(n.text)
}
