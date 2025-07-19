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
	return scope.WithPrefix(prefix, func(_ ...rune) Parser[T] {
		return func(_ c.Buffer[rune, int]) (T, Error) {
			return nodeBuilder(), nil
		}
	},
	)
}

func (scope *Scope[T]) makeParser(except ...rune) Parser[T] {
	parseByPrefix := makeParserTree(
		scope.prefixes,
		except...,
	)

	parsers := make([]Parser[T], 0, len(scope.parsers)+1)
	parsers = append(parsers, parseByPrefix)

	for _, buildParser := range scope.parsers {
		parser := buildParser(except...)
		parsers = append(parsers, parser)
	}

	return func(buf c.Buffer[rune, int]) (T, Error) {
		pos := buf.Position()
		errs := make([]Error, 0, len(parsers))

		for _, parse := range parsers {
			value, valErr := parse(buf)
			if valErr == nil {
				return value, nil
			}

			buf.Seek(pos)

			errs = append(errs, valErr)
		}

		var t T
		return t, MergeErrors(errs...)
	}
}
