package parser

import (
	"github.com/okneniz/cliche/node"
	c "github.com/okneniz/parsec/common"
)

// can use parser from common, need only for named group, captured group, etc
type GroupParserConfig struct {
	prefixes map[string]GroupParserBuilder[node.Node]
	parsers  []GroupParserBuilder[node.Node] // alternation wrapped to node
}

type GroupParserBuilder[T any] func(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	except ...rune,
) c.Combinator[rune, int, T]

func (cfg *GroupParserConfig) Parse(
	builders ...GroupParserBuilder[node.Node],
) *GroupParserConfig {
	cfg.parsers = append(cfg.parsers, builders...)
	return cfg
}

func (cfg *GroupParserConfig) ParsePrefix(
	prefix string, builder GroupParserBuilder[node.Node],
) *GroupParserConfig {
	cfg.prefixes[prefix] = builder
	return cfg
}

func (cfg *GroupParserConfig) parser(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parseAny := c.NoneOf[rune, int](except...) // to parse prefix rune by rune

	parseScopeByPrefix := NewGroupsParserTree(
		parseAny,
		parseAlternation,
		cfg.prefixes,
		except...,
	)

	parsers := make([]c.Combinator[rune, int, node.Node], 0, len(cfg.parsers)+1)
	parsers = append(parsers, c.Try(parseScopeByPrefix))

	for _, buildParser := range cfg.parsers {
		parser := buildParser(parseAlternation, except...)
		parsers = append(parsers, c.Try(parser))
	}

	return c.Choice(parsers...)
}
