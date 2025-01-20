package node

import (
	"github.com/okneniz/cliche/structs"
)

// TODO : rename to base?
type nestedNode struct {
	Expressions structs.Set[string] `json:"expressions,omitempty"`

	// TODO : rename to Next?
	Nested map[string]Node `json:"nested,omitempty"`
}

func newNestedNode() *nestedNode {
	n := new(nestedNode)
	n.Nested = make(map[string]Node)
	n.Expressions = structs.NewMapSet[string]()
	return n
}

func (n *nestedNode) GetNestedNodes() map[string]Node {
	return n.Nested
}

func (n *nestedNode) GetExpressions() structs.Set[string] {
	return n.Expressions
}

func (n *nestedNode) AddExpression(exp string) {
	n.Expressions.Add(exp)
}

func (n *nestedNode) IsEnd() bool {
	return n.Expressions.Size() > 0
}

func (n *nestedNode) Merge(other Node) {
	for key, newNode := range other.GetNestedNodes() {
		if prev, exists := n.Nested[key]; exists {
			prev.Merge(newNode)
		} else {
			n.Nested[key] = newNode
		}
	}

	other.GetExpressions().AddTo(n.Expressions)
}

func (n *nestedNode) VisitNested(
	scanner Scanner,
	input Input,
	from, to int,
	onMatch Callback,
) {
	for _, nested := range n.Nested {
		// pos := scanner.Position()
		// groupsPos := scanner.GroupsPosition()
		// namedGroupPos := scanner.NamedGroupsPosition()

		nested.Visit(scanner, input, from, to, onMatch)

		// scanner.Rewind(pos)
		// scanner.RewindGroups(groupsPos)
		// scanner.RewindNamedGroups(namedGroupPos)
	}
}

func (n *nestedNode) NestedSize() (int, bool) {
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
