package node

import "fmt"

type lookBehind struct {
	Value             Alternation `json:"value,omitempty"`
	subExpressionSize int
	*base
}

func NewLookBehind(alt Alternation) (*lookBehind, error) {
	size, fixedSize := alt.Size()
	if !fixedSize {
		return nil, fmt.Errorf("Invalid pattern in look-behind, must be fixed size")
	}

	return &lookBehind{
		Value:             alt,
		subExpressionSize: size,
		base:              newBase(fmt.Sprintf("(?<=%s)", alt.GetKey())),
	}, nil
}

func (n *lookBehind) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *lookBehind) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	// TODO : what about anchors?
	if from < n.subExpressionSize {
		return
	}

	pos := scanner.Position()

	n.Value.VisitAlternation(
		scanner,
		input,
		from-n.subExpressionSize,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			scanner.Rewind(pos)

			scanner.Match(n, from, from, n.IsLeaf(), true)
			onMatch(n, from, from, true)
			n.base.VisitNested(scanner, input, from, to, onMatch)

			scanner.Rewind(pos)
		},
	)
}

func (n *lookBehind) Size() (int, bool) {
	return 0, false
}
