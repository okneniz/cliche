package parser

import (
	c "github.com/okneniz/parsec/common"

	"github.com/okneniz/cliche/encoding/unicode"
	"github.com/okneniz/cliche/node"
)

type ClassScope struct {
	runes *Scope[rune]
	items *Scope[node.Table]
}

var (
	parseNegativePrefix = c.Try(c.Eq[rune, int](
		"expected '^' as prefix for negative class",
		'^',
	))
)

func (scope *ClassScope) Runes() *Scope[rune] {
	return scope.runes
}

func (scope *ClassScope) Items() *Scope[node.Table] {
	return scope.items
}

func (scope *ClassScope) makeParser() c.Combinator[rune, int, node.Node] {
	parseTable := scope.makeTableParser()

	parseLeftSquare := c.Eq[rune, int](
		"expected left square as begining of character class",
		'[',
	)

	parseRightSquare := c.Eq[rune, int](
		"expected right square as ending of character class",
		']',
	)

	return c.Try(func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
		isNegative := true

		_, err := parseLeftSquare(buf)
		if err != nil {
			return nil, err
		}

		_, err = parseNegativePrefix(buf)
		if err != nil {
			isNegative = false
		}

		table, err := parseTable(buf)
		if err != nil {
			return nil, err
		}

		_, err = parseRightSquare(buf)
		if err != nil {
			return nil, err
		}

		if isNegative {
			return node.NewNegativeClass(table), nil
		}

		return node.NewClass(table), nil
	})
}

func (scope *ClassScope) makeTableParser() c.Combinator[rune, int, node.Table] {
	var (
		parseClass    c.Combinator[rune, int, node.Table]
		parseSubClass c.Combinator[rune, int, node.Table]
	)

	parseItem := c.Try(scope.items.makeParser(
		"expected character class item",
		']',
	))

	parseCharOrRange := scope.makeRangeOrCharParser(
		"expected predefined character, range or characters or character",
		']',
	)

	// can't use common.Choice because parseClass is var
	// and not initiated at this moment to pass as function param
	parseTable := func(
		buf c.Buffer[rune, int],
	) (node.Table, c.Error[int]) {
		pos := buf.Position()

		item, itemErr := parseItem(buf)
		if itemErr == nil {
			return item, nil
		}

		subClass, subClassErr := parseSubClass(buf)
		if subClassErr == nil {
			return subClass, nil
		}

		charOrRange, charErr := parseCharOrRange(buf)
		if charErr == nil {
			return charOrRange, nil
		}

		return nil, c.NewParseError(
			pos,
			"expected character class",
			itemErr,
			subClassErr,
			charErr,
		)
	}

	parseSequenceOfTables := c.Try(c.Some(
		10,
		"expected at least one character class item",
		c.Try(parseTable),
	))

	parseClass = c.Cast(
		parseSequenceOfTables,
		func(tables []node.Table) (node.Table, error) {
			return unicode.MergeTables(tables...), nil
		},
	)

	parseLeftSquare := c.Try(c.Eq[rune, int](
		"expected left square as begining of character sub class",
		'[',
	))

	parseRightSquare := c.Try(c.Eq[rune, int](
		"expected right square as ending of character sub class",
		']',
	))

	parseSubClass = c.Try(func(buf c.Buffer[rune, int]) (node.Table, c.Error[int]) {
		isNegative := true

		_, err := parseLeftSquare(buf)
		if err != nil {
			return nil, err
		}

		_, err = parseNegativePrefix(buf)
		if err != nil {
			isNegative = false
		}

		tables, err := parseSequenceOfTables(buf)
		if err != nil {
			return nil, err
		}

		_, err = parseRightSquare(buf)
		if err != nil {
			return nil, err
		}

		table := unicode.MergeTables(tables...)
		if isNegative {
			return table.Invert(), nil
		}

		return table, nil
	})

	return parseClass
}

func (scope *ClassScope) makeRangeOrCharParser(
	errMessage string,
	except ...rune,
) c.Combinator[rune, int, node.Table] {
	parseMinus := c.Eq[rune, int](
		"expected '-' as separator for range of characters",
		'-',
	)

	parsePredefinedRune := scope.runes.makeParser(
		errMessage,
		except...,
	)

	parseAnyRune := c.Try(c.NoneOf[rune, int](errMessage, except...))

	parseRune := c.Choice(
		errMessage,
		parsePredefinedRune,
		parseAnyRune,
	)

	return func(buf c.Buffer[rune, int]) (node.Table, c.Error[int]) {
		pos := buf.Position()

		from, err := parseRune(buf)
		if err != nil {
			return nil, c.NewParseError(
				pos,
				"expected char or range of chars",
				err,
			)
		}

		pos = buf.Position()

		_, err = parseMinus(buf)
		if err != nil {
			if seekErr := buf.Seek(pos); seekErr != nil {
				return nil, c.NewParseError(
					buf.Position(),
					seekErr.Error(),
				)
			}

			return unicode.NewTable(from), nil
		}

		to, err := parseRune(buf)
		if err != nil {
			if seekErr := buf.Seek(pos); seekErr != nil {
				return nil, c.NewParseError(
					buf.Position(),
					seekErr.Error(),
				)
			}

			return unicode.NewTable(from), nil
		}

		if from > to {
			if seekErr := buf.Seek(pos); seekErr != nil {
				return nil, c.NewParseError(
					buf.Position(),
					seekErr.Error(),
				)
			}

			// TODO : validate tree after?
			validationErr := c.NewParseError(pos, "invalid bounds of range")

			return nil, c.NewParseError(
				pos,
				errMessage,
				validationErr,
			)
		}

		return unicode.NewTableByPredicate(func(x rune) bool {
			return from <= x && x <= to
		}), nil
	}
}
