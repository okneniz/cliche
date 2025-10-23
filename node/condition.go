package node

import "fmt"

type condition struct {
	cond *Predicate
	yes  Node
	no   Node
	*base
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
	}
}

func NewCondition(cond *Predicate, yes Node, no Node) Node {
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
	}
}

func (n *condition) Visit(scanner Scanner, input Input, from, to int, match Callback) {
	pos := scanner.Position()

	if n.cond.fun(scanner) {
		n.yes.Visit(
			scanner,
			input,
			from,
			to,
			func(x Node, f, t int, empty bool) {
				if len(x.GetNestedNodes()) == 0 {
					match(n, f, t, empty)
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
				if len(x.GetNestedNodes()) == 0 {
					match(n, f, t, empty)
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
