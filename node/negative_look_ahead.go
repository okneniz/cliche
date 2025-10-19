package node

import "fmt"

type negativeLookAhead struct {
	value Alternation
	*base
}

var _ Container = new(negativeLookAhead)

func NewNegativeLookAhead(alt Alternation) Node {
	return &negativeLookAhead{
		value: alt,
		base:  newBase(fmt.Sprintf("(?!%s)", alt.GetKey())),
	}
}

func (n *negativeLookAhead) GetValue() Node {
	return n.value
}

func (n *negativeLookAhead) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	matched := false
	pos := scanner.Position()

	n.value.VisitAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) bool {
			scanner.Rewind(pos)
			matched = true
			return true
		},
	)

	scanner.Rewind(pos)

	if !matched {
		scanner.Match(n, from, from, true)
		onMatch(n, from, from, true)
		n.base.VisitNested(scanner, input, from, to, onMatch)
	}

	scanner.Rewind(pos)
}

func (n *negativeLookAhead) Size() (int, bool) {
	return 0, false
}

func (n *negativeLookAhead) Copy() Node {
	return NewNegativeLookAhead(n.value.CopyAlternation())
}
