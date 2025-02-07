package node

import "fmt"

type negativeLookBehind struct {
	Value             Alternation `json:"value,omitempty"`
	subExpressionSize int
	*base
}

func NewNegativeLookBehind(expression Alternation) (*negativeLookBehind, error) {
	size, fixedSize := expression.Size()
	if !fixedSize {
		return nil, fmt.Errorf("Invalid pattern in negative look-behind, must be fixed size")
	}

	return &negativeLookBehind{
		Value:             expression,
		subExpressionSize: size,
		base:              newBase(),
	}, nil
}

func (n *negativeLookBehind) GetKey() string {
	return fmt.Sprintf("(?<!%s)", n.Value.GetKey())
}

func (n *negativeLookBehind) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
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

	n.Value.VisitAlternation(
		scanner,
		input,
		from-n.subExpressionSize,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			scanner.Rewind(pos)
			matched = true
			// TODO : stop here
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
