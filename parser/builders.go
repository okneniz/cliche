package parser

import (
	"github.com/okneniz/cliche/encoding/unicode"
	"github.com/okneniz/cliche/node"
	c "github.com/okneniz/parsec/common"
)

type ParserBuilder[T any] func(except ...rune) c.Combinator[rune, int, T]

func Const[T any](value T) ParserBuilder[T] {
	return func(_ ...rune) c.Combinator[rune, int, T] {
		return func(_ c.Buffer[rune, int]) (T, error) {
			return value, nil
		}
	}
}

func NodeAsTable(parsetTable ParserBuilder[node.Table]) ParserBuilder[node.Node] {
	return func(except ...rune) c.Combinator[rune, int, node.Node] {
		parse := parsetTable(except...)

		return func(buf c.Buffer[rune, int]) (node.Node, error) {
			table, err := parse(buf)
			if err != nil {
				return nil, err
			}

			return node.NewForTable(table), nil
		}
	}
}

// TODO move to encoding package?
func RuneAsTable(makeParser ParserBuilder[rune]) ParserBuilder[node.Table] {
	return func(except ...rune) c.Combinator[rune, int, node.Table] {
		parse := makeParser(except...)

		return func(buf c.Buffer[rune, int]) (node.Table, error) {
			r, err := parse(buf)
			if err != nil {
				return nil, err
			}

			return unicode.NewTable(r), nil
		}
	}
}

func NumberAsRune(makeParser ParserBuilder[int]) ParserBuilder[rune] {
	return func(except ...rune) c.Combinator[rune, int, rune] {
		parse := makeParser(except...)

		return func(buf c.Buffer[rune, int]) (rune, error) {
			x, err := parse(buf)
			if err != nil {
				return -1, err
			}

			return rune(x), nil
		}
	}
}

func InvertTable(makeParser ParserBuilder[node.Table]) ParserBuilder[node.Table] {
	return func(except ...rune) c.Combinator[rune, int, node.Table] {
		parse := makeParser(except...)

		return func(buf c.Buffer[rune, int]) (node.Table, error) {
			table, err := parse(buf)
			if err != nil {
				return nil, err
			}

			return table.Invert(), nil
		}
	}
}
