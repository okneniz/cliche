package node

import (
	"github.com/okneniz/cliche/structs"
)

type base struct {
	key         string
	expressions structs.Set[string]
	nested      map[string]Node
}

func newBase(key string) *base {
	n := new(base)
	n.key = key
	n.nested = make(map[string]Node)
	n.expressions = structs.NewMapSet[string]()
	return n
}

func (n *base) GetKey() string {
	return n.key
}

func (n *base) GetNestedNodes() map[string]Node {
	return n.nested
}

func (n *base) GetExpressions() structs.Set[string] {
	return n.expressions
}

func (n *base) AddExpression(exp string) {
	n.expressions.Add(exp)
}

func (n *base) IsLeaf() bool {
	return n.expressions.Size() > 0
}

func (n *base) VisitNested(
	scanner Scanner,
	input Input,
	from, to int,
	match Callback,
) {
	for _, nested := range n.nested {
		nested.Visit(scanner, input, from, to, match)
	}
}

func (n *base) NestedSize() (int, bool) {
	if len(n.nested) == 0 {
		return 0, true
	}

	var size *int

	for _, child := range n.nested {
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
