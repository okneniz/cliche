package node

type dot struct {
	*base
}

func NewDot() Node {
	return &dot{
		base: newBase("."),
	}
}

func (n *dot) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)
	matched := false

	if scanner.OptionsInclude(ScanOptionMultiline) {
		matched = true
	} else {
		matched = x != '\n'
	}

	if matched {
		pos := scanner.Position()

		onMatch(n, from, from, false)
		n.base.VisitNested(scanner, input, from+1, to, onMatch)
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
