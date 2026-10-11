package parser

import (
	"github.com/okneniz/cliche/encoding/unicode"
	"github.com/okneniz/cliche/node"
	"github.com/okneniz/parsec"
)

type ParserBuilder[T any] func(except ...rune) parsec.Combinator[rune, int, T, State]

func Const[T any](value T) ParserBuilder[T] {
	return func(_ ...rune) parsec.Combinator[rune, int, T, State] {
		return func(_ State, _ parsec.Buffer[rune, int]) (T, parsec.Error[int]) {
			return value, nil
		}
	}
}

func TableAsClass(
	makeParser ParserBuilder[node.Table],
) ParserBuilder[node.Node] {
	return func(except ...rune) parsec.Combinator[rune, int, node.Node, State] {
		parse := makeParser(except...)

		return func(state State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
			table, err := parse(state, buf)
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
	return func(except ...rune) parsec.Combinator[rune, int, node.Table, State] {
		parse := makeParser(except...)

		return func(state State, buf parsec.Buffer[rune, int]) (node.Table, parsec.Error[int]) {
			r, err := parse(state, buf)
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
	return func(except ...rune) parsec.Combinator[rune, int, rune, State] {
		parse := makeParser(except...)

		return func(state State, buf parsec.Buffer[rune, int]) (rune, parsec.Error[int]) {
			x, err := parse(state, buf)
			if err != nil {
				return -1, err
			}

			return rune(x), nil
		}
	}
}
