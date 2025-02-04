package node

import (
	"bytes"
	"strings"
)

type Alternation interface {
	Node

	VisitAlternation(
		scanner Scanner,
		input Input,
		from, to int,
		onMatch Callback,
	)
}

type alternation struct {
	// TODO : why not list, order is important
	Value     map[string]Node   `json:"value,omitempty"`
	lastNodes map[Node]struct{} // TODO : interface like key, is it ok?
	*nestedNode
}

func NewAlternation(variants []Node) *alternation {
	n := new(alternation)
	n.Value = make(map[string]Node, len(variants))
	n.lastNodes = make(map[Node]struct{}, len(variants))
	n.nestedNode = newNestedNode()

	variantKey := bytes.NewBuffer(nil)

	for _, variant := range variants {
		variant.Traverse(func(x Node) {
			variantKey.WriteString(x.GetKey())

			if len(x.GetNestedNodes()) == 0 {
				n.lastNodes[x] = struct{}{}
			}
		})

		x := variantKey.String()
		n.Value[x] = variant
		variantKey.Reset()
	}

	variantKey.Reset()

	return n
}

func (n *alternation) GetKey() string {
	variantKeys := make([]string, 0, len(n.Value))

	for _, variant := range n.Value {
		variantKeys = append(variantKeys, variant.GetKey())
	}

	return strings.Join(variantKeys, ",")
}

func (n *alternation) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Value {
		x.Traverse(f)
	}
}

// TODO : check it without groups too
func (n *alternation) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	n.visitVariants(
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

// TODO : как бы удалить и оставить только Visit?
func (n *alternation) VisitAlternation(
	scanner Scanner,
	input Input,
	from, to int,
	onMatch Callback,
) {
	n.visitVariants(
		scanner,
		input,
		from,
		to,
		func(variant Node, vFrom, vTo int, empty bool) {
			if _, exists := n.lastNodes[variant]; exists {
				onMatch(variant, vFrom, vTo, empty)
			}
		},
	)
}

func (n *alternation) visitVariants(scanner Scanner, input Input, from, to int, onMatch Callback) {
	for _, variant := range n.Value {
		variant.Visit(scanner, input, from, to, onMatch)
	}
}

func (n *alternation) Size() (int, bool) {
	var size *int
	for _, variant := range n.Value {
		if x, fixedSize := variant.Size(); fixedSize {
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

	if nestedSize, fixedSize := n.nestedNode.NestedSize(); fixedSize {
		return *size + nestedSize, true
	}

	return 0, false
}
