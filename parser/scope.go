package parser

import (
	c "github.com/okneniz/parsec/common"
)

type Scope[T any] struct {
	prefixes map[string]ParserBuilder[T]
	parsers  []ParserBuilder[T]
}

func NewScope[T any]() *Scope[T] {
	scope := new(Scope[T])
	scope.prefixes = make(map[string]ParserBuilder[T], 0)
	scope.parsers = make([]ParserBuilder[T], 0)
	return scope
}

func (scope *Scope[T]) Parse(
	builders ...ParserBuilder[T],
) *Scope[T] {
	scope.parsers = append(scope.parsers, builders...)
	return scope
}

func (scope *Scope[T]) WithPrefix(
	prefix string, builder ParserBuilder[T],
) *Scope[T] {
	scope.prefixes[prefix] = builder
	return scope
}

func (scope *Scope[T]) StringAsValue(
	prefix string, value T,
) *Scope[T] {
	return scope.WithPrefix(
		prefix,
		Const(value),
	)
}

func (scope *Scope[T]) StringAsFunc(
	prefix string, nodeBuilder func() T,
) *Scope[T] {
	return scope.WithPrefix(prefix, func(_ ...rune) c.Combinator[rune, int, T] {
		return func(_ c.Buffer[rune, int]) (T, c.Error[int]) {
			return nodeBuilder(), nil
		}
	},
	)
}

func (scope *Scope[T]) makeParser(
	errMessage string,
	except ...rune,
) c.Combinator[rune, int, T] {
	parseByPrefix := makeParserTree(
		scope.prefixes,
		except...,
	)

	parsers := make([]c.Combinator[rune, int, T], 0, len(scope.parsers)+1)
	parsers = append(parsers, c.Try(parseByPrefix))

	for _, buildParser := range scope.parsers {
		parser := buildParser(except...)
		parsers = append(parsers, c.Try(parser))
	}

	return c.Choice(
		errMessage,
		parsers...,
	)
}
