package node

import "fmt"

type negativeLookAhead struct {
	Value Alternation `json:"value,omitempty"`
	*base
}

func NewNegativeLookAhead(alt Alternation) Node {
	return &negativeLookAhead{
		Value: alt,
		base:  newBase(fmt.Sprintf("(?!%s)", alt.GetKey())),
	}
}

func (n *negativeLookAhead) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *negativeLookAhead) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	matched := false
	pos := scanner.Position()

	n.Value.VisitAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) bool {
			matched = true
			scanner.Rewind(pos)
			return true
		},
	)

	scanner.Rewind(pos)

	if !matched {
		scanner.Match(n, from, from, n.IsLeaf(), true)
		onMatch(n, from, from, true)
		n.base.VisitNested(scanner, input, from, to, onMatch)
	}

	scanner.Rewind(pos)
}

func (n *negativeLookAhead) Size() (int, bool) {
	return 0, false
}
