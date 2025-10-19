package node

import "fmt"

type notCapturedGroup struct {
	value Alternation
	*base
}

var _ Container = new(notCapturedGroup)

func NewNotCapturedGroup(alt Alternation) Node {
	g := &notCapturedGroup{
		value: alt,
		base:  newBase(fmt.Sprintf("(?:%s)", alt.GetKey())),
	}

	return g
}

func (n *notCapturedGroup) GetValue() Node {
	return n.value
}

func (n *notCapturedGroup) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	n.value.VisitAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) bool {
			pos := scanner.Position()

			scanner.Match(n, from, vTo, empty)
			onMatch(n, from, vTo, empty)

			nextFrom := nextFor(vTo, empty)
			n.base.VisitNested(scanner, input, nextFrom, to, onMatch)

			scanner.Rewind(pos)
			return false
		},
	)
}

func (n *notCapturedGroup) Size() (int, bool) {
	if size, fixedSize := n.value.Size(); fixedSize {
		if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
			return size + nestedSize, true
		}
	}

	return 0, false
}

func (n *notCapturedGroup) Copy() Node {
	return NewNotCapturedGroup(n.value.CopyAlternation())
}
