package cliche

import "github.com/okneniz/cliche/node"

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
		return variants[0]
	}

	return n
}

// RemovComments - remove comments from chains
func RemoveComments(n node.Node) node.Node {
	panic("not implemented yet")
}
