package node

import (
	"github.com/okneniz/cliche/structs"
)

type base struct {
	Expressions structs.Set[string] `json:"expressions,omitempty"`
	Nested      map[string]Node     `json:"nested,omitempty"`
}

func newBase() *base {
	n := new(base)
	n.Nested = make(map[string]Node)
	n.Expressions = structs.NewMapSet[string]()
	return n
}

func (n *base) GetNestedNodes() map[string]Node {
	return n.Nested
}

func (n *base) GetExpressions() structs.Set[string] {
	return n.Expressions
}

func (n *base) AddExpression(exp string) {
	n.Expressions.Add(exp)
}

func (n *base) IsLeaf() bool {
	return n.Expressions.Size() > 0
}

func (n *base) Merge(other Node) {
	for key, newNode := range other.GetNestedNodes() {
		if prev, exists := n.Nested[key]; exists {
			prev.Merge(newNode)
		} else {
			n.Nested[key] = newNode
		}
	}

	other.GetExpressions().AddTo(n.Expressions)
}

func (n *base) VisitNested(
	scanner Scanner,
	input Input,
	from, to int,
	onMatch Callback,
) {
	for _, nested := range n.Nested {
		nested.Visit(scanner, input, from, to, onMatch)
	}
}

func (n *base) NestedSize() (int, bool) {
	if len(n.Nested) == 0 {
		return 0, true
	}

	var size *int

	for _, child := range n.Nested {
		if x, fixedSize := child.Size(); fixedSize {
			if size != nil && *size != x {
				return 0, false
			}

			size = &x
		} else {
			return 0, false
		}
	}

	if size == nil {
		return 0, false
	}

	return *size, true
}
