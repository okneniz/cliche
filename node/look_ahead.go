package node

import "fmt"

type lookAhead struct {
	Value Alternation `json:"value,omitempty"`
	*base
}

func NewLookAhead(alt Alternation) Node {
	return &lookAhead{
		Value: alt,
		base:  newBase(fmt.Sprintf("(?=%s)", alt.GetKey())),
	}
}

func (n *lookAhead) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *lookAhead) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	pos := scanner.Position()
	holesPos := scanner.HolesPosition()

	n.Value.VisitAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, _ bool) bool {
			scanner.Rewind(pos)
			scanner.MarkAsHole(vFrom, vTo)

			scanner.Match(n, vFrom, vTo, n.IsLeaf(), true)
			onMatch(n, vFrom, vTo, true)
			scanner.RewindHoles(holesPos)

			n.base.VisitNested(scanner, input, from, to, onMatch)
			scanner.Rewind(pos)
			return false
		},
	)
}

func (n *lookAhead) Size() (int, bool) {
	return 0, false
}
