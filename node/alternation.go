package node

import (
	"strings"
)

type alternation struct {
	variants  []Node
	lastNodes map[Node]struct{}
	*base
}

func NewAlternation(variants []Node) Alternation {
	keys := make([]string, 0, len(variants))
	uniqVariants := make([]Node, 0, len(variants))

	for _, variant := range variants {
		key := ""
		variant.Traverse(func(x Node) { key += x.GetKey() })
		// TODO : keep only uniq variants by map
		uniqVariants = append(uniqVariants, variant)
		keys = append(keys, key)
	}

	n := new(alternation)
	n.base = newBase(strings.Join(keys, "|"))
	n.variants = uniqVariants
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

func (n *alternation) GetVariants() []Node {
	return n.variants
}

func (n *alternation) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.variants {
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
	n.VisitAlternation(
		scanner,
		input,
		from,
		to,
		func(_ Node, vFrom, vTo int, empty bool) bool {
			scanner.Match(n, from, vTo, n.IsLeaf(), empty)
			onMatch(n, from, vTo, empty)

			nextFrom := vTo
			if !empty {
				nextFrom++
			}

			n.base.VisitNested(scanner, input, nextFrom, to, onMatch)
			return false
		},
	)
}

// VisitAlternation - visit like container value (without nested nodes)
func (n *alternation) VisitAlternation(
	scanner Scanner,
	input Input,
	from,
	to int,
	onMatch AlternationCallback,
) {
	pos := scanner.Position()

	for _, variant := range n.variants {
		stop := false

		variant.Visit(
			scanner,
			input,
			from,
			to,
			func(x Node, vFrom, vTo int, empty bool) {
				if _, exists := n.lastNodes[x]; exists {
					vPos := scanner.Position()
					stop = onMatch(variant, vFrom, vTo, empty)
					scanner.Rewind(vPos)
				}
			},
		)

		scanner.Rewind(pos)

		if stop {
			break
		}
	}
}

// TODO : return list of sizes?
func (n *alternation) Size() (int, bool) {
	var size *int
	for _, variant := range n.variants {
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
