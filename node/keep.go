package node

import "github.com/okneniz/cliche/span"

type keep struct {
	*base
}

func NewKeep() Node {
	return &keep{
		base: newBase("\\K"),
	}
}

func (n *keep) Visit(scanner Scanner, input Input, sp span.Interface, match Callback) {
	pos := scanner.Position()

	if sp.From() == 0 {
		n.base.VisitNested(scanner, input, sp, match)
		scanner.Rewind(pos)
		return
	}

	holesPos := scanner.HolesPosition()

	scanner.MarkAsHole(span.Pair(0, sp.From()-1))
	match(n, span.Empty(sp.From()))

	n.base.VisitNested(scanner, input, sp, match)

	scanner.RewindHoles(holesPos)
	scanner.Rewind(pos)
}

func (n *keep) Size() (int, bool) {
	return 0, false // TODO : fix it
}

func (n *keep) Copy() Node {
	return NewKeep()
}
