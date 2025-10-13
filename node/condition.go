package node

import "fmt"

type condition struct {
	cond *Predicate
	yes  Node
	no   Node
	*base
	lastNodes map[Node]struct{}
}

type Predicate struct {
	key string
	fun func(scanner Scanner) bool
}

func NewPredicate(key string, f func(scanner Scanner) bool) *Predicate {
	return &Predicate{
		key: key,
		fun: f,
	}
}

func NewGuard(cond *Predicate, yes Node) Node {
	lastNodes := make(map[Node]struct{}, 1)

	Traverse(yes, func(x Node) bool {
		if len(x.GetNestedNodes()) == 0 {
			lastNodes[x] = struct{}{}
		}

		return false
	})

	return &condition{
		cond: cond,
		yes:  yes,
		base: newBase(
			fmt.Sprintf(
				"(?(%s)%s)",
				cond.key,
				yes.GetKey(),
			),
		),
		lastNodes: lastNodes,
	}
}

func NewCondition(cond *Predicate, yes Node, no Node) Node {
	lastNodes := make(map[Node]struct{}, 1)

	Traverse(yes, func(x Node) bool {
		if len(x.GetNestedNodes()) == 0 {
			lastNodes[x] = struct{}{}
		}

		return false
	})

	Traverse(no, func(x Node) bool {
		if len(x.GetNestedNodes()) == 0 {
			lastNodes[x] = struct{}{}
		}

		return false
	})

	return &condition{
		cond: cond,
		yes:  yes,
		no:   no,
		base: newBase(
			fmt.Sprintf(
				"(?(%s)%s|%s)",
				cond.key,
				yes.GetKey(),
				no.GetKey(),
			),
		),
		lastNodes: lastNodes,
	}
}

func (n *condition) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	pos := scanner.Position()

	if n.cond.fun(scanner) {
		n.yes.Visit(
			scanner,
			input,
			from,
			to,
			func(x Node, f, t int, empty bool) {
				if _, exists := n.lastNodes[x]; exists {
					scanner.Match(n, f, t, n.IsLeaf(), empty)
					onMatch(n, f, t, empty)
				}
			},
		)
	} else if n.no != nil {
		n.no.Visit(
			scanner,
			input,
			from,
			to,
			func(x Node, f, t int, empty bool) {
				if _, exists := n.lastNodes[x]; exists {
					scanner.Match(n, f, t, n.IsLeaf(), empty)
					onMatch(n, f, t, empty) // TODO : or onMatch(x, f, t, empty)?
				}
			},
		)
	}

	scanner.Rewind(pos)
}

func (n *condition) Size() (int, bool) {
	return 0, false
}

func (n *condition) Copy() Node {
	if n.no == nil {
		return NewGuard(n.cond, n.yes)
	}

	return NewCondition(n.cond, n.yes, n.no)
}
