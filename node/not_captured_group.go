package node

import "fmt"

type notCapturedGroup struct {
	Value Alternation `json:"value,omitempty"`
	*base
}

func NewNotCapturedGroup(alt Alternation) Node {
	g := &notCapturedGroup{
		Value: alt,
		base:  newBase(fmt.Sprintf("(?:%s)", alt.GetKey())),
	}

	return g
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
		func(_ Node, vFrom, vTo int, empty bool) bool {
			pos := scanner.Position()

			scanner.Match(n, from, vTo, n.IsLeaf(), empty)
			onMatch(n, from, vTo, empty)
			n.base.VisitNested(scanner, input, vTo+1, to, onMatch)

			scanner.Rewind(pos)
			return false
		},
	)
}

func (n *notCapturedGroup) Size() (int, bool) {
	if size, fixedSize := n.Value.Size(); fixedSize {
		if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}
