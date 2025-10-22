package node

type endOfLine struct {
	*base
}

func NewEndOfLine() Node {
	return &endOfLine{
		base: newBase("$"),
	}
}

func (n *endOfLine) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if n.isEndOfLine(input, from) {
		pos := scanner.Position()

		onMatch(n, from, from, true)
		n.base.VisitNested(scanner, input, from, to, onMatch)
		scanner.Rewind(pos)
	}
}

// TODO : check \n\r too
func (n *endOfLine) isEndOfLine(input Input, idx int) bool {
	if idx > input.Size() {
		return false
	}

	if idx == input.Size() {
		return true
	}

	return input.ReadAt(idx) == '\n'
}

func (n *endOfLine) Size() (int, bool) {
	if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
		return nestedSize, true
	}

	return 0, false
}

func (n *endOfLine) Copy() Node {
	return NewEndOfLine()
}
