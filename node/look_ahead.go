package node

import (
	"fmt"

	"github.com/okneniz/cliche/span"
)

// https://www.regular-expressions.info/keep.html

type lookAhead struct {
	value Alternation
	*base
}

var _ Container = new(lookAhead)

func NewLookAhead(alt Alternation) Node {
	return &lookAhead{
		value: alt,
		base:  newBase(fmt.Sprintf("(?=%s)", alt.GetKey())),
	}
}

func (n *lookAhead) GetValue() Node {
	return n.value
}

func (n *lookAhead) Visit(scanner Scanner, input Input, bounds span.Interface, match Callback) {
	pos := scanner.Position()
	holesPos := scanner.HolesPosition()

	n.value.VisitAlternation(
		scanner,
		input,
		bounds,
		func(x Node, sp span.Interface) bool {
			scanner.Rewind(pos)

			scanner.MarkAsHole(sp.From(), sp.To())
			match(n, span.Empty(sp.From()))
			scanner.RewindHoles(holesPos)

			n.base.VisitNested(scanner, input, bounds, match)
			scanner.Rewind(pos)
			return false
		},
	)
}

func (n *lookAhead) Size() (int, bool) {
	return 0, false
}

func (n *lookAhead) Copy() Node {
	return NewLookAhead(n.value.CopyAlternation())
}
