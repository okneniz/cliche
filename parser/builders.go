package parser

import (
	"github.com/okneniz/cliche/encoding/unicode"
	"github.com/okneniz/cliche/node"
	c "github.com/okneniz/parsec/common"
)

type Combinator[T any, P any, S any, E error] func(
	c.Buffer[T, P],
) (S, E)

type Parser[S any] Combinator[rune, int, S, Error]

type ParserBuilder[S any] func(except ...rune) Parser[S]

func Const[T any](value T) ParserBuilder[T] {
	return func(_ ...rune) Parser[T] {
		return func(_ c.Buffer[rune, int]) (T, Error) {
			return value, nil
		}
	}
}

func NodeAsTable(
	makeParser ParserBuilder[node.Table],
) ParserBuilder[node.Node] {
	return func(except ...rune) Parser[node.Node] {
		parse := makeParser(except...)

		return func(buf c.Buffer[rune, int]) (node.Node, Error) {
			table, err := parse(buf)
			if err != nil {
				return nil, err
			}

			return node.NewForTable(table), nil
		}
	}
}

func RuneAsTable(
	makeParser ParserBuilder[rune],
) ParserBuilder[node.Table] {
	return func(except ...rune) Parser[node.Table] {
		parse := makeParser(except...)

		return func(buf c.Buffer[rune, int]) (node.Table, Error) {
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
	return func(except ...rune) Parser[rune] {
		parse := makeParser(except...)

		return func(buf c.Buffer[rune, int]) (rune, Error) {
			x, err := parse(buf)
			if err != nil {
				return -1, err
			}

			return rune(x), nil
		}
	}
}
