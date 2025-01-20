package node

type startOfString struct {
	*nestedNode
}

func StartOfString() Node {
	return &startOfString{
		nestedNode: newNestedNode(),
	}
}

func (n *startOfString) GetKey() string {
	return "\\A"
}

func (n *startOfString) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *startOfString) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from == 0 {
		pos := scanner.Position()
		scanner.Match(n, from, from, n.IsEnd(), true)
		onMatch(n, from, from, true)
		n.nestedNode.VisitNested(scanner, input, from, to, onMatch)
		scanner.Rewind(pos)
	}
}

func (n *startOfString) Size() (int, bool) {
	return 0, true
}
