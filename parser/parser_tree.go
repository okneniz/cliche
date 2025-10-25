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
) c.Combinator[rune, int, T] {
	type branch struct {
		parser   c.Combinator[rune, int, T]
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

	return func(buf c.Buffer[rune, int]) (T, c.Error[int]) {
		current := root.children
		start := buf.Position()

		var parserWithLongestPrefix c.Combinator[rune, int, T]

		for len(current) > 0 {
			pos := buf.Position()

			r, err := buf.Read(true)
			if err != nil {
				if seekErr := buf.Seek(pos); seekErr != nil {
					return null, c.NewParseError(
						pos,
						err.Error(),
					)
				}

				break
			}

			next, exists := current[r]
			if !exists {
				if seekErr := buf.Seek(pos); seekErr != nil {
					return null, c.NewParseError(
						pos,
						err.Error(),
					)
				}

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

		return null, c.NewParseError(
			start,
			errMessage,
		)
	}
}

func makeGroupsParserTree(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	cases map[string]GroupParserBuilder[node.Node],
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parseAny := c.NoneOf[rune, int](
		"expected predefined groups",
		except...,
	) // to parse prefix rune by rune

	type branch struct {
		parser   c.Combinator[rune, int, node.Node]
		children map[rune]*branch
	}

	root := new(branch)
	root.children = make(map[rune]*branch, 0)

	for cs, buildeParser := range cases {
		current := root

		for _, r := range cs {
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

	return func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
		current := root.children
		start := buf.Position()

		var parserWithLongestPrefix c.Combinator[rune, int, node.Node]

		for len(current) > 0 {
			pos := buf.Position()

			r, err := parseAny(buf)
			if err != nil {
				if seekErr := buf.Seek(pos); seekErr != nil {
					return nil, c.NewParseError(
						pos,
						err.Error(),
					)
				}

				break
			}

			next, exists := current[r]
			if !exists {
				if seekErr := buf.Seek(pos); seekErr != nil {
					return nil, c.NewParseError(
						pos,
						err.Error(),
					)
				}

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

		return nil, c.NewParseError(
			start,
			errMessage,
		)
	}
}
