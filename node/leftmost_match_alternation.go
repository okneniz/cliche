package node

import (
	"strings"
)

// leftmostMatchAlternation - POSIX compliant alternation.
// It has to continue trying all alternatives even after a match is found,
// in order to find the longest one.
type leftmostMatchAlternation struct {
	Value     []Node `json:"value,omitempty"`
	lastNodes map[Node]struct{}
	*base
}

func NewLeftmostAlternation(variants []Node) Alternation {
	keys := make([]string, 0, len(variants))
	uniqVariants := make([]Node, 0, len(variants))
	cache := make(map[string]struct{})

	for _, variant := range variants {
		key := variant.GetKey()

		if _, exists := cache[key]; exists {
			continue
		}

		uniqVariants = append(uniqVariants, variant)
		keys = append(keys, key)
	}

	n := new(leftmostMatchAlternation)
	n.base = newBase(strings.Join(keys, "|"))
	n.Value = uniqVariants
	n.lastNodes = make(map[Node]struct{}, len(uniqVariants))

	for _, variant := range uniqVariants {
		variant.Traverse(func(x Node) {
			if len(x.GetNestedNodes()) == 0 {
				n.lastNodes[x] = struct{}{}
			}
		})
	}

	return n
}

func (n *leftmostMatchAlternation) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Value {
		x.Traverse(f)
	}
}

// Visit - visit like node
func (n *leftmostMatchAlternation) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	n.visitVariants(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) {
			pos := scanner.Position()
			scanner.Match(n, from, vTo, n.IsLeaf(), false)
			onMatch(n, from, vTo, empty)
			n.base.VisitNested(scanner, input, vTo+1, to, onMatch)
			scanner.Rewind(pos)
		},
	)
}

// VisitAlternation - visit like group values
func (n *leftmostMatchAlternation) VisitAlternation(
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

func (n *leftmostMatchAlternation) visitVariants(
	scanner Scanner,
	input Input,
	from, to int,
	onMatch Callback,
) {
	for _, variant := range n.Value {
		matched := false

		variant.Visit(scanner, input, from, to, func(variant Node, vFrom, vTo int, empty bool) {
			onMatch(variant, vFrom, vTo, empty)
			matched = true
		})

		if matched {
			break
		}
	}
}

func (n *leftmostMatchAlternation) Size() (int, bool) {
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

	if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
		return *size + nestedSize, true
	}

	return 0, false
}
