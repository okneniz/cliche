package node

type endOfLine struct {
	*nestedNode
}

func NewEndOfLine() Node {
	return &endOfLine{
		nestedNode: newNestedNode(),
	}
}

func (n *endOfLine) GetKey() string {
	return "$"
}

func (n *endOfLine) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *endOfLine) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if n.isEndOfLine(input, from) {
		pos := scanner.Position()

		scanner.Match(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.VisitNested(scanner, input, from, to, onMatch)

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
	if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
		return nestedSize, true
	}

	return 0, false
}
