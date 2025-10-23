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

func (n *negativeClass) Visit(scanner Scanner, input Input, from, to int, match Callback) {
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

		match(n, from, from, false)
		n.base.VisitNested(scanner, input, from+1, to, match)

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
