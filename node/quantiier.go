package node

// https://www.regular-expressions.info/repeat.html

// TODO : rename to repeat?

type quantifier struct {
	Quantity *Quantity `json:"quantity"`
	Value    Node      `json:"value,omitempty"`
	*nestedNode
}

func NewQuantifier(q *Quantity, value Node) Node {
	return &quantifier{
		Quantity:   q,
		Value:      value,
		nestedNode: newNestedNode(),
	}
}

func (n *quantifier) GetKey() string {
	return n.Value.GetKey() + n.Quantity.String()
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
		scanner.Match(n, from, mTo, n.IsEnd(), false)
		onMatch(n, from, mTo, empty)
		n.nestedNode.VisitNested(scanner, input, mTo+1, to, onMatch)
		scanner.Rewind(pos)
	})

	scanner.Rewind(start)

	// for zero matches like .? or .* or .{0,X}
	if n.Quantity.Optional() {
		scanner.Match(n, from, from, n.IsEnd(), true)
		n.nestedNode.VisitNested(scanner, input, from, to, onMatch)
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
		if n.Quantity.Gt(count) { // TODO : why not just inlude?
			if n.Quantity.Include(count) { // TODO : why?, maybe remove it?
				onMatch(match, mFrom, mTo, empty)
			}

			next := count + 1

			if n.Quantity.Gt(next) { // TODO : remove it? double check?
				n.recursiveVisit(next, scanner, input, mTo+1, to, onMatch)
			}
		}
	})
}

// TODO : add tests to fail on parsing not fixed size quantificators in look behind assertions
func (n *quantifier) Size() (int, bool) {
	// TODO : fix it
	// TODO : size * quantity
	if size, fixedSize := n.Value.Size(); fixedSize {
		if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}
