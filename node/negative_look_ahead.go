package node

import (
	"fmt"

	"github.com/okneniz/cliche/span"
)

type negativeLookAhead struct {
	value Alternation
	*base
}

var _ Container = new(negativeLookAhead)

func NewNegativeLookAhead(alt Alternation) Node {
	return &negativeLookAhead{
		value: alt,
		base:  newBase(fmt.Sprintf("(?!%s)", alt.GetKey())),
	}
}

func (n *negativeLookAhead) GetValue() Node {
	return n.value
}

func (n *negativeLookAhead) Visit(scanner Scanner, input Input, bounds span.Interface, match Callback) {
	matched := false
	pos := scanner.Position()

	for range n.value.VisitAlternation(scanner, input, bounds) {
		matched = true
		break
	}

	scanner.Rewind(pos)

	// TODO : move to loop
	if !matched {
		match(n, span.Empty(bounds.From()))
		n.base.VisitNested(scanner, input, bounds, match)
		scanner.Rewind(pos)
	}
}

func (n *negativeLookAhead) Size() (int, bool) {
	return 0, false
}

func (n *negativeLookAhead) Copy() Node {
	return NewNegativeLookAhead(n.value.CopyAlternation())
}
