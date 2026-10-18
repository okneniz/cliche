package parser

import (
	"github.com/okneniz/cliche/encoding/unicode"
	"github.com/okneniz/cliche/node"
	"github.com/okneniz/parsec"
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
) parsec.Combinator[rune, int, node.Node, *State] {
	parseItem := scope.items.makeParser(errMessage, except...)

	parseRune := parsec.Cast(
		parsec.NoneOf[rune, int, *State](
			errMessage,
			except...,
		),
		func(r rune) (node.Node, error) {
			return node.NewClass(unicode.NewTable(r)), nil
		},
	)

	return parsec.Choice(
		errMessage,
		parsec.Try(parseItem),
		parsec.Try(parseRune),
	)
}
