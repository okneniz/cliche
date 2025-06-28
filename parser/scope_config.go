package parser

import (
	c "github.com/okneniz/parsec/common"
)

type ScopeConfig[T any] struct {
	prefixes map[string]ParserBuilder[T]
	parsers  []ParserBuilder[T]
}

func NewScopeConfig[T any]() *ScopeConfig[T] {
	scope := new(ScopeConfig[T])
	scope.prefixes = make(map[string]ParserBuilder[T], 0)
	scope.parsers = make([]ParserBuilder[T], 0)
	return scope
}

func (scope *ScopeConfig[T]) Parse(
	builders ...ParserBuilder[T],
) *ScopeConfig[T] {
	scope.parsers = append(scope.parsers, builders...)
	return scope
}

func (scope *ScopeConfig[T]) WithPrefix(
	prefix string, builder ParserBuilder[T],
) *ScopeConfig[T] {
	scope.prefixes[prefix] = builder
	return scope
}

func (scope *ScopeConfig[T]) StringAsValue(
	prefix string, value T,
) *ScopeConfig[T] {
	return scope.WithPrefix(prefix, Const(value))
}

func (scope *ScopeConfig[T]) StringAsFunc(
	prefix string, nodeBuilder func() T,
) *ScopeConfig[T] {
	return scope.WithPrefix(
		prefix,
		func(_ ...rune) c.Combinator[rune, int, T] {
			return func(_ c.Buffer[rune, int]) (T, error) {
				return nodeBuilder(), nil
			}
		},
	)
}

func (scope *ScopeConfig[T]) makeParser(except ...rune) c.Combinator[rune, int, T] {
	// TODO : why ignore except?
	// TODO: don't ignore it - pass correct
	parseAny := c.Any[rune, int]() // to parse prefix rune by rune

	parseByPrefix := makeParserTree(
		parseAny,
		scope.prefixes,
		except...,
	)

	parsers := make([]c.Combinator[rune, int, T], 0, len(scope.parsers)+1)
	parsers = append(parsers, c.Try(parseByPrefix))

	for _, buildParser := range scope.parsers {
		parser := buildParser(except...)
		parsers = append(parsers, c.Try(parser))
	}

	return c.Choice(parsers...)
}
