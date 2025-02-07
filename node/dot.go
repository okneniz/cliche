package node

// not simple node with table because §diferent behaviour for different scan options
// TODO : add something to empty json value, and in another spec symbols
type dot struct {
	*nestedNode
}

func NewDot() Node {
	return &dot{
		nestedNode: newNestedNode(),
	}
}

func (n *dot) GetKey() string {
	return "."
}

func (n *dot) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *dot) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	if input.ReadAt(from) != '\n' {
		pos := scanner.Position()

		scanner.Match(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.VisitNested(scanner, input, from+1, to, onMatch)

		scanner.Rewind(pos)
	}
}

func (n *dot) Size() (int, bool) {
	if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
		return 1 + nestedSize, true
	}

	return 0, false
}
