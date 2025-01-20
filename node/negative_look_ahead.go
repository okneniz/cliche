package node

import "fmt"

type negativeLookAhead struct {
	Value Alternation `json:"value,omitempty"`
	*nestedNode
}

func NewNegativeLookAhead(expression Alternation) *negativeLookAhead {
	return &negativeLookAhead{
		Value:      expression,
		nestedNode: newNestedNode(),
	}
}

func (n *negativeLookAhead) GetKey() string {
	return fmt.Sprintf("(?!%s)", n.Value.GetKey())
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
		func(_ Node, vFrom, vTo int, empty bool) {
			matched = true
			scanner.Rewind(pos)
			// TODO : stop here
		},
	)

	scanner.Rewind(pos)

	if !matched {
		scanner.Rewind(pos)

		scanner.Match(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)

		n.nestedNode.VisitNested(scanner, input, from, to, onMatch)
	}

	scanner.Rewind(pos)
}

func (n *negativeLookAhead) Size() (int, bool) {
	return 0, false
}
