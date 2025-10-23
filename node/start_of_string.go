package node

type startOfString struct {
	*base
}

func NewStartOfString() Node {
	return &startOfString{
		base: newBase("\\A"),
	}
}

func (n *startOfString) Visit(scanner Scanner, input Input, from, to int, match Callback) {
	if from != 0 {
		return
	}

	pos := scanner.Position()
	match(n, from, from, true)
	n.base.VisitNested(scanner, input, from, to, match)
	scanner.Rewind(pos)
}

func (n *startOfString) Size() (int, bool) {
	return 0, true
}

func (n *startOfString) Copy() Node {
	return NewStartOfString()
}
