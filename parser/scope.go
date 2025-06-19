package parser

import (
	"fmt"

	c "github.com/okneniz/parsec/common"
)

type ParserScope[T any] struct {
	prefixes map[string]ParserBuilder[T]
	parsers  []ParserBuilder[T]
}

func NewParserScope[T any]() *ParserScope[T] {
	scope := new(ParserScope[T])
	scope.prefixes = make(map[string]ParserBuilder[T], 0)
	scope.parsers = make([]ParserBuilder[T], 0)
	return scope
}

func (scope *ParserScope[T]) Parse(
	builders ...ParserBuilder[T],
) *ParserScope[T] {
	scope.parsers = append(scope.parsers, builders...)
	return scope
}

func (scope *ParserScope[T]) WithPrefix(
	prefix string, builder ParserBuilder[T],
) *ParserScope[T] {
	scope.prefixes[prefix] = builder
	return scope
}

func (scope *ParserScope[T]) StringAsValue(
	prefix string, value T,
) *ParserScope[T] {
	return scope.WithPrefix(prefix, Const(value))
}

func (scope *ParserScope[T]) StringAsFunc(
	prefix string, nodeBuilder func() T,
) *ParserScope[T] {
	return scope.WithPrefix(
		prefix,
		func(_ ...rune) c.Combinator[rune, int, T] {
			return func(_ c.Buffer[rune, int]) (T, error) {
				return nodeBuilder(), nil
			}
		},
	)
}

func (scope *ParserScope[T]) parser(except ...rune) c.Combinator[rune, int, T] {
	// TODO : why ignore except?
	// TODO: don't ignore it - pass correct
	parseAny := c.Any[rune, int]() // to parse prefix rune by rune

	parseScopeByPrefix := newParserTree(
		parseAny,
		scope.prefixes,
		except...,
	)

	parsers := make([]c.Combinator[rune, int, T], 0, len(scope.parsers)+1)
	parsers = append(parsers, c.Try(parseScopeByPrefix))

	for _, buildParser := range scope.parsers {
		parser := buildParser(except...)
		parsers = append(parsers, c.Try(parser))
	}

	return c.Choice(parsers...)
}

func (scope *ParserScope[T]) String() string {
	return fmt.Sprintf("%T{%v}", scope, scope.prefixes)
}
