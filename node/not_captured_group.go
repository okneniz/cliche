package node

import (
	"fmt"

	"github.com/okneniz/cliche/span"
)

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

func (n *notCapturedGroup) Visit(scanner Scanner, input Input, bounds span.Interface, match Callback) {
	for _, sp := range n.value.VisitAlternation(scanner, input, bounds) {
		pos := scanner.Position()

		match(n, sp)

		nextFrom := nextFor(sp.To(), sp.Empty())
		next := span.Pair(nextFrom, bounds.To())
		n.base.VisitNested(scanner, input, next, match)

		scanner.Rewind(pos)
	}
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
