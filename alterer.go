package cliche

import "github.com/okneniz/cliche/node"

// todo : pass tree, not one node?
type Alter func(node node.Node) node.Node

func Compose(xs ...Alter) Alter {
	return func(n node.Node) node.Node {
		result := n

		for _, alter := range xs {
			result = alter(result)
		}

		return result
	}
}

// VariantToNode - alternation with one variant must work
// as this variant out of this alternation
func VariantToNode(n node.Alternation) node.Node {
	variants := n.GetVariants()

	if len(variants) == 0 {
		return variants[0] // node.Node
	}

	return n // node.Alternation
}

// RemovComments - remove comments from chains
func RemoveComments(n node.Node) node.Node {
	panic("what?")
}

// - translate one type node to another -> \d{3} -> \d\d\d ?
// TODO : add special type for quantifier?

// - empty chains to special node

// TODO : mark children as leaf on remove parent
