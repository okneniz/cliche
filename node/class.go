package node

import (
	"unicode"

	"github.com/okneniz/cliche/span"
)

type class struct {
	table Table
	*base
}

func NewClass(table Table) Node {
	return &class{
		table: table,
		base:  newBase(table.String()),
	}
}

func (n *class) Visit(scanner Scanner, input Input, sp span.Interface, match Callback) {
	if sp.From() >= input.Size() {
		return
	}

	x := input.ReadAt(sp.From())
	matched := false

	if scanner.OptionsInclude(ScanOptionCaseInsensitive) {
		matched = n.table.Include(unicode.ToUpper(x)) || n.table.Include(unicode.ToLower(x))
	} else {
		matched = n.table.Include(x)
	}

	if matched {
		pos := scanner.Position()
		match(n, span.Pair(sp.From(), sp.From()))

		nextFrom := sp.From() + 1
		nextTo := nextFrom
		if sp.To() > nextTo {
			nextTo = sp.To()
		}

		next := span.Pair(nextFrom, nextTo)
		n.VisitNested(scanner, input, next, match)
		scanner.Rewind(pos)
	}
}

func (n *class) Size() (int, bool) {
	if nestedSize, fixedSize := n.NestedSize(); fixedSize {
		return 1 + nestedSize, true
	}

	return 0, false
}

func (n *class) Copy() Node {
	return NewClass(n.table)
}
