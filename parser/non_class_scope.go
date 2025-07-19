package parser

import (
	"github.com/okneniz/cliche/encoding/unicode"
	"github.com/okneniz/cliche/node"
	c "github.com/okneniz/parsec/common"
)

type NonClassScope struct {
	items *Scope[node.Node]
}

func (scope *NonClassScope) Items() *Scope[node.Node] {
	return scope.items
}

func (scope *NonClassScope) makeParser(except ...rune) Parser[node.Node] {
	parseItem := scope.items.makeParser(except...)
	parseRune := NoneOf(except...)

	return func(buf c.Buffer[rune, int]) (node.Node, Error) {
		pos := buf.Position()

		item, itemErr := parseItem(buf)
		if itemErr == nil {
			return item, nil
		}

		buf.Seek(pos)

		r, runeErr := parseRune(buf)
		if runeErr == nil {
			return node.NewForTable(unicode.NewTable(r)), nil
		}

		buf.Seek(pos)

		return nil, MergeErrors(
			itemErr,
			runeErr,
		)
	}
}
