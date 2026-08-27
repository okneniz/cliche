package node

import "github.com/okneniz/cliche/span"

type endOfStringAndNewLine struct {
	*base
}

func NewEndOfStringAndNewLine() Node {
	return &endOfStringAndNewLine{
		base: newBase("\\Z"),
	}
}

func (n *endOfStringAndNewLine) Visit(scanner Scanner, input Input, sp span.Interface, match Callback) {
	if n.isEnd(input, sp.From()) || n.isEndAndNewLine(input, sp.From()) {
		pos := scanner.Position()
		match(n, span.Empty(sp.From()))
		n.base.VisitNested(scanner, input, sp, match)
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
