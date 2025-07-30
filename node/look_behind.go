package node

import "fmt"

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

func (n *lookBehind) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.nested {
		x.Traverse(f)
	}
}

func (n *lookBehind) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	// TODO : what about anchors?
	if from < n.subExpressionSize {
		return
	}

	pos := scanner.Position()

	n.value.VisitAlternation(
		scanner,
		input,
		from-n.subExpressionSize,
		to,
		func(_ Node, vFrom, vTo int, empty bool) bool {
			scanner.Rewind(pos)

			scanner.Match(n, from, from, n.IsLeaf(), true)
			onMatch(n, from, from, true)
			n.base.VisitNested(scanner, input, from, to, onMatch)

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
