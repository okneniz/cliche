package node

// https://www.regular-expressions.info/charclass.html
type simpleNode struct {
	key       string
	predicate func(rune) bool
	*nestedNode
}

func NewForTable(table Table) Node {
	return &simpleNode{
		key: table.String(),
		predicate: func(r rune) bool {
			return table.Include(r)
		},
		nestedNode: newNestedNode(),
	}
}

func (n *simpleNode) GetKey() string {
	return n.key
}

func (n *simpleNode) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *simpleNode) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	if n.predicate(input.ReadAt(from)) {
		pos := scanner.Position()
		groupsPos := scanner.GroupsPosition()

		scanner.Match(n, from, from, n.IsEnd(), false)
		onMatch(n, from, from, false)
		n.nestedNode.VisitNested(scanner, input, from+1, to, onMatch)

		scanner.Rewind(pos)
		scanner.RewindGroups(groupsPos)
	}
}

func (n *simpleNode) Size() (int, bool) {
	if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
		return 1 + nestedSize, true
	}

	return 0, false
}
