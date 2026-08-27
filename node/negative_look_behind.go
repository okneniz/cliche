package node

import (
	"fmt"

	"github.com/okneniz/cliche/span"
)

type negativeLookBehind struct {
	subExpressionSize int
	value             Alternation
	*base
}

func NewNegativeLookBehind(alt Alternation) (Node, error) {
	size, fixedSize := alt.Size()
	if !fixedSize {
		return nil, fmt.Errorf("Invalid pattern in negative look-behind, must be fixed size")
	}

	return &negativeLookBehind{
		subExpressionSize: size,
		value:             alt,
		base:              newBase(fmt.Sprintf("(?<!%s)", alt.GetKey())),
	}, nil
}

func (n *negativeLookBehind) Visit(scanner Scanner, input Input, bounds span.Interface, match Callback) {
	// TODO : what about anchors?
	pos := scanner.Position()

	if bounds.From() < n.subExpressionSize {
		match(n, span.Empty(bounds.From()))
		n.base.VisitNested(scanner, input, bounds, match)
		scanner.Rewind(pos)
		return
	}

	matched := false
	altSp := span.Pair(bounds.From()-n.subExpressionSize, bounds.To())

	n.value.VisitAlternation(
		scanner,
		input,
		altSp,
		func(_ Node, _ span.Interface) bool {
			scanner.Rewind(pos)
			matched = true
			return true
		},
	)

	scanner.Rewind(pos)

	if !matched {
		match(n, span.Empty(bounds.From()))
		n.base.VisitNested(scanner, input, bounds, match)
		scanner.Rewind(pos)
	}
}

func (n *negativeLookBehind) Size() (int, bool) {
	return 0, false
}

func (n *negativeLookBehind) Copy() Node {
	x, _ := NewNegativeLookBehind(n.value.CopyAlternation())
	return x
}
