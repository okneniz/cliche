package parser

import (
	// "fmt"

	"github.com/okneniz/cliche/node"
	c "github.com/okneniz/parsec/common"
)

func newParserTree[T any](
	parse c.Combinator[rune, int, rune],
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

	return func(buf c.Buffer[rune, int]) (T, error) {
		current := root.children

		var parserWithLongesPrefix c.Combinator[rune, int, T]
		// parsedPrefix := []rune{} // TODO : remove it

		for len(current) > 0 {
			pos := buf.Position()

			r, err := parse(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			// parsedPrefix = append(parsedPrefix, r)

			next, exists := current[r]
			if !exists {
				buf.Seek(pos)
				break
			}

			if next.parser != nil {
				parserWithLongesPrefix = next.parser
			}

			current = next.children
		}

		// fmt.Println("parsed prefix", string(parsedPrefix))

		if parserWithLongesPrefix != nil {
			return parserWithLongesPrefix(buf)
		}

		return null, c.NothingMatched
	}
}

func NewGroupsParserTree(
	parse c.Combinator[rune, int, rune],
	parseAlternation c.Combinator[rune, int, node.Alternation],
	cases map[string]GroupParserBuilder[node.Node],
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	type branch struct {
		parser   c.Combinator[rune, int, node.Node]
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

	var null node.Node

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		current := root.children

		var parserWithLongesPrefix c.Combinator[rune, int, node.Node]
		// parsedPrefix := []rune{} // TODO : remove it

		for len(current) > 0 {
			pos := buf.Position()

			r, err := parse(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			// parsedPrefix = append(parsedPrefix, r)

			next, exists := current[r]
			if !exists {
				buf.Seek(pos)
				break
			}

			if next.parser != nil {
				parserWithLongesPrefix = next.parser
			}

			current = next.children
		}

		// fmt.Println("parsed group prefix", string(parsedPrefix))

		if parserWithLongesPrefix != nil {
			return parserWithLongesPrefix(buf)
		}

		return null, c.NothingMatched
	}
}
