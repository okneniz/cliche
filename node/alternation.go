package node

import (
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
	// order of variants is important
	Value     []Node `json:"value,omitempty"`
	lastNodes map[Node]struct{}
	*base
}

func NewAlternation(variants []Node) Alternation {
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
func (n *alternation) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
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
	// TODO : interup after first match?

	// POSIX ERE Alternation Returns The Longest Match
	//
	// In the tutorial topic about alternation, I explained that the regex engine will stop
	// as soon as it finds a matching alternative.
	// The POSIX standard, however, mandates that the longest match be returned.
	// When applying Set|SetValue to SetValue, a POSIX-compliant regex engine will
	// match SetValue entirely.
	// Even if the engine is a regex-directed NFA engine, POSIX requires that it
	// simulates DFA text-directed matching by trying all alternatives,
	// and returning the longest match, in this case SetValue.
	// A traditional NFA engine would match Set, as do all other regex flavors discussed
	// on this website.

	// A POSIX-compliant engine will still find the leftmost match.
	// If you apply Set|SetValue to Set or SetValue once, it will match Set.
	// The first position in the string is the leftmost position where our regex can find a
	//  valid match.
	// The fact that a longer match can be found further in the string is irrelevant.
	// If you apply the regex a second time, continuing at the first space in the string,
	// then SetValue will be matched.
	// A traditional NFA engine would match Set at the start of the string as the first match,
	// and Set at the start of the 3rd word in the string as the second match.

	// BUT

	// https://www.regular-expressions.info/alternation.html
	//
	// Remember That The Regex Engine Is Eager
	//
	// The consequence is that in certain situations, the order of the alternatives matters.
	// With expression "Get|GetValue|Set|SetValue" and string SetValue,
	// should be matched third variant - "Set"
	//
	// TODO : add test for if it possible

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

	if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
		return *size + nestedSize, true
	}

	return 0, false
}
