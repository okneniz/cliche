package node

type startOfString struct {
	*base
}

func NewStartOfString() Node {
	return &startOfString{
		base: newBase("\\A"),
	}
}

func (n *startOfString) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *startOfString) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from != 0 {
		return
	}

	pos := scanner.Position()
	scanner.Match(n, from, from, n.IsLeaf(), true)
	onMatch(n, from, from, true)
	n.base.VisitNested(scanner, input, from, to, onMatch)
	scanner.Rewind(pos)
}

func (n *startOfString) Size() (int, bool) {
	return 0, true
}
