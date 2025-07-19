package parser

import (
	"fmt"

	"github.com/okneniz/cliche/node"
	c "github.com/okneniz/parsec/common"
)

type GroupParserBuilder[T any] func(
	parseAlternation Parser[node.Alternation],
	except ...rune,
) Parser[T]

// can use parser from common, need only for named group, not captured group, look around, conditions
type GroupScope struct {
	prefixes map[string]GroupParserBuilder[node.Node]
	parsers  []GroupParserBuilder[node.Node] // alternation wrapped to node
}

func (cfg *GroupScope) Parse(
	builders ...GroupParserBuilder[node.Node],
) *GroupScope {
	cfg.parsers = append(cfg.parsers, builders...)
	return cfg
}

func (cfg *GroupScope) ParsePrefix(
	prefix string, builder GroupParserBuilder[node.Node],
) *GroupScope {
	cfg.prefixes[prefix] = builder
	return cfg
}

func (cfg *GroupScope) makeParser(
	parseAlternation Parser[node.Alternation],
	except ...rune,
) Parser[node.Node] {

	parseScopeByPrefix := makeGroupsParserTree(
		parseAlternation,
		cfg.prefixes,
		except...,
	)

	parsers := make([]Parser[node.Node], 0, len(cfg.parsers)+1)
	parsers = append(parsers, parseScopeByPrefix)

	for _, buildParser := range cfg.parsers {
		f := buildParser(parseAlternation, except...)
		parsers = append(parsers, f)
	}

	return func(buf c.Buffer[rune, int]) (node.Node, *ParsingError) {
		pos := buf.Position()
		errs := make([]*ParsingError, 0, len(parsers))

		for i, parse := range parsers {
			fmt.Println("try to parse group", i)

			value, err := parse(buf)
			if err == nil {
				return value, nil
			}

			fmt.Println("group value parsing failed:", err)

			buf.Seek(pos)
			errs = append(errs, err)
		}

		return nil, MergeErrors(errs...)
	}
}
