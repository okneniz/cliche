package node

import (
	"strings"
)

type alternation struct {
	variants []Node
	*base
}

func NewAlternation(variants []Node) Alternation {
	keys := make([]string, 0, len(variants))
	uniqVariants := make([]Node, 0, len(variants))

	c := make(map[string]struct{}, len(variants))

	for _, variant := range variants {
		key := ""

		Traverse(variant, func(x Node) bool {
			key += x.GetKey()
			return false
		})

		if _, exists := c[key]; exists {
			continue
		}

		c[key] = struct{}{}
		uniqVariants = append(uniqVariants, variant)
		keys = append(keys, key)
	}

	n := new(alternation)
	n.base = newBase("alternation<" + strings.Join(keys, "|") + ">")
	n.variants = uniqVariants

	return n
}

func (n *alternation) GetVariants() []Node {
	return n.variants
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
			scanner.Match(n, from, vTo, empty)
			onMatch(n, from, vTo, empty)

			nextFrom := nextFor(vTo, empty)
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
				if len(x.GetNestedNodes()) == 0 {
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

func (n *alternation) copyVariants() []Node {
	variants := make([]Node, len(n.variants))

	for i, x := range n.variants {
		variants[i] = x.Copy()
	}

	return variants
}

func (n *alternation) Copy() Node {
	return NewAlternation(n.copyVariants())
}

func (n *alternation) CopyAlternation() Alternation {
	return NewAlternation(n.copyVariants())
}
