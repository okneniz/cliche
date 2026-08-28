package node

import (
	"iter"
	"strings"

	"github.com/okneniz/cliche/span"
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
	bounds span.Interface,
	match Callback,
) {
	for _, sp := range n.VisitAlternation(scanner, input, bounds) {
		match(n, sp)
		nextFrom := nextFor(sp.To(), sp.Empty())
		next := span.Pair(nextFrom, bounds.To())
		n.base.VisitNested(scanner, input, next, match)
	}
}

// VisitAlternation - visit like container value (without nested nodes)
func (n *alternation) VisitAlternation(
	scanner Scanner,
	input Input,
	bounds span.Interface,
) iter.Seq2[Node, span.Interface] {
	return func(yield func(Node, span.Interface) bool) {
		pos := scanner.Position()
		stop := false

		for _, variant := range n.variants {
			emptVariant := true
			lastNotEmptyTo := bounds.From()

			variant.Visit(scanner, input, bounds, func(x Node, sp span.Interface) {
				if !sp.Empty() {
					lastNotEmptyTo = sp.To()
					emptVariant = false
				}

				if len(x.GetNestedNodes()) > 0 {
					return // find leaf
				}

				vsp := span.Empty(bounds.From())

				// последний node может быть пустым, например $
				// поэтому запоминаем последний span до него
				// чтобы границы подстроки были правильные
				if !emptVariant {
					vsp = span.Pair(bounds.From(), lastNotEmptyTo)
				}

				stop = stop || !yield(n, vsp)
			})

			scanner.Rewind(pos)

			if stop {
				break
			}
		}

		scanner.Rewind(pos)
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
