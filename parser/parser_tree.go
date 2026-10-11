package parser

import (
	"maps"
	"slices"
	"strings"

	"github.com/okneniz/cliche/node"
	"github.com/okneniz/parsec"
)

func makeParserTree[T any](
	builders map[string]ParserBuilder[T],
	except ...rune,
) parsec.Combinator[rune, int, T, State] {
	prefixes := slices.Collect(maps.Keys(builders))
	errMessage := "one of " + strings.Join(prefixes, ", ")

	cases := make(map[string]parsec.Combinator[rune, int, T, State])
	for key, makeParser := range builders {
		parser := makeParser(except...)
		cases[key] = parser
	}

	return parsec.MapTree(errMessage, cases, func(s string) []rune {
		return []rune(s)
	})
}

func makeGroupsParserTree(
	parseAlternation parsec.Combinator[rune, int, node.Alternation, State],
	builders map[string]GroupParserBuilder[node.Node],
	except ...rune,
) parsec.Combinator[rune, int, node.Node, State] {
	prefixes := slices.Collect(maps.Keys(builders))
	errMessage := "one of " + strings.Join(prefixes, ", ")

	cases := make(map[string]parsec.Combinator[rune, int, node.Node, State])
	for key, makeParser := range builders {
		parser := makeParser(parseAlternation, except...)
		cases[key] = parser
	}

	return parsec.MapTree(errMessage, cases, func(s string) []rune {
		return []rune(s)
	})
}
