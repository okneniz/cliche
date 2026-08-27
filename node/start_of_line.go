package node

import "github.com/okneniz/cliche/span"

type startOfLine struct {
	*base
}

func NewStartOfLine() Node {
	return &startOfLine{
		base: newBase("^"),
	}
}

func (n *startOfLine) Visit(scanner Scanner, input Input, sp span.Interface, match Callback) {
	if sp.From() == 0 || n.isEndOfLine(input, sp.From()-1) { // TODO : check \n\r too
		pos := scanner.Position()
		match(n, span.Empty(sp.From()))
		n.base.VisitNested(scanner, input, sp, match)
		scanner.Rewind(pos)
	}
}

func (n *startOfLine) isEndOfLine(input Input, idx int) bool {
	if idx < 0 {
		return false
	}

	x := input.ReadAt(idx)

	switch x {
	case '\n':
		return true
	case '\r':
		if idx == 0 {
			return true
		}

		return input.ReadAt(idx-1) == '\n'
	default:
		return false
	}
}

func (n *startOfLine) Size() (int, bool) {
	if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
		return nestedSize, true
	}

	return 0, false
}

func (n *startOfLine) Copy() Node {
	return NewStartOfLine()
}
