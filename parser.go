package cliche

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	c "github.com/okneniz/parsec/common"
	"golang.org/x/text/unicode/rangetable"
)

var (
	DefaultParser          = NewParser()
	InvalidQuantifierError = errors.New("target of repeat operator is not specified")
)

type Parser interface {
	Parse(c.Buffer[rune, int]) (Node, error)
}

type CustomParser struct {
	parseExpression       c.Combinator[rune, int, Node]
	parseNestedExpression c.Combinator[rune, int, Node]
	alternationSep        c.Combinator[rune, int, rune] // TODO : better name?
}

func NewParser() *CustomParser {
	p := new(CustomParser)
	p.alternationSep = c.Eq[rune, int]('|')

	// parse alternation
	alternation := func(buf c.Buffer[rune, int]) (*alternation, error) {
		variant, err := p.parseNestedExpression(buf)
		if err != nil {
			return nil, err
		}

		variants := make([]Node, 0, 1)
		variants = append(variants, variant)

		for !buf.IsEOF() {
			pos := buf.Position()

			_, err = p.alternationSep(buf)
			if err != nil {
				buf.Seek(pos)
				break // return error instead break?
			}

			variant, err = p.parseNestedExpression(buf)
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
	parseNode := p.parseOptionalQuantifier(
		choice(
			p.parseBrackets(),
			p.parseCharacterClasses('|'),
			p.parseNotCapturedGroup(alternation),
			p.parseNamedGroup(alternation),
			p.parseGroup(alternation),
			p.parseInvalidQuantifier(),
			p.parseEscapedMetaCharacters(),
			p.parseMetaCharacters(),
			p.parseEscapedSpecSymbols(),
			p.parseCharacter('|'),
		),
	)

	// parse node of nested expression
	parseNestedNode := p.parseOptionalQuantifier(
		choice(
			p.parseCharacterClasses('|', ')'),
			p.parseNotCapturedGroup(alternation),
			p.parseNamedGroup(alternation),
			p.parseGroup(alternation),
			p.parseInvalidQuantifier(),
			p.parseEscapedMetaCharacters(),
			p.parseMetaCharacters(),
			p.parseEscapedSpecSymbols(),
			p.parseCharacter('|', ')'),
		),
	)

	p.parseExpression = func(buf c.Buffer[rune, int]) (Node, error) {
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

	p.parseNestedExpression = func(buf c.Buffer[rune, int]) (Node, error) {
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

	return p
}

func (p *CustomParser) Parse(buf c.Buffer[rune, int]) (Node, error) {
	expression, err := p.parseExpression(buf)
	if err != nil {
		return nil, err
	}
	if buf.IsEOF() {
		return expression, nil
	}

	variants := make([]Node, 0, 1)
	variants = append(variants, expression)

	for !buf.IsEOF() {
		_, err = p.alternationSep(buf)
		if err != nil {
			return nil, err
		}

		expression, err = p.parseExpression(buf)
		if err != nil {
			return nil, err
		}

		variants = append(variants, expression)
	}

	return newAlternation(variants), nil
}

func (p *CustomParser) parseCharacterClasses(except ...rune) c.Combinator[rune, int, Node] {
	parseTable := c.Choice[rune, int, *unicode.RangeTable](
		c.Try(p.parseRangeTable(append(except, ']')...)),
		c.Try(p.parseEscapedMetaCharactersTable()),
		c.Try(p.parseEscapedSpecSymbolsTable()),
		c.Try(p.parseCharacterTable(append(except, ']')...)),
	)

	return choice(
		p.parseNegatedCharacterClass(parseTable),
		p.parseCharacterClass(parseTable),
	)
}

func (p *CustomParser) parseEscapedSpecSymbols() c.Combinator[rune, int, Node] {
	symbols := ".?+*^$[]{}()"
	cases := make(map[rune]c.Combinator[rune, int, Node])

	for _, v := range symbols {
		x := v

		cases[x] = func(buf c.Buffer[rune, int]) (Node, error) {
			return &simpleNode{
				key: string(x),
				predicate: func(r rune) bool {
					return r == x
				},
				nestedNode: newNestedNode(),
			}, nil
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

func (p *CustomParser) parseInvalidQuantifier() c.Combinator[rune, int, Node] {
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

func (p *CustomParser) parseOptionalQuantifier(
	expression c.Combinator[rune, int, Node],
) c.Combinator[rune, int, Node] {
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

func (p *CustomParser) parseCharacter(except ...rune) c.Combinator[rune, int, Node] {
	parse := c.NoneOf[rune, int](except...)

	return func(buf c.Buffer[rune, int]) (Node, error) {
		x, err := parse(buf)
		if err != nil {
			return nil, err
		}

		return &simpleNode{
			key: string(x),
			predicate: func(r rune) bool {
				return r == x
			},
			nestedNode: newNestedNode(),
		}, nil
	}
}

func (p *CustomParser) parseMetaCharacters() c.Combinator[rune, int, Node] {
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

func (p *CustomParser) parseEscapedMetaCharacters() c.Combinator[rune, int, Node] {
	not := func(p func(rune) bool) func(rune) bool {
		return func(x rune) bool {
			return !p(x)
		}
	}

	propertyTable := c.Try(p.parsePropertyName())
	parseHexChar := c.Try(p.parseHexNumber(2, 2))
	parseUnicodeChar := c.Try(p.parseHexNumber(1, 4))
	parseOctalChar := c.Try(braces(p.parseOctal(3)))

	return c.Skip(
		c.Eq[rune, int]('\\'),
		c.MapAs(
			map[rune]c.Combinator[rune, int, Node]{
				'd': func(buf c.Buffer[rune, int]) (Node, error) {
					return &simpleNode{
						key:        "\\d",
						predicate:  unicode.IsDigit,
						nestedNode: newNestedNode(),
					}, nil
				},
				'D': func(buf c.Buffer[rune, int]) (Node, error) {
					return &simpleNode{
						key:        "\\D",
						predicate:  not(unicode.IsDigit),
						nestedNode: newNestedNode(),
					}, nil
				},
				'w': func(buf c.Buffer[rune, int]) (Node, error) {
					return &simpleNode{
						key:        "\\w",
						predicate:  isWord,
						nestedNode: newNestedNode(),
					}, nil
				},
				'W': func(buf c.Buffer[rune, int]) (Node, error) {
					return &simpleNode{
						key:        "\\w",
						predicate:  not(isWord),
						nestedNode: newNestedNode(),
					}, nil
				},
				's': func(buf c.Buffer[rune, int]) (Node, error) {
					return &simpleNode{
						key:        "\\s",
						predicate:  unicode.IsSpace,
						nestedNode: newNestedNode(),
					}, nil
				},
				'S': func(buf c.Buffer[rune, int]) (Node, error) {
					return &simpleNode{
						key:        "\\s",
						predicate:  not(unicode.IsSpace),
						nestedNode: newNestedNode(),
					}, nil
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
				'h': func(buf c.Buffer[rune, int]) (Node, error) {
					return &simpleNode{
						key:        "\\h",
						predicate:  isHex,
						nestedNode: newNestedNode(),
					}, nil
				},
				'H': func(buf c.Buffer[rune, int]) (Node, error) {
					return &simpleNode{
						key:        "\\H",
						predicate:  not(isHex),
						nestedNode: newNestedNode(),
					}, nil
				},
				'p': c.Try(func(buf c.Buffer[rune, int]) (Node, error) {
					table, err := propertyTable(buf)
					if err != nil {
						return nil, err
					}

					return &simpleNode{
						key:        rangeTableKey(table),
						nestedNode: newNestedNode(),
						predicate: func(x rune) bool {
							return unicode.In(x, table)
						},
					}, nil
				}),
				'P': c.Try(func(buf c.Buffer[rune, int]) (Node, error) {
					table, err := propertyTable(buf)
					if err != nil {
						return nil, err
					}

					negatiatedTable := negatiateTable(table)

					return &simpleNode{
						key:        rangeTableKey(negatiatedTable),
						nestedNode: newNestedNode(),
						predicate: func(x rune) bool {
							return unicode.In(x, negatiatedTable)
						},
					}, nil
				}),
				'x': c.Try(func(buf c.Buffer[rune, int]) (Node, error) {
					num, err := parseHexChar(buf)
					if err != nil {
						return nil, err
					}

					r := rune(num)

					// TODO : check bounds

					return &simpleNode{
						key:        string(r),
						nestedNode: newNestedNode(),
						predicate: func(x rune) bool {
							return x == r
						},
					}, nil
				}),
				'o': c.Try(func(buf c.Buffer[rune, int]) (Node, error) {
					num, err := parseOctalChar(buf)
					if err != nil {
						return nil, err
					}

					r := rune(num)

					// TODO : check bounds

					return &simpleNode{
						key:        string(r),
						nestedNode: newNestedNode(),
						predicate: func(x rune) bool {
							return x == r
						},
					}, nil
				}),
				'u': c.Try(func(buf c.Buffer[rune, int]) (Node, error) {
					num, err := parseUnicodeChar(buf)
					if err != nil {
						return nil, err
					}

					r := rune(num)

					// TODO : check bounds

					return &simpleNode{
						key:        string(r),
						nestedNode: newNestedNode(),
						predicate: func(x rune) bool {
							return x == r
						},
					}, nil
				}),
			},
			c.Any[rune, int](),
		),
	)
}

func (p *CustomParser) parseHexNumber(from, to int) c.Combinator[rune, int, int] {
	parse := Quantifier(from, to, c.OneOf[rune, int]([]rune("0123456789abcdefABCDEF")...))

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

func (p *CustomParser) parseOctal(size int) c.Combinator[rune, int, int] {
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

func (p *CustomParser) parsePropertyName() c.Combinator[rune, int, *unicode.RangeTable] {
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

	return choice(cases...)
}

func (p *CustomParser) parseGroup(parse c.Combinator[rune, int, *alternation]) c.Combinator[rune, int, Node] {
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

func (p *CustomParser) parseNotCapturedGroup(parse c.Combinator[rune, int, *alternation]) c.Combinator[rune, int, Node] {
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

func (p *CustomParser) parseNamedGroup(parse c.Combinator[rune, int, *alternation], except ...rune) c.Combinator[rune, int, Node] {
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

func (p *CustomParser) parseBracket(name string, predicate func(rune) bool) c.Combinator[rune, int, Node] {
	parse := c.SequenceOf[rune, int]([]rune(name)...)

	return func(buf c.Buffer[rune, int]) (Node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		// TODO : use tables to better compaction
		return &simpleNode{
			key:        name, // TODO : sometimes use another key to compact tree
			nestedNode: newNestedNode(),
			predicate:  predicate,
		}, nil
	}
}

func (p *CustomParser) parseBrackets() c.Combinator[rune, int, Node] {
	alnum := p.parseBracket(":alnum:", func(x rune) bool {
		return unicode.IsLetter(x) || unicode.IsMark(x) || unicode.IsDigit(x)
	})
	notAlnum := p.parseBracket(":^alnum:", func(x rune) bool {
		return !(unicode.IsLetter(x) || unicode.IsMark(x) || unicode.IsDigit(x))
	})
	alpha := p.parseBracket(":alpha:", func(x rune) bool {
		return unicode.IsLetter(x) || unicode.IsMark(x)
	})
	notAlpha := p.parseBracket(":^alpha:", func(x rune) bool {
		return !(unicode.IsLetter(x) || unicode.IsMark(x))
	})
	ascii := p.parseBracket(":ascii:", func(x rune) bool {
		return x < unicode.MaxASCII
	})
	notAscii := p.parseBracket(":^ascii:", func(x rune) bool {
		return x >= unicode.MaxASCII
	})
	blank := p.parseBracket(":blank:", func(x rune) bool {
		return x == ' ' || x == '\t'
	})
	notBlank := p.parseBracket(":^blank:", func(x rune) bool {
		return !(x == ' ' || x == '\t')
	})
	digit := p.parseBracket(":digit:", func(x rune) bool {
		return unicode.IsDigit(x)
	})
	notDigit := p.parseBracket(":^digit:", func(x rune) bool {
		return !unicode.IsDigit(x)
	})
	lower := p.parseBracket(":lower:", func(x rune) bool {
		return unicode.IsLower(x)
	})
	notLower := p.parseBracket(":^lower:", func(x rune) bool {
		return !unicode.IsLower(x)
	})
	upper := p.parseBracket(":upper:", func(x rune) bool {
		return unicode.IsUpper(x)
	})
	notUpper := p.parseBracket(":^upper:", func(x rune) bool {
		return !unicode.IsUpper(x)
	})
	space := p.parseBracket(":space:", func(x rune) bool {
		return unicode.IsSpace(x)
	})
	notSpace := p.parseBracket(":^space:", func(x rune) bool {
		return !unicode.IsSpace(x)
	})
	cntrl := p.parseBracket(":cntrl:", func(x rune) bool {
		return unicode.IsControl(x)
	})
	notCntrl := p.parseBracket(":^cntrl:", func(x rune) bool {
		return !unicode.IsControl(x)
	})
	print := p.parseBracket(":print:", func(x rune) bool {
		return unicode.IsPrint(x)
	})
	notPrint := p.parseBracket(":^print:", func(x rune) bool {
		return !unicode.IsPrint(x)
	})
	graph := p.parseBracket(":graph:", func(x rune) bool {
		return unicode.IsGraphic(x) && !unicode.IsSpace(x)
	})
	notGraph := p.parseBracket(":^graph:", func(x rune) bool {
		return !(unicode.IsGraphic(x) && !unicode.IsSpace(x))
	})
	punct := p.parseBracket(":punct:", func(x rune) bool {
		return unicode.IsPunct(x)
	})
	notPunct := p.parseBracket(":^punct:", func(x rune) bool {
		return !unicode.IsPunct(x)
	})
	xdigit := p.parseBracket(":xdigit:", func(x rune) bool {
		return isHex(x)
	})
	notXdigit := p.parseBracket(":^xdigit:", func(x rune) bool {
		return !isHex(x)
	})
	word := p.parseBracket(":word:", func(x rune) bool {
		return isWord(x)
	})
	notWord := p.parseBracket(":^word:", func(x rune) bool {
		return !isWord(x)
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
			xdigit,
			notXdigit,
			word,
			notWord,
		),
	))
}

func (p *CustomParser) parseEscapedSpecSymbolsTable() c.Combinator[rune, int, *unicode.RangeTable] {
	symbols := "[]{}()"
	cases := make(map[rune]c.Combinator[rune, int, *unicode.RangeTable])

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

func (p *CustomParser) parseEscapedMetaCharactersTable() c.Combinator[rune, int, *unicode.RangeTable] {
	notDigitTable := negatiateTable(unicode.Digit)

	runes := make([]rune, 0)
	for x := rune(1); x <= unicode.MaxRune; x++ {
		if isWord(x) {
			runes = append(runes, x)
		}
	}
	wordTable := rangetable.New(runes...)

	runes = make([]rune, 0)
	for x := rune(1); x <= unicode.MaxRune; x++ {
		if !isWord(x) {
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

	runes = make([]rune, 0)
	for x := rune(1); x <= unicode.MaxRune; x++ {
		if isHex(x) {
			runes = append(runes, x)
		}
	}
	hexTable := rangetable.New(runes...)

	runes = make([]rune, 0)
	for x := rune(1); x <= unicode.MaxRune; x++ {
		if !isHex(x) {
			runes = append(runes, x)
		}
	}
	notHexTable := rangetable.New(runes...)

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
					return wordTable, nil
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
				'h': func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
					return hexTable, nil
				},
				'H': func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
					return notHexTable, nil
				},
			},
			c.Any[rune, int](),
		),
	)
}

func (p *CustomParser) parseCharacterClass(
	table c.Combinator[rune, int, *unicode.RangeTable],
) c.Combinator[rune, int, Node] {
	parse := squares(c.Some(1, table))

	return func(buf c.Buffer[rune, int]) (Node, error) {
		tables, err := parse(buf)
		if err != nil {
			return nil, err
		}

		table := rangetable.Merge(tables...)

		return &simpleNode{
			key:        rangeTableKey(table),
			nestedNode: newNestedNode(),
			predicate: func(x rune) bool {
				return unicode.In(x, table)
			},
		}, nil
	}
}

// TODO : check this type of compaction in test
func (p *CustomParser) parseNegatedCharacterClass(
	table c.Combinator[rune, int, *unicode.RangeTable],
) c.Combinator[rune, int, Node] {
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

		table := rangetable.Merge(tables...)

		runes := make([]rune, 0)
		for x := rune(1); x <= unicode.MaxRune; x++ {
			if !unicode.In(x, table) {
				runes = append(runes, x)
			}
		}
		negatedTable := rangetable.New(runes...)

		return &simpleNode{
			key:        rangeTableKey(negatedTable),
			nestedNode: newNestedNode(),
			predicate: func(x rune) bool {
				return unicode.In(x, negatedTable)
			},
		}, nil
	}
}

func (p *CustomParser) parseRangeTable(except ...rune) c.Combinator[rune, int, *unicode.RangeTable] {
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

func (p *CustomParser) parseCharacterTable(except ...rune) c.Combinator[rune, int, *unicode.RangeTable] {
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
