package node

import (
	"fmt"

	"golang.org/x/exp/maps"
)

func Unify(start Node) []Node {
	changed := true
	roots := []Node{start}

	for changed {
		changed = false

		// change nested
		for _, root := range roots {
			traverse(root, func(parent, nested Node) {
				key := nested.GetKey()

				newNested, ok := unify(nested)
				if ok {
					changed = true

					delete(parent.GetNestedNodes(), key)

					for _, newNode := range newNested {
						newKey := newNode.GetKey()
						parent.GetNestedNodes()[newKey] = newNode
					}
				}
			})
		}

		newRoots := make([]Node, 0, len(roots))

		// change roots
		for _, root := range roots {
			unifiedRoots, ok := unify(root)
			if ok {
				changed = true
				newRoots = append(newRoots, unifiedRoots...)
			} else {
				newRoots = append(newRoots, root)
			}
		}

		roots = newRoots
	}

	return roots
}

func traverse(n Node, f func(parent, child Node)) {
	for _, nested := range n.GetNestedNodes() {
		traverse(nested, f)
		f(n, nested)
	}
}

func unify(node Node) ([]Node, bool) {
	switch n := node.(type) {
	case *comment:
		return removeComments(n)
	case *alternation:
		x, y := simplifyAlternation(n)
		// fmt.Println("wtf", x[0].GetKey(), y)
		return x, y
	// TODO : пока пораждает лишние метчи
	// нужно придумать способ группировки результата получше
	//
	// case *quantifier:
	// 	return stretchQuantificator(n)
	default:
		return nil, false
	}
}

// TODO : добавить удаление комментариев в хвосте тоже
func removeComments(c *comment) ([]Node, bool) {
	nested := c.GetNestedNodes()
	if len(nested) != 0 {
		return maps.Values(nested), true
	}

	// transform to empty? or just remove from tree?

	return nil, false
}

func simplifyAlternation(alt *alternation) ([]Node, bool) {
	if len(alt.variants) == 1 {
		variant := alt.variants[0]

		if alt.IsLeaf() {
			Traverse(variant, func(x Node) bool {
				if len(x.GetNestedNodes()) == 0 { // add to leaf
					moveExpressions(alt, x)
				}

				return false
			})
		}

		moveNested(alt, variant)
		return []Node{variant}, true
	}

	return nil, false
}

func moveExpressions(from, to Node) {
	from.GetExpressions().AddTo(to.GetExpressions())
}

func moveNested(from, to Node) {
	for key, nested := range from.GetNestedNodes() {
		to.GetNestedNodes()[key] = nested
	}
}

// для этого нужен специальный случай quantity который не имеет From
// {1,} = lead{1} -> {,}
// {3,} = {1} -> {1} -> leaf{1} -> leaf{1} -> {,}
//
// тоже самое для reluctant и possessive quantificator нужно будет делать как-то иначе
// и видимо другой тип node должен быть

// для этого доработок не нужно
// {3} = leaf{1} -> leaf{1} -> leaf{1}
// {2,4} = {1} -> leaf{1} -> leaf{1} -> leaf{1}
//
//nolint:unused // it's ok
func stretchQuantificator(q *quantifier) ([]Node, bool) {
	// fmt.Println("stretch quantifier", q.quantity.Optional())
	if q.quantity.Optional() {
		return nil, false
	}

	to, fixed := q.quantity.To()
	if !fixed {
		return nil, false
	}

	// это можно делать только с одиночными выражениями, не цепочками
	// у который нет NestedNode
	//
	// \d{3}
	//
	// с группами нельзя так как афектит capturing и holes

	// start loop with index
	// if quantity include index - it's leaf = add expression

	first := q.value.Copy()
	last := first

	if q.quantity.Include(1) {
		moveExpressions(q, last)
	}

	// if leaf / no leaf

	// fmt.Println("first", first, to, 2 <= to, q.quantity)

	for i := 2; i <= to; i++ {
		next := first.Copy()

		if q.quantity.Include(i) {
			moveExpressions(q, next)
		}

		last.GetNestedNodes()[next.GetKey()] = next

		// fmt.Println(last.GetKey(), "->", next.GetKey())
		last = next

		// newKey := ""
		// first.Traverse(func(x Node) { newKey += x.GetKey() })
		// fmt.Println("newKey", newKey)
	}

	moveNested(q, last)

	fmt.Println("change quantifier", q.GetKey())

	// newKey := ""
	// first.Traverse(func(x Node) { newKey += x.GetKey() })
	// fmt.Println("to", newKey)

	return []Node{first}, true
}
