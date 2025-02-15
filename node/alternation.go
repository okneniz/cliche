package node

import (
	"strings"
)

type alternation struct {
	Value     []Node `json:"value,omitempty"`
	lastNodes map[Node]struct{}
	*base
}

func NewAlternation(variants []Node) Alternation {
	keys := make([]string, 0, len(variants))
	uniqVariants := make([]Node, 0, len(variants))

	for _, variant := range variants {
		key := ""
		variant.Traverse(func(x Node) { key += x.GetKey() })
		uniqVariants = append(uniqVariants, variant)
		keys = append(keys, key)
	}

	n := new(alternation)
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

func (n *alternation) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.Value {
		x.Traverse(f)
	}
}

// Visit - visit like node
func (n *alternation) Visit(
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
		func(_ Node, vFrom, vTo int, empty bool) bool {
			pos := scanner.Position()
			scanner.Match(n, from, vTo, n.IsLeaf(), empty)
			onMatch(n, from, vTo, empty)

			nextFrom := vTo
			if !empty {
				nextFrom++
			}

			n.base.VisitNested(scanner, input, nextFrom, to, onMatch)
			scanner.Rewind(pos)
			return false
		},
	)
}

// VisitAlternation - visit like group values
func (n *alternation) VisitAlternation(
	scanner Scanner,
	input Input,
	from, to int,
	onMatchVariant AlternationCallback,
) {
	n.visitVariants(
		scanner,
		input,
		from,
		to,
		func(variant Node, vFrom, vTo int, empty bool) bool {
			if _, exists := n.lastNodes[variant]; exists {
				return onMatchVariant(variant, vFrom, vTo, empty)
			}

			return false
		},
	)
}

func (n *alternation) visitVariants(
	scanner Scanner,
	input Input,
	from,
	to int,
	onMatch AlternationCallback,
) {
	for _, variant := range n.Value {
		stop := false

		variant.Visit(
			scanner,
			input,
			from,
			to,
			func(variant Node, vFrom, vTo int, empty bool) {
				stop = onMatch(variant, vFrom, vTo, empty)
			},
		)

		if stop {
			break
		}
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

	if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
		return *size + nestedSize, true
	}

	return 0, false
}
