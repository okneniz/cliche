package node

import (
	"fmt"

	"github.com/okneniz/cliche/span"
)

type lookBehind struct {
	value             Alternation
	subExpressionSize int
	*base
}

func NewLookBehind(alt Alternation) (Node, error) {
	size, fixedSize := alt.Size()
	if !fixedSize {
		return nil, fmt.Errorf("Invalid pattern in look-behind, must be fixed size")
	}

	return &lookBehind{
		value:             alt,
		subExpressionSize: size,
		base:              newBase(fmt.Sprintf("(?<=%s)", alt.GetKey())),
	}, nil
}

func (n *lookBehind) Visit(scanner Scanner, input Input, sp span.Interface, match Callback) {
	// TODO : what about anchors?
	if sp.From() < n.subExpressionSize {
		return
	}

	pos := scanner.Position()
	altSp := span.New(sp.From()-n.subExpressionSize, sp.To(), false)

	n.value.VisitAlternation(
		scanner,
		input,
		altSp,
		func(_ Node, _ span.Interface) bool {
			scanner.Rewind(pos)

			match(n, span.Empty(sp.From()))
			n.base.VisitNested(scanner, input, sp, match)

			scanner.Rewind(pos)

			return false
		},
	)
}

func (n *lookBehind) Size() (int, bool) {
	return 0, false
}

func (n *lookBehind) Copy() Node {
	x, _ := NewLookBehind(n.value.CopyAlternation())
	return x
}
