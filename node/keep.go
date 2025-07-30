package node

type keep struct {
	*base
}

func NewKeep() Node {
	return &keep{
		base: newBase("\\K"),
	}
}

func (n *keep) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.nested {
		x.Traverse(f)
	}
}

func (n *keep) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	pos := scanner.Position()

	if from == 0 {
		n.base.VisitNested(scanner, input, from, to, onMatch)
		scanner.Rewind(pos)
		return
	}

	holesPos := scanner.HolesPosition()

	scanner.MarkAsHole(0, from-1)
	scanner.Match(n, from, from, n.IsLeaf(), true)
	onMatch(n, from, from, true)

	n.base.VisitNested(scanner, input, from, to, onMatch)

	scanner.RewindHoles(holesPos)
	scanner.Rewind(pos)
}

func (n *keep) Size() (int, bool) {
	return 0, false // TODO : fix it
}

func (n *keep) Copy() Node {
	return NewKeep()
}
