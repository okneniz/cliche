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

func (scope *NonClassScope) makeParser(
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parseItem := scope.items.makeParser(except...)
	parseRune := c.NoneOf[rune, int](except...)

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		pos := buf.Position()

		item, err := parseItem(buf)
		if err == nil {
			return item, nil
		}

		buf.Seek(pos)

		r, err := parseRune(buf)
		if err == nil {
			return node.NewForTable(unicode.NewTable(r)), nil
		}

		buf.Seek(pos)

		return nil, err
	}
}
