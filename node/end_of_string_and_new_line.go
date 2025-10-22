package node

type endOfStringAndNewLine struct {
	*base
}

func NewEndOfStringAndNewLine() Node {
	return &endOfStringAndNewLine{
		base: newBase("\\Z"),
	}
}

func (n *endOfStringAndNewLine) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if n.isEnd(input, from) || n.isEndAndNewLine(input, from) {
		pos := scanner.Position()
		onMatch(n, from, from, true)
		n.base.VisitNested(scanner, input, from, to, onMatch)
		scanner.Rewind(pos)
	}
}

func (n *endOfStringAndNewLine) isEnd(input Input, from int) bool {
	return from == input.Size()
}

func (n *endOfStringAndNewLine) isEndAndNewLine(input Input, from int) bool {
	last := input.Size() - 1
	if from != last {
		return false
	}

	return input.ReadAt(last) == '\n'
}

func (n *endOfStringAndNewLine) Size() (int, bool) {
	if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
		return nestedSize, true
	}

	return 0, false
}

func (n *endOfStringAndNewLine) Copy() Node {
	return NewEndOfStringAndNewLine()
}
