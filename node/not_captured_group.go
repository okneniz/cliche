package node

import "fmt"

type notCapturedGroup struct {
	Value Alternation `json:"value,omitempty"`
	*nestedNode
}

func NewNotCapturedGroup(expression Alternation) Node {
	g := &notCapturedGroup{
		Value:      expression,
		nestedNode: newNestedNode(),
	}

	return g
}

func (n *notCapturedGroup) GetKey() string {
	return fmt.Sprintf("(?:%s)", n.Value.GetKey())
}

func (n *notCapturedGroup) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Nested {
		x.Traverse(f)
	}
}

func (n *notCapturedGroup) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	n.Value.VisitAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			pos := scanner.Position()

			scanner.Match(n, from, vTo, n.IsEnd(), false)
			onMatch(n, from, vTo, empty)
			n.nestedNode.VisitNested(scanner, input, vTo+1, to, onMatch)

			scanner.Rewind(pos)
		},
	)
}

func (n *notCapturedGroup) Size() (int, bool) {
	if size, fixedSize := n.Value.Size(); fixedSize {
		if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}
