package node

import (
	"github.com/okneniz/cliche/quantity"
	"github.com/okneniz/cliche/span"
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
	startGroup := scanner.GroupsPosition()

	n.recursiveVisit(1, scanner, input, from, to, func(value Node, sp span.Interface) {
		pos := scanner.Position()

		if startGroup != scanner.GroupsPosition() {
			if lastGroupSpan, ok := scanner.GetGroup(scanner.GroupsPosition()); ok {
				scanner.RewindGroups(startGroup)
				scanner.MatchGroup(lastGroupSpan.From(), lastGroupSpan.To())
			}
		}

		match(n, span.New(from, sp.To(), sp.Empty()))
		nextFrom := nextFor(sp.To(), sp.Empty())
		n.base.VisitNested(scanner, input, nextFrom, to, match)

		scanner.RewindGroups(startGroup)
		scanner.Rewind(pos)
	})

	scanner.RewindGroups(startGroup)
	scanner.Rewind(start)

	// for zero matches like .? or .* or .{0,X}
	if n.quantity.Optional() {
		match(n, span.Empty(from))
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
	if input.Size() <= from {
		return
	}

	n.value.Visit(scanner, input, from, to, func(m Node, sp span.Interface) {
		if n.quantity.Gt(count) {
			if n.quantity.Include(count) {
				match(m, sp)
			}

			n.recursiveVisit(count+1, scanner, input, sp.To()+1, to, match)
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
