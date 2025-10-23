package node

type keep struct {
	*base
}

func NewKeep() Node {
	return &keep{
		base: newBase("\\K"),
	}
}

func (n *keep) Visit(scanner Scanner, input Input, from, to int, match Callback) {
	pos := scanner.Position()

	if from == 0 {
		n.base.VisitNested(scanner, input, from, to, match)
		scanner.Rewind(pos)
		return
	}

	holesPos := scanner.HolesPosition()

	scanner.MarkAsHole(0, from-1)
	match(n, from, from, true)

	n.base.VisitNested(scanner, input, from, to, match)

	scanner.RewindHoles(holesPos)
	scanner.Rewind(pos)
}

func (n *keep) Size() (int, bool) {
	return 0, false // TODO : fix it
}

func (n *keep) Copy() Node {
	return NewKeep()
}
