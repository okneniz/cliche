package node

type character struct {
	match func(rune) bool
	*base
}

func NewForTable(table Table) Node {
	return &character{
		match: table.Include,
		base:  newBase(table.String()),
	}
}

func (n *character) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.nested {
		x.Traverse(f)
	}
}

func (n *character) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	if n.match(input.ReadAt(from)) {
		pos := scanner.Position()

		scanner.Match(n, from, from, n.IsLeaf(), false)
		onMatch(n, from, from, false)
		n.base.VisitNested(scanner, input, from+1, to, onMatch)

		scanner.Rewind(pos)
	}
}

func (n *character) Size() (int, bool) {
	if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
		return 1 + nestedSize, true
	}

	return 0, false
}
