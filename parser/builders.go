package parser

import (
	"github.com/okneniz/cliche/encoding/unicode"
	"github.com/okneniz/cliche/node"
	c "github.com/okneniz/parsec/common"
)

type ParserBuilder[S any] func(except ...rune) c.Combinator[rune, int, S]

func Const[T any](value T) ParserBuilder[T] {
	return func(_ ...rune) c.Combinator[rune, int, T] {
		return func(_ c.Buffer[rune, int]) (T, c.Error[int]) {
			return value, nil
		}
	}
}

func TableAsClass(
	makeParser ParserBuilder[node.Table],
) ParserBuilder[node.Node] {
	return func(except ...rune) c.Combinator[rune, int, node.Node] {
		parse := makeParser(except...)

		return func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
			table, err := parse(buf)
			if err != nil {
				return nil, err
			}

			return node.NewClass(table), nil
		}
	}
}

func RuneAsTable(
	makeParser ParserBuilder[rune],
) ParserBuilder[node.Table] {
	return func(except ...rune) c.Combinator[rune, int, node.Table] {
		parse := makeParser(except...)

		return func(buf c.Buffer[rune, int]) (node.Table, c.Error[int]) {
			r, err := parse(buf)
			if err != nil {
				return nil, err
			}

			return unicode.NewTable(r), nil
		}
	}
}

func NumberAsRune(
	makeParser ParserBuilder[int],
) ParserBuilder[rune] {
	return func(except ...rune) c.Combinator[rune, int, rune] {
		parse := makeParser(except...)

		return func(buf c.Buffer[rune, int]) (rune, c.Error[int]) {
			x, err := parse(buf)
			if err != nil {
				return -1, err
			}

			return rune(x), nil
		}
	}
}
