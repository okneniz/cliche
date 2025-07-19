package parser

import (
	"fmt"
	"strings"

	"github.com/okneniz/cliche/node"
	c "github.com/okneniz/parsec/common"
	"golang.org/x/exp/maps"
)

func makeParserTree[T any](
	cases map[string]ParserBuilder[T],
	except ...rune,
) Parser[T] {
	type branch struct {
		parser   Parser[T]
		children map[rune]*branch
	}

	root := new(branch)
	root.children = make(map[rune]*branch, 0)

	for cs, buildeParser := range cases {
		current := root

		for _, r := range cs {
			// TODO : handle conflicts

			child, exists := current.children[r]
			if !exists {
				child = &branch{
					children: make(map[rune]*branch),
				}

				current.children[r] = child
			}

			current = child
		}

		current.parser = buildeParser(except...)
	}

	var null T

	prefixes := maps.Keys(cases)
	errMessage := fmt.Sprintf(
		"one of %s",
		strings.Join(prefixes, ", "),
	)

	return func(buf c.Buffer[rune, int]) (T, *MultipleParsingError) {
		current := root.children
		start := buf.Position()

		var parserWithLongestPrefix Parser[T]

		for len(current) > 0 {
			pos := buf.Position()

			r, err := buf.Read(true)
			if err != nil {
				buf.Seek(pos)
				break
			}

			next, exists := current[r]
			if !exists {
				buf.Seek(pos)
				break
			}

			if next.parser != nil {
				parserWithLongestPrefix = next.parser
			}

			current = next.children
		}

		if parserWithLongestPrefix != nil {
			return parserWithLongestPrefix(buf)
		}

		return null, Expected(errMessage, start, c.NothingMatched)
	}
}

func makeGroupsParserTree(
	parseAlternation Parser[node.Alternation],
	cases map[string]GroupParserBuilder[node.Node],
	except ...rune,
) Parser[node.Node] {
	parseAny := NoneOf(except...) // to parse prefix rune by rune

	type branch struct {
		parser   Parser[node.Node]
		children map[rune]*branch
	}

	root := new(branch)
	root.children = make(map[rune]*branch, 0)

	for cs, buildeParser := range cases {
		current := root

		for _, r := range cs {
			// TODO : handle conflicts

			child, exists := current.children[r]
			if !exists {
				child = &branch{
					children: make(map[rune]*branch),
				}

				current.children[r] = child
			}

			current = child
		}

		current.parser = buildeParser(parseAlternation, except...)
	}

	prefixes := maps.Keys(cases)
	errMessage := fmt.Sprintf(
		"one of %s",
		strings.Join(prefixes, ", "),
	)

	return func(buf c.Buffer[rune, int]) (node.Node, *MultipleParsingError) {
		current := root.children
		start := buf.Position()

		var parserWithLongestPrefix Parser[node.Node]

		for len(current) > 0 {
			pos := buf.Position()

			r, err := parseAny(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			next, exists := current[r]
			if !exists {
				buf.Seek(pos)
				break
			}

			if next.parser != nil {
				parserWithLongestPrefix = next.parser
			}

			current = next.children
		}

		if parserWithLongestPrefix != nil {
			return parserWithLongestPrefix(buf)
		}

		return nil, Expected(errMessage, start, c.NothingMatched)
	}
}
