package node

type endOfString struct {
	*nestedNode
}

func EndOfString() Node {
	return &endOfString{
		nestedNode: newNestedNode(),
	}
}

func (n *endOfString) GetKey() string {
	return "\\z"
}

func (n *endOfString) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *endOfString) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from == input.Size() {
		pos := scanner.Position()
		scanner.Match(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.VisitNested(scanner, input, from, to, onMatch)
		scanner.Rewind(pos)
	}
}

func (n *endOfString) Size() (int, bool) {
	if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
		return nestedSize, true
	}

	return 0, false
}
