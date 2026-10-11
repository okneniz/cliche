package parser

import (
	"unicode"

	unicodePkg "github.com/okneniz/cliche/encoding/unicode"
	"github.com/okneniz/cliche/node"
	"github.com/okneniz/parsec"
)

type ClassScope struct {
	runes *Scope[rune]
	items *Scope[node.Table]
}

var (
	parseNegativePrefix = parsec.Try(parsec.Eq[rune, int, State](
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

func (scope *ClassScope) makeParser() parsec.Combinator[rune, int, node.Node, State] {
	parseTable := scope.makeTableParser()

	parseLeftSquare := parsec.Eq[rune, int, State](
		"expected left square as beginning of character class",
		'[',
	)

	parseRightSquare := parsec.Eq[rune, int, State](
		"expected right square as ending of character class",
		']',
	)

	return parsec.Try(func(state State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
		isNegative := true

		_, err := parseLeftSquare(state, buf)
		if err != nil {
			return nil, err
		}

		_, err = parseNegativePrefix(state, buf)
		if err != nil {
			isNegative = false
		}

		table, err := parseTable(state, buf)
		if err != nil {
			return nil, err
		}

		_, err = parseRightSquare(state, buf)
		if err != nil {
			return nil, err
		}

		if isNegative {
			return node.NewNegativeClass(table), nil
		}

		return node.NewClass(table), nil
	})
}

func (scope *ClassScope) makeTableParser() parsec.Combinator[rune, int, node.Table, State] {
	var (
		parseClass    parsec.Combinator[rune, int, node.Table, State]
		parseSubClass parsec.Combinator[rune, int, node.Table, State]
	)

	parseItem := parsec.Try(scope.items.makeParser(
		"expected character class item",
		']',
	))

	parseCharOrRange := scope.makeRangeOrCharParser(
		"expected predefined character, range or characters or character",
		']',
	)

	// can't use parsec.Choice because parseClass is var
	// and not initiated at this moment to pass as function param
	parseTable := func(
		state State,
		buf parsec.Buffer[rune, int],
	) (node.Table, parsec.Error[int]) {
		pos := buf.Position()

		item, itemErr := parseItem(state, buf)
		if itemErr == nil {
			return item, nil
		}

		subClass, subClassErr := parseSubClass(state, buf)
		if subClassErr == nil {
			return subClass, nil
		}

		charOrRange, charErr := parseCharOrRange(state, buf)
		if charErr == nil {
			return charOrRange, nil
		}

		return nil, parsec.NewParseError(
			pos,
			"expected character class",
			itemErr,
			subClassErr,
			charErr,
		)
	}

	parseSequenceOfTables := parsec.Try(parsec.Many(0, parsec.Try(parseTable)))

	parseClass = parsec.Cast(
		parseSequenceOfTables,
		func(tables []node.Table) (node.Table, error) {
			return unicodePkg.MergeTables(tables...), nil
		},
	)

	parseLeftSquare := parsec.Try(parsec.Eq[rune, int, State](
		"expected left square as beginning of character sub class",
		'[',
	))

	parseRightSquare := parsec.Try(parsec.Eq[rune, int, State](
		"expected right square as ending of character sub class",
		']',
	))

	parseSubClass = parsec.Try(func(state State, buf parsec.Buffer[rune, int]) (node.Table, parsec.Error[int]) {
		isNegative := true

		_, err := parseLeftSquare(state, buf)
		if err != nil {
			return nil, err
		}

		_, err = parseNegativePrefix(state, buf)
		if err != nil {
			isNegative = false
		}

		tables, err := parseSequenceOfTables(state, buf)
		if err != nil {
			return nil, err
		}

		_, err = parseRightSquare(state, buf)
		if err != nil {
			return nil, err
		}

		table := unicodePkg.MergeTables(tables...)
		if isNegative {
			// TODO : надо как-то опционально выставлять max для разных движков
			return table.Invert(unicode.MaxRune), nil
		}

		return table, nil
	})

	return parseClass
}

func (scope *ClassScope) makeRangeOrCharParser(
	errMessage string,
	except ...rune,
) parsec.Combinator[rune, int, node.Table, State] {
	parseMinus := parsec.Eq[rune, int, State](
		"expected '-' as separator for range of characters",
		'-',
	)

	parsePredefinedRune := scope.runes.makeParser(
		errMessage,
		except...,
	)

	parseAnyRune := parsec.Try(parsec.NoneOf[rune, int, State](errMessage, except...))

	parseRune := parsec.Choice(
		errMessage,
		parsePredefinedRune,
		parseAnyRune,
	)

	return func(state State, buf parsec.Buffer[rune, int]) (node.Table, parsec.Error[int]) {
		pos := buf.Position()

		from, err := parseRune(state, buf)
		if err != nil {
			return nil, parsec.NewParseError(
				pos,
				"expected char or range of chars",
				err,
			)
		}

		pos = buf.Position()

		_, err = parseMinus(state, buf)
		if err != nil {
			if seekErr := buf.Seek(pos); seekErr != nil {
				return nil, parsec.NewParseError(
					buf.Position(),
					seekErr.Error(),
				)
			}

			return unicodePkg.NewTable(from), nil
		}

		to, err := parseRune(state, buf)
		if err != nil {
			if seekErr := buf.Seek(pos); seekErr != nil {
				return nil, parsec.NewParseError(
					buf.Position(),
					seekErr.Error(),
				)
			}

			return unicodePkg.NewTable(from), nil
		}

		// if from > to {
		// 	if seekErr := buf.Seek(pos); seekErr != nil {
		// 		return nil, parsec.NewParseError(
		// 			buf.Position(),
		// 			seekErr.Error(),
		// 		)
		// 	}

		// 	// TODO : validate tree after?
		// 	validationErr := parsec.NewParseError(pos, "invalid bounds of range")

		// 	return nil, parsec.NewParseError(
		// 		pos,
		// 		errMessage,
		// 		validationErr,
		// 	)
		// }

		return unicodePkg.NewTableByPredicate(unicode.MaxRune, func(x rune) bool {
			return from <= x && x <= to
		}), nil
	}
}
