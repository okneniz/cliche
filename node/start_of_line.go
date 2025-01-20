package node

type startOfLine struct {
	*nestedNode
}

func NewStartOfLine() Node {
	return &startOfLine{
		nestedNode: newNestedNode(),
	}
}

func (n *startOfLine) GetKey() string {
	return "^"
}

func (n *startOfLine) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *startOfLine) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	if from == 0 || n.isEndOfLine(input, from-1) { // TODO : check \n\r too
		pos := scanner.Position()

		scanner.Match(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.VisitNested(scanner, input, from, to, onMatch)

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
	if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
		return nestedSize, true
	}

	return 0, false
}
