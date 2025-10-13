package node

type endOfString struct {
	*base
}

func NewEndOfString() Node {
	return &endOfString{
		base: newBase("\\z"),
	}
}

func (n *endOfString) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from == input.Size() {
		pos := scanner.Position()
		scanner.Match(n, from, from, n.IsLeaf(), true)
		onMatch(n, from, from, true)
		n.base.VisitNested(scanner, input, from, to, onMatch)
		scanner.Rewind(pos)
	}
}

func (n *endOfString) Size() (int, bool) {
	if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
		return nestedSize, true
	}

	return 0, false
}

func (n *endOfString) Copy() Node {
	return NewEndOfString()
}
