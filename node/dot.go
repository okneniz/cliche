package node

import "github.com/okneniz/cliche/span"

type dot struct {
	*base
}

func NewDot() Node {
	return &dot{
		base: newBase("."),
	}
}

func (n *dot) Visit(scanner Scanner, input Input, sp span.Interface, match Callback) {
	if sp.From() >= input.Size() {
		return
	}

	x := input.ReadAt(sp.From())
	matched := false

	if scanner.OptionsInclude(ScanOptionMultiline) {
		matched = true
	} else {
		matched = x != '\n'
	}

	if matched {
		pos := scanner.Position()

		match(n, span.Pair(sp.From(), sp.From()))
		next := span.Pair(sp.From()+1, sp.To())
		n.base.VisitNested(scanner, input, next, match)
		scanner.Rewind(pos)
	}
}

func (n *dot) Size() (int, bool) {
	if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
		return 1 + nestedSize, true
	}

	return 0, false
}

func (n *dot) Copy() Node {
	return NewDot()
}
