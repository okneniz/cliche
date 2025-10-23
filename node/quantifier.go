package node

import (
	"github.com/okneniz/cliche/quantity"
)

// https://www.regular-expressions.info/repeat.html

type quantifier struct {
	quantity *quantity.Quantity
	value    Node
	*base
}

var _ Container = new(quantifier)

func NewQuantifier(q *quantity.Quantity, value Node) Node {
	n := &quantifier{
		quantity: q,
		value:    value,
		base:     newBase(value.GetKey() + q.String()),
	}

	return n
}

func (n *quantifier) GetValue() Node {
	return n.value
}

func (n *quantifier) Visit(scanner Scanner, input Input, from, to int, match Callback) {
	start := scanner.Position()

	n.recursiveVisit(1, scanner, input, from, to, func(value Node, mFrom, mTo int, empty bool) {
		pos := scanner.Position()
		match(n, from, mTo, empty)
		nextFrom := nextFor(mTo, empty)
		n.base.VisitNested(scanner, input, nextFrom, to, match)
		scanner.Rewind(pos)
	})

	scanner.Rewind(start)

	// for zero matches like .? or .* or .{0,X}
	if n.quantity.Optional() {
		match(n, from, from, true)
		n.base.VisitNested(scanner, input, from, to, match)
		scanner.Rewind(start)
	}
}

// TODO :rewrite without recursion, if it possible
func (n *quantifier) recursiveVisit(
	count int,
	scanner Scanner,
	input Input,
	from, to int,
	match Callback,
) {
	// TODO : maybe return n, ignore match?
	n.value.Visit(scanner, input, from, to, func(m Node, mFrom, mTo int, empty bool) {
		if n.quantity.Gt(count) {
			if n.quantity.Include(count) {
				match(m, mFrom, mTo, empty)
			}

			n.recursiveVisit(count+1, scanner, input, mTo+1, to, match)
		}
	})
}

// TODO : return list of sizes?
// TODO : add tests to fail on parsing not fixed size quantificators in look behind assertions
func (n *quantifier) Size() (int, bool) {
	// TODO : fix it
	// TODO : size * quantity
	if size, fixedSize := n.value.Size(); fixedSize {
		if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}

func (n *quantifier) Copy() Node {
	return NewQuantifier(n.quantity, n.value.Copy())
}
