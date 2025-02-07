package node

import "fmt"

type lookAhead struct {
	Value Alternation `json:"value,omitempty"`
	*nestedNode
}

func NewLookAhead(expression Alternation) Node {
	return &lookAhead{
		Value:      expression,
		nestedNode: newNestedNode(),
	}
}

func (n *lookAhead) GetKey() string {
	return fmt.Sprintf("(?=%s)", n.Value.GetKey())
}

func (n *lookAhead) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *lookAhead) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	n.Value.VisitAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			pos := scanner.Position()
			holesPos := scanner.HolesPosition()

			// what about empty spans, just skip it?
			scanner.MarkAsHole(from, vTo) // or just scanner rewind to "FROM" pos without holes?
			scanner.Match(n, from, from, n.IsEnd(), true)
			onMatch(n, from, from, true)

			scanner.RewindHoles(holesPos)
			n.nestedNode.VisitNested(scanner, input, from, to, onMatch)

			scanner.Rewind(pos)
		},
	)
}

func (n *lookAhead) Size() (int, bool) {
	return 0, false
}
