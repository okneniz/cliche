package cliche

import (
	"slices"
	"strconv"
	"strings"
	"unicode"

	unicodeEncoding "github.com/okneniz/cliche/encoding/unicode"
	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/parser"
	c "github.com/okneniz/parsec/common"
)

var _ node.Table = new(unicodeEncoding.UnicodeTable)

// split to another groups of options
//
// common:
//   - ranges chars (chars is not same as escaped chars (char vs table))
//   - chars can be range in classes
// not in class:
//   - ^$
//   - \A, \z
// in class:
//   - bracket
//   - quantifiers

var (
	Brackets = []parser.Option[*parser.CustomParser]{
		parser.WithBracket("alnum", func(x rune) bool {
			return unicode.IsLetter(x) || unicode.IsMark(x) || unicode.IsDigit(x)
		}),
		parser.WithBracket("alpha", func(x rune) bool {
			return unicode.IsLetter(x) || unicode.IsMark(x)
		}),
		parser.WithBracket("ascii", func(x rune) bool {
			return x < unicode.MaxASCII
		}),
		parser.WithBracket("blank", func(x rune) bool {
			return x == ' ' || x == '\t'
		}),
		parser.WithBracket("digit", func(x rune) bool {
			return unicode.IsDigit(x)
		}),
		parser.WithBracket("lower", func(x rune) bool {
			return unicode.IsLower(x)
		}),
		parser.WithBracket("upper", func(x rune) bool {
			return unicode.IsUpper(x)
		}),
		parser.WithBracket("space", func(x rune) bool {
			return unicode.IsSpace(x)
		}),
		parser.WithBracket("cntrl", func(x rune) bool {
			return unicode.IsControl(x)
		}),
		parser.WithBracket("print", func(x rune) bool {
			return unicode.IsPrint(x)
		}),
		parser.WithBracket("graph", func(x rune) bool {
			return unicode.IsGraphic(x) && !unicode.IsSpace(x)
		}),
		parser.WithBracket("punct", func(x rune) bool {
			return unicode.IsPunct(x)
		}),
		parser.WithBracket("xdigit", func(x rune) bool {
			return isHex(x)
		}),
		parser.WithBracket("word", func(x rune) bool {
			return isWord(x)
		}),
	}

	EscapedMetacharacters = []parser.Option[*parser.CustomParser]{
		parser.WithEscapedMetaChar("d", func(x rune) bool {
			return unicode.IsDigit(x)
		}),
		parser.WithEscapedMetaChar("D", func(x rune) bool {
			return !unicode.IsDigit(x)
		}),
		parser.WithEscapedMetaChar("w", func(x rune) bool {
			return isWord(x)
		}),
		parser.WithEscapedMetaChar("W", func(x rune) bool {
			return !isWord(x)
		}),
		parser.WithEscapedMetaChar("s", func(x rune) bool {
			return unicode.IsSpace(x)
		}),
		parser.WithEscapedMetaChar("S", func(x rune) bool {
			return !unicode.IsSpace(x)
		}),
		parser.WithEscapedMetaChar("h", func(x rune) bool {
			return isHex(x)
		}),
		parser.WithEscapedMetaChar("H", func(x rune) bool {
			return !isHex(x)
		}),
		parser.WithPrefix(`\A`, func(buf c.Buffer[rune, int]) (node.Node, error) {
			return node.NewStartOfString(), nil
		}),
		parser.WithPrefix(`\z`, func(buf c.Buffer[rune, int]) (node.Node, error) {
			return node.NewEndOfString(), nil
		}),
		parser.WithPrefix(`\K`, func(buf c.Buffer[rune, int]) (node.Node, error) {
			return node.NewKeep(), nil
		}),
	}

	OnigmoOptions = slices.Concat(
		Brackets,
		EscapedMetacharacters,
		[]parser.Option[*parser.CustomParser]{
			parser.WithPrefix(".", func(buf c.Buffer[rune, int]) (node.Node, error) {
				return node.NewDot(), nil
			}),
			parser.WithPrefix("^", func(buf c.Buffer[rune, int]) (node.Node, error) {
				return node.NewStartOfLine(), nil
			}),
			parser.WithPrefix("$", func(buf c.Buffer[rune, int]) (node.Node, error) {
				return node.NewEndOfLine(), nil
			}),
			parser.WithPrefix(`\p`, func(buf c.Buffer[rune, int]) (node.Node, error) {
				table, err := propertyTable(buf)
				if err != nil {
					return nil, err
				}

				return node.NewForTable(table), nil
			}),
			parser.WithInClassPrefix(`\p`, func(buf c.Buffer[rune, int]) (node.Table, error) {
				table, err := propertyTable(buf)
				if err != nil {
					return nil, err
				}

				return table, nil
			}),
			parser.WithPrefix(`\P`, func(buf c.Buffer[rune, int]) (node.Node, error) {
				table, err := propertyTable(buf)
				if err != nil {
					return nil, err
				}

				return node.NewForTable(table.Invert()), nil
			}),
			parser.WithInClassPrefix(`\P`, func(buf c.Buffer[rune, int]) (node.Table, error) {
				table, err := propertyTable(buf)
				if err != nil {
					return nil, err
				}

				return table.Invert(), nil
			}),
			parser.WithPrefix(`\x`, func(buf c.Buffer[rune, int]) (node.Node, error) {
				num, err := parseHexChar(buf)
				if err != nil {
					return nil, err
				}

				// TODO : check bounds
				r := rune(num)

				table := unicodeEncoding.NewTableFor(r)
				return node.NewForTable(table), nil
			}),
			parser.WithInClassPrefix(`\x`, func(buf c.Buffer[rune, int]) (node.Table, error) {
				num, err := parseHexChar(buf)
				if err != nil {
					return nil, err
				}

				// TODO : check bounds
				r := rune(num)

				return unicodeEncoding.NewTableFor(r), nil
			}),
			parser.WithPrefix(`\o`, func(buf c.Buffer[rune, int]) (node.Node, error) {
				num, err := parseOctalChar(buf)
				if err != nil {
					return nil, err
				}

				// TODO : check bounds
				r := rune(num)

				table := unicodeEncoding.NewTableFor(r)
				return node.NewForTable(table), nil
			}),
			parser.WithInClassPrefix(`\o`, func(buf c.Buffer[rune, int]) (node.Table, error) {
				num, err := parseOctalChar(buf)
				if err != nil {
					return nil, err
				}

				// TODO : check bounds
				r := rune(num)

				return unicodeEncoding.NewTableFor(r), nil
			}),
			parser.WithPrefix(`\u`, func(buf c.Buffer[rune, int]) (node.Node, error) {
				num, err := parseUnicodeChar(buf)
				if err != nil {
					return nil, err
				}

				// TODO : check bounds
				r := rune(num)

				table := unicodeEncoding.NewTableFor(r)
				return node.NewForTable(table), nil
			}),
			parser.WithInClassPrefix(`\u`, func(buf c.Buffer[rune, int]) (node.Table, error) {
				num, err := parseUnicodeChar(buf)
				if err != nil {
					return nil, err
				}

				// TODO : check bounds
				r := rune(num)

				return unicodeEncoding.NewTableFor(r), nil
			}),
			parser.WithPrefix(`\k`, func(buf c.Buffer[rune, int]) (node.Node, error) {
				parse := parser.Angles( // TODO : prebuild it by closure
					c.Some(
						0,
						c.Try(c.NoneOf[rune, int]('>')),
					),
				)

				name, err := parse(buf)
				if err != nil {
					return nil, err
				}

				return node.NewForNameReference(string(name)), nil
			}),
			parser.WithParser(parseBackReference),
		},
	)

	propertyTable    = parsePropertyName()
	parseHexChar     = parseHexNumber(2, 2) // TODO : check size in different docs
	parseUnicodeChar = parseHexNumber(1, 4) // TODO : check size in different docs
	parseOctalChar   = parser.Braces(parseOctal(3))

	OnigmoParser = parser.NewParser(OnigmoOptions...)
)

func parseBackReference(
	except ...rune,
) c.Combinator[rune, int, node.Node] {
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
		parser.Quantifier(1, 2, c.OneOf[rune, int](digits...)),
	)

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		runes, err := parse(buf)
		if err != nil {
			return nil, err
		}

		str := strings.ToLower(string(runes))

		index, err := strconv.ParseInt(str, 16, 64)
		if err != nil {
			return nil, err
		}

		return node.NodeForReference(int(index)), nil
	}
}

func parseHexNumber(
	from, to int,
) c.Combinator[rune, int, int] {
	parse := parser.Quantifier(
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

func parseOctal(
	size int,
) c.Combinator[rune, int, int] {
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

func parsePropertyName() c.Combinator[rune, int, node.Table] {
	allProperties := make(map[string]node.Table)

	for k, v := range unicode.Categories {
		x := v
		allProperties[k] = unicodeEncoding.NewTable(x)
	}

	for k, v := range unicode.Properties {
		x := v
		allProperties[k] = unicodeEncoding.NewTable(x)
	}

	for k, v := range unicode.Scripts {
		x := v
		allProperties[k] = unicodeEncoding.NewTable(x)
	}

	cases := make([]c.Combinator[rune, int, node.Table], 0, len(allProperties))

	for name, t := range allProperties {
		parse := c.SequenceOf[rune, int]([]rune("{" + name + "}")...)
		table := t

		cases = append(cases, func(buf c.Buffer[rune, int]) (node.Table, error) {
			_, err := parse(buf)
			if err != nil {
				return nil, err
			}

			return table, nil
		})
	}

	return parser.TryAll(cases...)
}

func isWord(x rune) bool {
	return x == '_' || unicode.IsLetter(x) || unicode.IsDigit(x)
}

func isHex(x rune) bool {
	return x >= '0' && x <= '9' ||
		x >= 'a' && x <= 'f' ||
		x >= 'A' && x <= 'F'
}
