package cliche

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	c "github.com/okneniz/parsec/common"
	"golang.org/x/text/unicode/rangetable"
)

var (
	propertyTable    = parsePropertyName()
	parseHexChar     = parseHexNumber(2, 2) // TODO : check size in different docs
	parseUnicodeChar = parseHexNumber(1, 4) // TODO : check size in different docs
	parseOctalChar   = braces(parseOctal(3))

	DefaultOptions = []Option[*CustomParser]{
		WithBracket("alnum", func(x rune) bool {
			return unicode.IsLetter(x) || unicode.IsMark(x) || unicode.IsDigit(x)
		}),
		WithBracket("alpha", func(x rune) bool {
			return unicode.IsLetter(x) || unicode.IsMark(x)
		}),
		WithBracket("ascii", func(x rune) bool {
			return x < unicode.MaxASCII
		}),
		WithBracket("blank", func(x rune) bool {
			return x == ' ' || x == '\t'
		}),
		WithBracket("digit", func(x rune) bool {
			return unicode.IsDigit(x)
		}),
		WithBracket("lower", func(x rune) bool {
			return unicode.IsLower(x)
		}),
		WithBracket("upper", func(x rune) bool {
			return unicode.IsUpper(x)
		}),
		WithBracket("space", func(x rune) bool {
			return unicode.IsSpace(x)
		}),
		WithBracket("cntrl", func(x rune) bool {
			return unicode.IsControl(x)
		}),
		WithBracket("print", func(x rune) bool {
			return unicode.IsPrint(x)
		}),
		WithBracket("graph", func(x rune) bool {
			return unicode.IsGraphic(x) && !unicode.IsSpace(x)
		}),
		WithBracket("punct", func(x rune) bool {
			return unicode.IsPunct(x)
		}),
		WithBracket("xdigit", func(x rune) bool {
			return isHex(x)
		}),
		WithBracket("word", func(x rune) bool {
			return isWord(x)
		}),
		WithEscapedMetaChar("d", func(x rune) bool {
			return unicode.IsDigit(x)
		}),
		WithEscapedMetaChar("D", func(x rune) bool {
			return !unicode.IsDigit(x)
		}),
		WithEscapedMetaChar("w", func(x rune) bool {
			return isWord(x)
		}),
		WithEscapedMetaChar("W", func(x rune) bool {
			return !isWord(x)
		}),
		WithEscapedMetaChar("s", func(x rune) bool {
			return unicode.IsSpace(x)
		}),
		WithEscapedMetaChar("S", func(x rune) bool {
			return !unicode.IsSpace(x)
		}),
		WithEscapedMetaChar("h", func(x rune) bool {
			return isHex(x)
		}),
		WithEscapedMetaChar("H", func(x rune) bool {
			return !isHex(x)
		}),
		WithPrefix("\\p", func(buf c.Buffer[rune, int]) (Node, error) {
			table, err := propertyTable(buf)
			if err != nil {
				return nil, err
			}

			return nodeForTable(table), nil
		}),
		WithInClassPrefix("\\p", func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
			table, err := propertyTable(buf)
			if err != nil {
				return nil, err
			}

			return table, nil
		}),
		WithPrefix("\\P", func(buf c.Buffer[rune, int]) (Node, error) {
			table, err := propertyTable(buf)
			if err != nil {
				return nil, err
			}

			return nodeForTable(negatiateTable(table)), nil
		}),
		WithInClassPrefix("\\P", func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
			table, err := propertyTable(buf)
			if err != nil {
				return nil, err
			}

			return negatiateTable(table), nil
		}),
		WithPrefix("\\x", func(buf c.Buffer[rune, int]) (Node, error) {
			num, err := parseHexChar(buf)
			if err != nil {
				return nil, err
			}

			// TODO : check bounds
			r := rune(num)

			return nodeForTable(rangetable.New(r)), nil
		}),
		WithInClassPrefix("\\x", func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
			num, err := parseHexChar(buf)
			if err != nil {
				return nil, err
			}

			// TODO : check bounds
			r := rune(num)

			return rangetable.New(r), nil
		}),
		WithPrefix("\\o", func(buf c.Buffer[rune, int]) (Node, error) {
			num, err := parseOctalChar(buf)
			if err != nil {
				return nil, err
			}

			// TODO : check bounds
			r := rune(num)

			return nodeForTable(rangetable.New(r)), nil
		}),
		WithInClassPrefix("\\o", func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
			num, err := parseOctalChar(buf)
			if err != nil {
				return nil, err
			}

			// TODO : check bounds
			r := rune(num)

			return rangetable.New(r), nil
		}),
		WithPrefix("\\u", func(buf c.Buffer[rune, int]) (Node, error) {
			num, err := parseUnicodeChar(buf)
			if err != nil {
				return nil, err
			}

			// TODO : check bounds
			r := rune(num)

			return nodeForTable(rangetable.New(r)), nil
		}),
		WithInClassPrefix("\\u", func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
			num, err := parseUnicodeChar(buf)
			if err != nil {
				return nil, err
			}

			// TODO : check bounds
			r := rune(num)

			return rangetable.New(r), nil
		}),
		WithParser(parseBackReference),
	}
)

func WithBracket(
	name string, predicate func(rune) bool,
) Option[*CustomParser] {
	// TODO : validate name to avoid conflicts with default spec symbols ".?+*^$[]{}()"

	table := predicateToTable(predicate)
	negatiatedTable := negatiateTable(table)

	parseNode := func(buf c.Buffer[rune, int]) (Node, error) {
		return nodeForTable(table), nil
	}

	parseNegatedNode := func(buf c.Buffer[rune, int]) (Node, error) {
		return nodeForTable(negatiatedTable), nil
	}

	parseTable := func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
		return table, nil
	}

	parseNegatedTable := func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
		return negatiatedTable, nil
	}

	return func(parser *CustomParser) {
		parser.prefixes["[[:"+name+":]]"] = parseNode
		parser.prefixes["[[:^"+name+":]]"] = parseNegatedNode

		parser.inClassPrefixes["[[:"+name+":]]"] = parseTable
		parser.inClassPrefixes["[[:^"+name+":]]"] = parseNegatedTable
	}
}

func WithEscapedMetaChar(
	name string, predicate func(rune) bool,
) Option[*CustomParser] {
	// TODO : validate char

	table := predicateToTable(predicate)
	parse := func(buf c.Buffer[rune, int]) (Node, error) {
		return nodeForTable(table), nil
	}
	parseTable := func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
		return table, nil
	}

	return func(parser *CustomParser) {
		parser.prefixes["\\"+name] = parse
		parser.inClassPrefixes["\\"+name] = parseTable
	}
}

func WithPrefix(
	name string, parse c.Combinator[rune, int, Node],
) Option[*CustomParser] {
	return func(parser *CustomParser) {
		// TODO : validate name
		parser.prefixes[name] = parse
	}
}

func WithInClassPrefix(
	name string, parse c.Combinator[rune, int, *unicode.RangeTable],
) Option[*CustomParser] {
	return func(parser *CustomParser) {
		// TODO : validate name
		parser.inClassPrefixes["\\"+name] = parse
	}
}

func WithParser(
	p func(except ...rune) c.Combinator[rune, int, Node],
) Option[*CustomParser] {
	return func(parser *CustomParser) {
		parser.parsers = append(parser.parsers, p)
	}
}

func WithInClassParser(
	p func(except ...rune) c.Combinator[rune, int, *unicode.RangeTable],
) Option[*CustomParser] {
	return func(parser *CustomParser) {
		parser.inClassParsers = append(parser.inClassParsers, p)
	}
}

func parseBackReference(except ...rune) c.Combinator[rune, int, Node] {
	digits := []rune("0123456789")

	if len(except) > 0 {
		exceptM := make(map[rune]struct{}, len(except))
		for _, c := range except {
			exceptM[c] = struct{}{}
		}

		for _, c := range digits {
			if _, exists := exceptM[c]; exists {
				panic("exceptions include digit " + string(c))
			}
		}
	}

	// is it possible to have back reference more than nine?
	// for example \13 or \99 ?
	parse := c.Skip[rune, int](
		c.Eq[rune, int]('\\'),
		Quantifier(1, 2, c.OneOf[rune, int](digits...)),
	)

	return func(buf c.Buffer[rune, int]) (Node, error) {
		fmt.Println("wow", buf)

		runes, err := parse(buf)
		if err != nil {
			fmt.Println("WTF", err)
			return nil, err
		}

		str := strings.ToLower(string(runes))

		index, err := strconv.ParseInt(str, 16, 64)
		if err != nil {
			return nil, err
		}

		return nodeForReference(int(index)), nil
	}
}

func parseHexNumber(
	from, to int,
) c.Combinator[rune, int, int] {
	parse := Quantifier(
		from,
		to,
		c.OneOf[rune, int]([]rune("0123456789abcdefABCDEF")...),
	)

	return func(buf c.Buffer[rune, int]) (int, error) {
		runes, err := parse(buf)
		if err != nil {
			return -1, err
		}

		str := strings.ToLower(string(runes))

		num, err := strconv.ParseInt(str, 16, 64)
		if err != nil {
			return -1, err
		}

		return int(num), nil
	}
}

func parseOctal(size int) c.Combinator[rune, int, int] {
	allowed := "01234567"
	parse := c.Count(size, c.OneOf[rune, int]([]rune(allowed)...))

	return func(buf c.Buffer[rune, int]) (int, error) {
		runes, err := parse(buf)
		if err != nil {
			return -1, err
		}

		str := strings.ToLower(string(runes))

		num, err := strconv.ParseInt(str, 8, 64)
		if err != nil {
			return -1, err
		}

		return int(num), nil
	}
}

func parsePropertyName() c.Combinator[rune, int, *unicode.RangeTable] {
	allProperties := make(map[string]*unicode.RangeTable)

	for k, v := range unicode.Categories {
		x := v
		allProperties[k] = x
	}

	for k, v := range unicode.Properties {
		x := v
		allProperties[k] = x
	}

	for k, v := range unicode.Scripts {
		x := v
		allProperties[k] = x
	}

	cases := make([]c.Combinator[rune, int, *unicode.RangeTable], 0, len(allProperties)*3)

	for name, t := range allProperties {
		parse := c.SequenceOf[rune, int]([]rune("{" + name + "}")...)
		table := t

		cases = append(cases, func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
			_, err := parse(buf)
			if err != nil {
				return nil, err
			}

			return table, nil
		})
	}

	return tryAll(cases...)
}
