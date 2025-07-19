package parser

import (
	c "github.com/okneniz/parsec/common"
)

type Scope[T any] struct {
	prefixes map[string]ParserBuilder[T, *MultipleParsingError]
	parsers  []ParserBuilder[T, *MultipleParsingError]
}

func NewScope[T any]() *Scope[T] {
	scope := new(Scope[T])
	scope.prefixes = make(map[string]ParserBuilder[T, *MultipleParsingError], 0)
	scope.parsers = make([]ParserBuilder[T, *MultipleParsingError], 0)
	return scope
}

func (scope *Scope[T]) Parse(
	builders ...ParserBuilder[T, *MultipleParsingError],
) *Scope[T] {
	scope.parsers = append(scope.parsers, builders...)
	return scope
}

func (scope *Scope[T]) WithPrefix(
	prefix string, builder ParserBuilder[T, *MultipleParsingError],
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
	return scope.WithPrefix(
		prefix,
		func(_ ...rune) Parser[T, *MultipleParsingError] {
			return func(_ c.Buffer[rune, int]) (T, *MultipleParsingError) {
				return nodeBuilder(), nil
			}
		},
	)
}

func (scope *Scope[T]) makeParser(
	except ...rune,
) Parser[T, *MultipleParsingError] {
	parseByPrefix := makeParserTree(
		scope.prefixes,
		except...,
	)

	parsers := make([]Parser[T, *MultipleParsingError], 0, len(scope.parsers)+1)
	parsers = append(parsers, parseByPrefix)

	for _, buildParser := range scope.parsers {
		parser := buildParser(except...)
		parsers = append(parsers, parser)
	}

	return func(buf c.Buffer[rune, int]) (T, *MultipleParsingError) {
		pos := buf.Position()
		errs := make([]*MultipleParsingError, 0, len(parsers))

		for _, parse := range parsers {
			value, err := parse(buf)
			if err == nil {
				return value, nil
			}

			buf.Seek(pos)

			errs = append(errs, err)
		}

		var t T
		return t, MergeErrors(errs...)
	}
}
