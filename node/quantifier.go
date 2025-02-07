package node

// https://www.regular-expressions.info/repeat.html

type quantifier struct {
	quantity *Quantity
	Value    Node `json:"value,omitempty"`
	*base
}

func NewQuantifier(q *Quantity, value Node) Node {
	return &quantifier{
		quantity: q,
		Value:    value,
		base:     newBase(value.GetKey() + q.String()),
	}
}

func (n *quantifier) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *quantifier) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	start := scanner.Position()

	n.recursiveVisit(1, scanner, input, from, to, func(_ Node, _, mTo int, empty bool) {
		pos := scanner.Position()
		scanner.Match(n, from, mTo, n.IsLeaf(), false)
		onMatch(n, from, mTo, empty)
		n.base.VisitNested(scanner, input, mTo+1, to, onMatch)
		scanner.Rewind(pos)
	})

	scanner.Rewind(start)

	// for zero matches like .? or .* or .{0,X}
	if n.quantity.Optional() {
		scanner.Match(n, from, from, n.IsLeaf(), true)
		n.base.VisitNested(scanner, input, from, to, onMatch)
	}

	scanner.Rewind(start)
}

func (n *quantifier) recursiveVisit(
	count int,
	scanner Scanner,
	input Input,
	from, to int,
	onMatch Callback,
) {
	n.Value.Visit(scanner, input, from, to, func(match Node, mFrom, mTo int, empty bool) {
		if n.quantity.Gt(count) {
			if n.quantity.Include(count) {
				onMatch(match, mFrom, mTo, empty)
			}

			n.recursiveVisit(count+1, scanner, input, mTo+1, to, onMatch)
		}
	})
}

// TODO : add tests to fail on parsing not fixed size quantificators in look behind assertions
func (n *quantifier) Size() (int, bool) {
	// TODO : fix it
	// TODO : size * quantity
	if size, fixedSize := n.Value.Size(); fixedSize {
		if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}
