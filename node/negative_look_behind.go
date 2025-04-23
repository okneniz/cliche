package node

import "fmt"

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

func (n *negativeLookBehind) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.nested {
		x.Traverse(f)
	}
}

func (n *negativeLookBehind) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	// TODO : what about anchors?
	pos := scanner.Position()

	if from < n.subExpressionSize {
		scanner.Match(n, from, from, n.IsLeaf(), true)
		onMatch(n, from, from, true)
		n.base.VisitNested(scanner, input, from, to, onMatch)
		scanner.Rewind(pos)
		return
	}

	matched := false

	n.value.VisitAlternation(
		scanner,
		input,
		from-n.subExpressionSize,
		to,
		func(_ Node, vFrom, vTo int, empty bool) bool {
			scanner.Rewind(pos)
			matched = true
			return true
		},
	)

	scanner.Rewind(pos)

	if !matched {
		scanner.Rewind(pos)
		scanner.Match(n, from, from, n.IsLeaf(), true)
		onMatch(n, from, from, true)
		n.base.VisitNested(scanner, input, from, to, onMatch)
	}

	scanner.Rewind(pos)
}

func (n *negativeLookBehind) Size() (int, bool) {
	return 0, false
}
