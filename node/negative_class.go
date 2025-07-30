package node

import (
	"unicode"
)

type negativeClass struct {
	table Table
	*base
}

func NewNegativeClass(table Table) Node {
	return &negativeClass{
		table: table,
		base:  newBase("^" + table.String()),
	}
}

func (n *negativeClass) Traverse(f func(Node)) {
	f(n)

	for _, x := range n.nested {
		x.Traverse(f)
	}
}

func (n *negativeClass) Visit(scanner Scanner, input Input, from, to int, onMatch Callback) {
	if from >= input.Size() {
		return
	}

	x := input.ReadAt(from)
	matched := false

	if scanner.OptionsInclude(ScanOptionCaseInsensetive) {
		matched = n.table.Include(unicode.ToUpper(x)) || n.table.Include(unicode.ToLower(x))
	} else {
		matched = n.table.Include(x)
	}

	if !matched {
		pos := scanner.Position()

		scanner.Match(n, from, from, n.IsLeaf(), false)
		onMatch(n, from, from, false)
		n.base.VisitNested(scanner, input, from+1, to, onMatch)

		scanner.Rewind(pos)
	}
}

func (n *negativeClass) Size() (int, bool) {
	if nestedSize, fixedSize := n.base.NestedSize(); fixedSize {
		return 1 + nestedSize, true
	}

	return 0, false
}

func (n *negativeClass) Copy() Node {
	return NewNegativeClass(n.table)
}
