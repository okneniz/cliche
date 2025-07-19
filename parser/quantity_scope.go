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
	except ...rune,
) Parser[*quantity.Quantity, *MultipleParsingError] {
	parse := scope.items.makeParser(except...)

	return func(
		buf c.Buffer[rune, int],
	) (*quantity.Quantity, *MultipleParsingError) {
		pos := buf.Position()

		q, err := parse(buf)
		if err != nil {
			buf.Seek(pos)
			return nil, err
		}

		return q, nil
	}
}
