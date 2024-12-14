package cliche

import (
	"errors"
	"fmt"
	"unicode"

	"golang.org/x/text/unicode/rangetable"

	c "github.com/okneniz/parsec/common"
)

type parser = c.Combinator[rune, int, Node]
type tableParser = c.Combinator[rune, int, *unicode.RangeTable]

var (
	defaultParser          = parseRegexp()
	InvalidQuantifierError = errors.New("target of repeat operator is not specified")
)

func SkipString(data string) c.Combinator[rune, int, struct{}] {
	none := struct{}{}

	return func(buf c.Buffer[rune, int]) (struct{}, error) {
		l := len(data)
		for _, x := range data {
			r, err := buf.Read(true)
			if err != nil {
				return none, err
			}
			if x != r {
				return none, c.NothingMatched
			}
			l -= 1
		}

		if l != 0 {
			return none, c.NothingMatched
		}

		return none, nil
	}
}

// TODO : return error for invalid escaped chars like '\x' (check on rubular)

func parseRegexp() parser {
	var parseExpression parser
	var parseNestedExpression parser

	sep := c.Eq[rune, int]('|')

	// parse alternation
	alternation := func(buf c.Buffer[rune, int]) (*alternation, error) {
		variant, err := parseNestedExpression(buf)
		if err != nil {
			return nil, err
		}

		variants := make([]Node, 0, 1)
		variants = append(variants, variant)

		for !buf.IsEOF() {
			pos := buf.Position()

			_, err = sep(buf)
			if err != nil {
				buf.Seek(pos)
				break // return error instead break?
			}

			variant, err = parseNestedExpression(buf)
			if err != nil {
				buf.Seek(pos)
				break // return error instead break?
			}

			variants = append(variants, variant)
		}

		// TODO : check length and eof

		return newAlternation(variants), nil
	}

	// parse node
	parseNode := parseOptionalQuantifier(
		choice(
			parseBrackets(),
			parseCharacterClasses('|'),
			parseNotCapturedGroup(alternation),
			parseNamedGroup(alternation),
			parseGroup(alternation),
			parseInvalidQuantifier(),
			parseEscapedMetaCharacters(),
			parseMetaCharacters(),
			parseEscapedSpecSymbols(),
			parseCharacter('|'),
		),
	)

	// parse node of nested expression
	parseNestedNode := parseOptionalQuantifier(
		choice(
			parseCharacterClasses('|', ')'),
			parseNotCapturedGroup(alternation),
			parseNamedGroup(alternation),
			parseGroup(alternation),
			parseInvalidQuantifier(),
			parseEscapedMetaCharacters(),
			parseMetaCharacters(),
			parseEscapedSpecSymbols(),
			parseCharacter('|', ')'),
		),
	)

	parseExpression = func(buf c.Buffer[rune, int]) (Node, error) {
		first, err := parseNode(buf)
		if err != nil {
			return nil, err
		}

		last := first

		for !buf.IsEOF() {
			pos := buf.Position()

			next, err := parseNode(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			last.GetNestedNodes()[next.GetKey()] = next
			last = next
		}

		return first, nil
	}

	parseNestedExpression = func(buf c.Buffer[rune, int]) (Node, error) {
		first, err := parseNestedNode(buf)
		if err != nil {
			return nil, err
		}

		last := first

		for !buf.IsEOF() {
			pos := buf.Position()

			next, err := parseNestedNode(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			last.GetNestedNodes()[next.GetKey()] = next
			last = next
		}

		return first, nil
	}

	// parse alternation or expression
	return func(buf c.Buffer[rune, int]) (Node, error) {
		expression, err := parseExpression(buf)
		if err != nil {
			return nil, err
		}
		if buf.IsEOF() {
			return expression, nil
		}

		variants := make([]Node, 0, 1)
		variants = append(variants, expression)

		for !buf.IsEOF() {
			_, err = sep(buf)
			if err != nil {
				return nil, err
			}

			expression, err = parseExpression(buf)
			if err != nil {
				return nil, err
			}

			variants = append(variants, expression)
		}

		return newAlternation(variants), nil
	}
}

func parseCharacterClasses(except ...rune) parser {
	parseTable := c.Choice[rune, int, *unicode.RangeTable](
		c.Try(parseRangeTable(append(except, ']')...)),
		c.Try(parseEscapedMetaCharactersTable()),
		c.Try(parseEscapedSpecSymbolsTable()),
		c.Try(parseCharacterTable(append(except, ']')...)),
	)

	return choice(
		parseNegatedCharacterClass(parseTable),
		parseCharacterClass(parseTable),
	)
}

func choice(parsers ...parser) parser {
	attempts := make([]parser, len(parsers))

	for i, parse := range parsers {
		attempts[i] = c.Try(parse)
	}

	return c.Choice(attempts...)
}

func between[T any, S any](
	before c.Combinator[rune, int, S],
	body c.Combinator[rune, int, T],
	after c.Combinator[rune, int, S],
) c.Combinator[rune, int, T] {
	return c.Between(before, body, after)
}

func parens[T any](
	body c.Combinator[rune, int, T],
) c.Combinator[rune, int, T] {
	return between(
		c.Eq[rune, int]('('),
		body,
		c.Eq[rune, int](')'),
	)
}

func angles[T any](
	body c.Combinator[rune, int, T],
) c.Combinator[rune, int, T] {
	return between(
		c.Eq[rune, int]('<'),
		body,
		c.Eq[rune, int]('>'),
	)
}

func squares[T any](
	body c.Combinator[rune, int, T],
) c.Combinator[rune, int, T] {
	return between(
		c.Eq[rune, int]('['),
		body,
		c.Eq[rune, int](']'),
	)
}

func number() c.Combinator[rune, int, int] {
	digit := c.Try[rune, int](c.Range[rune, int]('0', '9'))
	zero := rune('0')

	return func(buf c.Buffer[rune, int]) (int, error) {
		token, err := digit(buf)
		if err != nil {
			return 0, err
		}

		result := int(token - zero)
		for {
			token, err = digit(buf)
			if err != nil {
				break
			}

			result = result * 10
			result += int(token - zero)
		}

		return result, nil
	}
}

func parseEscapedSpecSymbols() parser {
	symbols := ".?+*^$[]{}()"
	cases := make(map[rune]parser)

	for _, v := range symbols {
		r := v

		cases[r] = func(buf c.Buffer[rune, int]) (Node, error) {
			x := char{
				Value:      r,
				nestedNode: newNestedNode(),
			}

			return &x, nil
		}
	}

	return c.Skip(
		c.Eq[rune, int]('\\'),
		c.MapAs(
			cases,
			c.Any[rune, int](),
		),
	)
}

func parseInvalidQuantifier() parser {
	invalidChars := map[rune]struct{}{
		'?': {},
		'*': {},
		'+': {},
	}

	return func(buf c.Buffer[rune, int]) (Node, error) {
		x, err := buf.Read(false)
		if err != nil {
			return nil, err
		}

		if _, exists := invalidChars[x]; exists {
			return nil, InvalidQuantifierError
		}

		return nil, c.NothingMatched
	}
}

func parseOptionalQuantifier(expression parser) parser {
	any := c.Any[rune, int]()
	digit := c.Try(number())
	comma := c.Try(c.Eq[rune, int](','))
	rightBrace := c.Eq[rune, int]('}')

	parse := c.Choice(
		c.Try(func(buf c.Buffer[rune, int]) (*quantifier, error) { // {1,1}
			from, err := digit(buf)
			if err != nil {
				return nil, err
			}

			_, err = comma(buf)
			if err != nil {
				return nil, err
			}

			to, err := digit(buf)
			if err != nil {
				return nil, err
			}

			if from > to {
				return nil, c.NothingMatched
			}

			_, err = rightBrace(buf)
			if err != nil {
				return nil, err
			}

			if from == to {
				return &quantifier{
					From: from,
					To:   nil,
					More: false,
				}, nil
			}

			return &quantifier{
				From: from,
				To:   &to,
				More: false,
			}, nil
		}),
		c.Try(func(buf c.Buffer[rune, int]) (*quantifier, error) { // {,1}
			_, err := comma(buf)
			if err != nil {
				return nil, err
			}

			to, err := digit(buf)
			if err != nil {
				return nil, err
			}

			_, err = rightBrace(buf)
			if err != nil {
				return nil, err
			}

			return &quantifier{
				From: 0,
				To:   &to,
				More: false,
			}, nil
		}),
		c.Try(func(buf c.Buffer[rune, int]) (*quantifier, error) { // {1,}
			from, err := digit(buf)
			if err != nil {
				return nil, err
			}

			_, err = comma(buf)
			if err != nil {
				return nil, err
			}

			_, err = rightBrace(buf)
			if err != nil {
				return nil, err
			}

			return &quantifier{
				From: from,
				To:   nil,
				More: true,
			}, nil
		}),
		func(buf c.Buffer[rune, int]) (*quantifier, error) { // {1}
			from, err := digit(buf)
			if err != nil {
				return nil, err
			}

			_, err = rightBrace(buf)
			if err != nil {
				return nil, err
			}

			return &quantifier{
				From: from,
				More: false,
			}, nil
		},
	)

	parseQuantifier := c.Try(
		c.MapAs(
			map[rune]c.Combinator[rune, int, *quantifier]{
				'?': func(buf c.Buffer[rune, int]) (*quantifier, error) {
					return &quantifier{
						From: 0,
						To:   pointer(1),
						More: false,
					}, nil
				},
				'+': func(buf c.Buffer[rune, int]) (*quantifier, error) {
					return &quantifier{
						From: 1,
						More: true,
					}, nil
				},
				'*': func(buf c.Buffer[rune, int]) (*quantifier, error) {
					return &quantifier{
						From: 0,
						More: true,
					}, nil
				},
				'{': parse,
			},
			any,
		),
	)

	return func(buf c.Buffer[rune, int]) (Node, error) {
		x, err := expression(buf)
		if err != nil {
			return nil, err
		}

		q, err := parseQuantifier(buf)
		if err != nil {
			return x, nil
		}

		q.Value = x
		q.nestedNode = newNestedNode()

		return q, nil
	}
}

func parseCharacter(except ...rune) parser {
	parse := c.NoneOf[rune, int](except...)

	return func(buf c.Buffer[rune, int]) (Node, error) {
		c, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := char{
			Value:      c,
			nestedNode: newNestedNode(),
		}

		return &x, nil
	}
}

func newNestedNode() *nestedNode {
	n := new(nestedNode)
	n.Nested = make(map[string]Node)
	return n
}

func parseMetaCharacters() parser {
	return c.MapAs(
		map[rune]c.Combinator[rune, int, Node]{
			'.': func(buf c.Buffer[rune, int]) (Node, error) {
				x := dot{
					nestedNode: newNestedNode(),
				}

				return &x, nil
			},
			'^': func(buf c.Buffer[rune, int]) (Node, error) {
				x := startOfLine{
					nestedNode: newNestedNode(),
				}

				return &x, nil
			},
			'$': func(buf c.Buffer[rune, int]) (Node, error) {
				x := endOfLine{
					nestedNode: newNestedNode(),
				}

				return &x, nil
			},
		},
		c.Any[rune, int](),
	)
}

func parseEscapedMetaCharacters() parser {
	return c.Skip(
		c.Eq[rune, int]('\\'),
		c.MapAs(
			map[rune]c.Combinator[rune, int, Node]{
				'd': func(buf c.Buffer[rune, int]) (Node, error) {
					x := digit{
						nestedNode: newNestedNode(),
					}

					return &x, nil
				},
				'D': func(buf c.Buffer[rune, int]) (Node, error) {
					x := nonDigit{
						nestedNode: newNestedNode(),
					}

					return &x, nil
				},
				'w': func(buf c.Buffer[rune, int]) (Node, error) {
					x := word{
						nestedNode: newNestedNode(),
					}

					return &x, nil
				},
				'W': func(buf c.Buffer[rune, int]) (Node, error) {
					x := nonWord{
						nestedNode: newNestedNode(),
					}

					return &x, nil
				},
				's': func(buf c.Buffer[rune, int]) (Node, error) {
					x := space{
						nestedNode: newNestedNode(),
					}

					return &x, nil
				},
				'S': func(buf c.Buffer[rune, int]) (Node, error) {
					x := nonSpace{
						nestedNode: newNestedNode(),
					}

					return &x, nil
				},
				'A': func(buf c.Buffer[rune, int]) (Node, error) {
					x := startOfString{
						nestedNode: newNestedNode(),
					}

					return &x, nil
				},
				'z': func(buf c.Buffer[rune, int]) (Node, error) {
					x := endOfString{
						nestedNode: newNestedNode(),
					}

					return &x, nil
				},
			},
			c.Any[rune, int](),
		),
	)
}

func parseGroup(parse c.Combinator[rune, int, *alternation]) parser {
	return parens(
		func(buf c.Buffer[rune, int]) (Node, error) {
			value, err := parse(buf)
			if err != nil {
				return nil, err
			}

			x := &group{
				nestedNode: newNestedNode(),
			}

			// TODO : is it good enough for ID?
			x.uniqID = fmt.Sprintf("%p", x)
			x.Value = value

			return x, nil
		},
	)
}

func parseNotCapturedGroup(parse c.Combinator[rune, int, *alternation]) parser {
	before := SkipString("?:")

	return parens(
		func(buf c.Buffer[rune, int]) (Node, error) {
			_, err := before(buf)
			if err != nil {
				return nil, err
			}

			value, err := parse(buf)
			if err != nil {
				return nil, err
			}

			x := notCapturedGroup{
				Value:      value,
				nestedNode: newNestedNode(),
			}

			return &x, nil
		},
	)
}

func parseNamedGroup(parse c.Combinator[rune, int, *alternation], except ...rune) parser {
	groupName := c.Skip(
		c.Eq[rune, int]('?'),
		angles(
			c.Some(
				0,
				c.Try(c.NoneOf[rune, int](append(except, '>')...)),
			),
		),
	)

	return parens(
		func(buf c.Buffer[rune, int]) (Node, error) {
			name, err := groupName(buf)
			if err != nil {
				return nil, err
			}

			variants, err := parse(buf)
			if err != nil {
				return nil, err
			}

			x := namedGroup{
				Name:       string(name),
				Value:      variants,
				nestedNode: newNestedNode(),
			}

			return &x, nil
		},
	)
}

func parseCharacterClass(table tableParser) parser {
	parse := squares(c.Some(1, table))

	return func(buf c.Buffer[rune, int]) (Node, error) {
		tables, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := characterClass{
			table:      rangetable.Merge(tables...),
			nestedNode: newNestedNode(),
		}

		return &x, nil
	}
}

func parseBracket(name string, predicate func(rune) bool) parser {
	parse := c.SequenceOf[rune, int]([]rune(name)...)

	return func(buf c.Buffer[rune, int]) (Node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		return &bracket{
			key:        name,
			nestedNode: newNestedNode(),
			predicate:  predicate,
		}, nil
	}
}

func parseBrackets() parser {
	alnum := parseBracket(":alnum:", func(x rune) bool {
		return unicode.IsLetter(x) || unicode.IsMark(x) || unicode.IsDigit(x)
	})
	notAlnum := parseBracket(":^alnum:", func(x rune) bool {
		return !(unicode.IsLetter(x) || unicode.IsMark(x) || unicode.IsDigit(x))
	})
	alpha := parseBracket(":alpha:", func(x rune) bool {
		return unicode.IsLetter(x) || unicode.IsMark(x)
	})
	notAlpha := parseBracket(":^alpha:", func(x rune) bool {
		return !(unicode.IsLetter(x) || unicode.IsMark(x))
	})
	ascii := parseBracket(":ascii:", func(x rune) bool {
		return x < unicode.MaxASCII
	})
	notAscii := parseBracket(":^ascii:", func(x rune) bool {
		return x >= unicode.MaxASCII
	})
	blank := parseBracket(":blank:", func(x rune) bool {
		return x == ' ' || x == '\t'
	})
	notBlank := parseBracket(":^blank:", func(x rune) bool {
		return !(x == ' ' || x == '\t')
	})
	digit := parseBracket(":digit:", func(x rune) bool {
		return unicode.IsDigit(x)
	})
	notDigit := parseBracket(":^digit:", func(x rune) bool {
		return !unicode.IsDigit(x)
	})
	lower := parseBracket(":lower:", func(x rune) bool {
		return unicode.IsLower(x)
	})
	notLower := parseBracket(":^lower:", func(x rune) bool {
		return !unicode.IsLower(x)
	})
	upper := parseBracket(":upper:", func(x rune) bool {
		return unicode.IsUpper(x)
	})
	notUpper := parseBracket(":^upper:", func(x rune) bool {
		return !unicode.IsUpper(x)
	})
	space := parseBracket(":space:", func(x rune) bool {
		return unicode.IsSpace(x)
	})
	notSpace := parseBracket(":^space:", func(x rune) bool {
		return !unicode.IsSpace(x)
	})
	cntrl := parseBracket(":cntrl:", func(x rune) bool {
		return unicode.IsControl(x)
	})
	notCntrl := parseBracket(":^cntrl:", func(x rune) bool {
		return !unicode.IsControl(x)
	})
	print := parseBracket(":print:", func(x rune) bool {
		return unicode.IsPrint(x)
	})
	notPrint := parseBracket(":^print:", func(x rune) bool {
		return !unicode.IsPrint(x)
	})
	graph := parseBracket(":graph:", func(x rune) bool {
		return unicode.IsGraphic(x) && !unicode.IsSpace(x)
	})
	notGraph := parseBracket(":^graph:", func(x rune) bool {
		return !(unicode.IsGraphic(x) && !unicode.IsSpace(x))
	})
	punct := parseBracket(":punct:", func(x rune) bool {
		return unicode.IsPunct(x)
	})
	notPunct := parseBracket(":^punct:", func(x rune) bool {
		return !unicode.IsPunct(x)
	})

	return squares(squares(
		choice(
			alnum,
			notAlnum,
			alpha,
			notAlpha,
			ascii,
			notAscii,
			blank,
			notBlank,
			digit,
			notDigit,
			lower,
			notLower,
			upper,
			notUpper,
			space,
			notSpace,
			cntrl,
			notCntrl,
			print,
			notPrint,
			graph,
			notGraph,
			punct,
			notPunct,
		),
	))

	// return func(buf c.Buffer[rune, int]) (Node, error) {
	// return squares(
	// 	ps.MapStrings(
	// 		map[string]parser{
	// 			":xdigit:", func(x rune) bool { return false },
	// 			":word:", func(x rune) bool { return false },
	// 		},
	// 	),
	// )
	// }
}

func parseNegatedCharacterClass(table tableParser) parser {
	parse := squares(
		c.Skip(
			c.Eq[rune, int]('^'),
			c.Some(1, table),
		),
	)

	return func(buf c.Buffer[rune, int]) (Node, error) {
		tables, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := negatedCharacterClass{
			table:      rangetable.Merge(tables...),
			nestedNode: newNestedNode(),
		}

		return &x, nil
	}
}

func parseRangeTable(except ...rune) tableParser {
	item := c.NoneOf[rune, int](except...)
	sep := c.Eq[rune, int]('-')

	return func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
		from, err := item(buf)
		if err != nil {
			return nil, err
		}

		_, err = sep(buf)
		if err != nil {
			return nil, err
		}

		to, err := item(buf)
		if err != nil {
			return nil, err
		}

		// TODO : check range

		runes := make([]rune, 0, to-from)

		for r := from; r <= to; r++ {
			runes = append(runes, r)
		}

		return rangetable.New(runes...), nil
	}
}

func parseEscapedSpecSymbolsTable() tableParser {
	symbols := "[]{}()"
	cases := make(map[rune]tableParser)

	for _, v := range symbols {
		r := v

		cases[r] = func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
			return rangetable.New(r), nil
		}
	}

	return c.Skip(
		c.Eq[rune, int]('\\'),
		c.MapAs(
			cases,
			c.Any[rune, int](),
		),
	)
}

func parseEscapedMetaCharactersTable() tableParser {
	// TODO : move to consts
	runes := make([]rune, 0)
	for x := rune(1); x <= unicode.MaxRune; x++ {
		if !unicode.IsDigit(x) {
			runes = append(runes, x)
		}
	}
	notDigitTable := rangetable.New(runes...)

	runes = make([]rune, 0)
	for x := rune(1); x <= unicode.MaxRune; x++ {
		if !unicode.IsLetter(x) {
			runes = append(runes, x)
		}
	}
	notWordTable := rangetable.New(runes...)

	runes = make([]rune, 0)
	for x := rune(1); x <= unicode.MaxRune; x++ {
		if !unicode.IsSpace(x) {
			runes = append(runes, x)
		}
	}
	notSpaceTable := rangetable.New(runes...)

	return c.Skip(
		c.Eq[rune, int]('\\'),
		c.MapAs(
			map[rune]c.Combinator[rune, int, *unicode.RangeTable]{
				'd': func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
					return unicode.Digit, nil
				},
				'D': func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
					return notDigitTable, nil
				},
				'w': func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
					return unicode.Letter, nil
				},
				'W': func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
					return notWordTable, nil
				},
				's': func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
					return unicode.Space, nil
				},
				'S': func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
					return notSpaceTable, nil
				},
			},
			c.Any[rune, int](),
		),
	)
}

func parseCharacterTable(except ...rune) tableParser {
	parse := c.NoneOf[rune, int](except...)

	return func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
		c, err := parse(buf)
		if err != nil {
			return nil, err
		}

		table := rangetable.New(c)

		return table, nil
	}
}
