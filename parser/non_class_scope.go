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
	errMessage string,
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parseItem := scope.items.makeParser(errMessage, except...)

	parseRune := c.Cast(
		c.NoneOf[rune, int](
			errMessage,
			except...,
		),
		func(r rune) (node.Node, error) {
			return node.NewClass(unicode.NewTable(r)), nil
		},
	)

	return c.Choice(
		errMessage,
		c.Try(parseItem),
		c.Try(parseRune),
	)
}
