package node

type comment struct {
	*base
	text string
}

func NewComment(text string) Node {
	return &comment{
		base: newBase("comment"),
		text: text,
	}
}

func (n *comment) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.nested {
		x.Traverse(f)
	}
}

func (n *comment) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	pos := scanner.Position()

	scanner.Match(n, from, from, n.IsLeaf(), true)
	onMatch(n, from, from, true)
	n.base.VisitNested(scanner, input, from, to, onMatch)

	scanner.Rewind(pos)
}

func (n *comment) Size() (int, bool) {
	return 0, false
}
