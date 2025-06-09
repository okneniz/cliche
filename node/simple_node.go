package node

// https://www.regular-expressions.info/charclass.html
type simpleNode struct {
	predicate func(rune) bool
	*base
}

func NewForTable(table Table) Node {
	return &simpleNode{
		predicate: table.Include,
		base:      newBase(table.String()),
	}
}

func (n *simpleNode) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.nested {
		x.Traverse(f)
	}
}

func (n *simpleNode) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	if n.predicate(input.ReadAt(from)) {
		pos := scanner.Position()

		scanner.Match(n, from, from, n.IsLeaf(), false)
		onMatch(n, from, from, false)
		n.base.VisitNested(scanner, input, from+1, to, onMatch)

		scanner.Rewind(pos)
	}
}

func (n *simpleNode) Size() (int, bool) {
	if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
		return 1 + nestedSize, true
	}

	return 0, false
}
