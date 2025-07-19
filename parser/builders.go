package parser

import (
	"github.com/okneniz/cliche/encoding/unicode"
	"github.com/okneniz/cliche/node"
	c "github.com/okneniz/parsec/common"
)

type Combinator[T any, P any, S any, E error] func(
	c.Buffer[T, P],
) (S, E)

type Parser[S any, E error] Combinator[rune, int, S, E]

type ParserBuilder[S any, E error] func(
	except ...rune,
) Parser[S, E]

func Const[T any](value T) ParserBuilder[T, *MultipleParsingError] {
	return func(_ ...rune) Parser[T, *MultipleParsingError] {
		return func(_ c.Buffer[rune, int]) (T, *MultipleParsingError) {
			return value, nil
		}
	}
}

func NodeAsTable(
	makeParser ParserBuilder[node.Table, *MultipleParsingError],
) ParserBuilder[node.Node, *MultipleParsingError] {
	return func(except ...rune) Parser[node.Node, *MultipleParsingError] {
		parse := makeParser(except...)

		return func(buf c.Buffer[rune, int]) (node.Node, *MultipleParsingError) {
			table, err := parse(buf)
			if err != nil {
				return nil, err
			}

			return node.NewForTable(table), nil
		}
	}
}

func RuneAsTable(
	makeParser ParserBuilder[rune, *MultipleParsingError],
) ParserBuilder[node.Table, *MultipleParsingError] {
	return func(except ...rune) Parser[node.Table, *MultipleParsingError] {
		parse := makeParser(except...)

		return func(buf c.Buffer[rune, int]) (node.Table, *MultipleParsingError) {
			r, err := parse(buf)
			if err != nil {
				return nil, err
			}

			return unicode.NewTable(r), nil
		}
	}
}

func NumberAsRune(
	makeParser ParserBuilder[int, *MultipleParsingError],
) ParserBuilder[rune, *MultipleParsingError] {
	return func(except ...rune) Parser[rune, *MultipleParsingError] {
		parse := makeParser(except...)

		return func(buf c.Buffer[rune, int]) (rune, *MultipleParsingError) {
			x, err := parse(buf)
			if err != nil {
				return -1, err
			}

			return rune(x), nil
		}
	}
}
