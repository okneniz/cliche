package parser

import (
	"github.com/okneniz/cliche/quantity"
	"github.com/okneniz/parsec"
)

type QuantityScope struct {
	items *Scope[*quantity.Quantity]
}

func (cfg *QuantityScope) Items() *Scope[*quantity.Quantity] {
	return cfg.items
}

func (scope *QuantityScope) makeParser(
	errMessage string,
	except ...rune,
) parsec.Combinator[rune, int, *quantity.Quantity, State] {
	return scope.items.makeParser(
		errMessage,
		except...,
	)
}
