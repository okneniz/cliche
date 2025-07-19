package parser

import (
	"fmt"

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

func (scope *ClassScope) makeParser() Parser[node.Node] {
	parseTable := scope.makeTableParser(false)

	return func(buf c.Buffer[rune, int]) (node.Node, Error) {
		table, err := parseTable(buf)
		if err != nil {
			return nil, err
		}

		return node.NewForTable(table), nil
	}
}

func (scope *ClassScope) makeTableParser(isSubclass bool) Parser[node.Table] {
	var (
		parseClass    Parser[node.Table]
		parseSubClass Parser[node.Table]
	)

	parseClassItem := scope.items.makeParser(']')
	parseClassChar := scope.makeRangeOrCharParser(']')

	parseTable := func(
		buf c.Buffer[rune, int],
	) (node.Table, Error) {
		pos := buf.Position()

		classItem, classErr := parseClassItem(buf)
		fmt.Println("class item", classItem, classErr)
		if classErr == nil {
			return classItem, nil
		}

		buf.Seek(pos)

		subClass, subClassErr := parseSubClass(buf)
		fmt.Println("sub class", subClass, subClassErr)
		if subClassErr == nil {
			return subClass, nil
		}

		buf.Seek(pos)

		classChar, charErr := parseClassChar(buf)
		fmt.Println("class char", classChar, charErr)
		if charErr == nil {
			return classChar, nil
		}

		buf.Seek(pos)

		return nil, MergeErrors(
			classErr,
			subClassErr,
			charErr,
		)
	}

	parseSequenceOfTables := func(
		buf c.Buffer[rune, int],
	) ([]node.Table, Error) {
		result := make([]node.Table, 0)
		start := buf.Position()

		for !buf.IsEOF() {
			pos := buf.Position()

			table, err := parseTable(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			result = append(result, table)
		}

		if len(result) == 0 {
			// TODO : what about empty class? []
			return nil, Expected("sequence of class items", start, c.NotEnoughElements)
		}

		return result, nil
	}

	negativePrefix := Eq('^')
	leftSquare := Eq('[')
	rightSquare := Eq(']')

	parseClass = func(buf c.Buffer[rune, int]) (node.Table, Error) {
		pos := buf.Position()
		isNegative := true

		_, err := leftSquare(buf)
		if err != nil {
			buf.Seek(pos)
			return nil, err
		}

		beforePrefixPos := buf.Position()
		_, prefixErr := negativePrefix(buf)
		if prefixErr != nil {
			isNegative = false
			buf.Seek(beforePrefixPos)
		}

		tables, seqErr := parseSequenceOfTables(buf)
		if seqErr != nil {
			buf.Seek(pos)
			return nil, seqErr
		}

		_, squareErr := rightSquare(buf)
		if err != nil {
			buf.Seek(pos)
			return nil, squareErr
		}

		table := unicode.MergeTables(tables...)
		if isNegative {
			return table.Invert(), nil
		}

		return table, nil
	}

	if isSubclass {
		parseSubClass = parseClass
	} else {
		parseSubClass = scope.makeTableParser(true)
	}

	return parseClass
}

func (scope *ClassScope) makeRangeOrCharParser(except ...rune) Parser[node.Table] {
	parseSeparator := Eq('-')
	parsePredefinedRune := scope.runes.makeParser(except...)
	parseAnyRune := NoneOf(except...)

	parseRune := func(buf c.Buffer[rune, int]) (rune, Error) {
		pos := buf.Position()

		x, runeErr := parsePredefinedRune(buf)
		fmt.Println("predefined class char", x, runeErr)
		if runeErr == nil {
			return x, nil
		}

		buf.Seek(pos)

		// return parseAnyRune(buf)
		x, aErr := parseAnyRune(buf)
		fmt.Println("WTF", x, aErr)
		if aErr != nil {
			return -1, aErr
		}

		return x, nil
	}

	return func(buf c.Buffer[rune, int]) (node.Table, Error) {
		pos := buf.Position()

		from, err := parseRune(buf)
		fmt.Println("first in range", from, err)
		if err != nil {
			return nil, Expected("char or range of chars", pos, err)
		}

		// TODO : разобраться с nil values и тд
		// https://go.dev/doc/faq#nil_error

		pos = buf.Position()

		sep, sepErr := parseSeparator(buf)
		fmt.Printf("separator %#v %#v %#v\n", sep, sepErr, err)
		if sepErr != nil {
			fmt.Printf("but??? %T %#v\n", sepErr, sepErr)
			fmt.Println("but why?", sepErr, sepErr.Error(), sepErr != nil, sepErr == nil)
			buf.Seek(pos)
			return unicode.NewTable(from), nil
		}

		fmt.Println("WOW???")

		to, err := parseRune(buf)
		fmt.Println("second in range", to, err)
		if err != nil {
			buf.Seek(pos)
			return unicode.NewTable(from), nil
		}

		fmt.Println("validation")

		if from > to {
			// TODO : how to return errors right here?
			// validate tree after?
			return nil, Expected("range of chars", pos, fmt.Errorf("invalid bounds"))
		}

		return unicode.NewTableByPredicate(func(x rune) bool {
			return from <= x && x <= to
		}), nil
	}
}
