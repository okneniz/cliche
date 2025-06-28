package parser

import (
	"github.com/okneniz/cliche/encoding/unicode"
	"github.com/okneniz/cliche/node"
	c "github.com/okneniz/parsec/common"
)

type ClassScope struct {
	runes *Scope[rune]
	items *Scope[node.Table]
}

func (scope *ClassScope) Runes() *Scope[rune] {
	return scope.runes
}

func (scope *ClassScope) Items() *Scope[node.Table] {
	return scope.items
}

func (scope *ClassScope) makeParser() c.Combinator[rune, int, node.Node] {
	parseTable := scope.makeTableParser(false)

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		table, err := parseTable(buf)
		if err != nil {
			return nil, err
		}

		return node.NewForTable(table), nil
	}
}

func (scope *ClassScope) makeTableParser(
	isSubclass bool,
) c.Combinator[rune, int, node.Table] {
	var (
		parseClass    c.Combinator[rune, int, node.Table]
		parseSubClass c.Combinator[rune, int, node.Table]
	)

	parseClassItem := scope.items.makeParser(']')
	parseClassChar := scope.makeRangeOrCharParser(']')

	parseTable := func(buf c.Buffer[rune, int]) (node.Table, error) {
		pos := buf.Position()

		classItem, err := parseClassItem(buf)
		if err == nil {
			return classItem, nil
		}

		buf.Seek(pos)

		subClass, err := parseSubClass(buf)
		if err == nil {
			return subClass, nil
		}

		buf.Seek(pos)

		classChar, err := parseClassChar(buf)
		if err == nil {
			return classChar, nil
		}

		buf.Seek(pos)

		return nil, err
	}

	parseSequenceOfTables := c.Some(1, c.Try(parseTable))

	parsePositive := func(buf c.Buffer[rune, int]) (node.Table, error) {
		tables, err := parseSequenceOfTables(buf)
		if err != nil {
			return nil, err
		}

		return unicode.MergeTables(tables...), nil
	}

	parseNegative := c.Skip(
		c.Eq[rune, int]('^'),
		func(buf c.Buffer[rune, int]) (node.Table, error) {
			table, err := parsePositive(buf)
			if err != nil {
				return nil, err
			}

			return table.Invert(), nil
		},
	)

	parseClass = Squares(
		c.Choice(
			c.Try(parseNegative),
			parsePositive,
		),
	)

	if isSubclass {
		parseSubClass = parseClass
	} else {
		parseSubClass = scope.makeTableParser(true)
	}

	return c.Try(parseClass)
}

func (scope *ClassScope) makeRangeOrCharParser(
	except ...rune,
) c.Combinator[rune, int, node.Table] {
	parseSeparator := c.Eq[rune, int]('-')

	parseRune := c.Choice(
		c.Try(scope.runes.makeParser(except...)),
		c.NoneOf[rune, int](except...),
	)

	return func(buf c.Buffer[rune, int]) (node.Table, error) {
		from, err := parseRune(buf)
		if err != nil {
			return nil, err
		}

		pos := buf.Position()

		_, err = parseSeparator(buf)
		if err != nil {
			buf.Seek(pos)
			return unicode.NewTable(from), nil
		}

		to, err := parseRune(buf)
		if err != nil {
			buf.Seek(pos)
			return unicode.NewTable(from), nil
		}

		// TODO : check bounds and return spsecial error

		return unicode.NewTableByPredicate(func(x rune) bool {
			return from <= x && x <= to
		}), nil
	}
}
