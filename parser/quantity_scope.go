package parser

import (
	"github.com/okneniz/cliche/quantity"
	c "github.com/okneniz/parsec/common"
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
) c.Combinator[rune, int, *quantity.Quantity] {
	return scope.items.makeParser(
		errMessage,
		except...,
	)
}
