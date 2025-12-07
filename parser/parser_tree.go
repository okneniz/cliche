package parser

import (
	"fmt"
	"strings"

	"github.com/okneniz/cliche/node"
	c "github.com/okneniz/parsec/common"
	"golang.org/x/exp/maps"
)

func makeParserTree[T any](
	builders map[string]ParserBuilder[T],
	except ...rune,
) c.Combinator[rune, int, T] {
	prefixes := maps.Keys(builders)
	errMessage := fmt.Sprintf(
		"one of %s",
		strings.Join(prefixes, ", "),
	)

	cases := make(map[string]c.Combinator[rune, int, T])
	for key, makeParser := range builders {
		parser := makeParser(except...)
		cases[key] = parser
	}

	return c.MapTree(errMessage, cases, func(s string) []rune {
		return []rune(s)
	})
}

func makeGroupsParserTree(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	builders map[string]GroupParserBuilder[node.Node],
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	prefixes := maps.Keys(builders)
	errMessage := fmt.Sprintf(
		"one of %s",
		strings.Join(prefixes, ", "),
	)

	cases := make(map[string]c.Combinator[rune, int, node.Node])
	for key, makeParser := range builders {
		parser := makeParser(parseAlternation, except...)
		cases[key] = parser
	}

	return c.MapTree(errMessage, cases, func(s string) []rune {
		return []rune(s)
	})
}
