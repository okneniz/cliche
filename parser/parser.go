package parser

import (
	"errors"

	"github.com/okneniz/cliche/buf"
	"github.com/okneniz/cliche/node"
	c "github.com/okneniz/parsec/common"
)

var (
	InvalidQuantifierError = errors.New("target of repeat operator is not specified")
)

type Option[T any] func(T)

type CustomParser struct {
	parseExpression       c.Combinator[rune, int, node.Node]
	parseNestedExpression c.Combinator[rune, int, node.Node]
	alternationSep        c.Combinator[rune, int, rune]

	prefixes      map[string]c.Combinator[rune, int, node.Node]
	prefixParsers *branch[node.Node]

	inClassPrefixes      map[string]c.Combinator[rune, int, node.Table]
	inClassPrefixParsers *branch[node.Table]

	parsers        []func(except ...rune) c.Combinator[rune, int, node.Node]
	inClassParsers []func(except ...rune) c.Combinator[rune, int, node.Table]
}

// передавать два метод
// table to key
// invert table / negatiate table

func NewParser(options ...Option[*CustomParser]) *CustomParser {
	p := new(CustomParser)
	p.prefixes = make(map[string]c.Combinator[rune, int, node.Node])
	p.inClassPrefixes = make(map[string]c.Combinator[rune, int, node.Table])

	// TODO : remove close parens and bracket from escaped chars?
	for _, r := range ".?+*^$[]{}()" { // spec symbols for expression
		x := r

		p.prefixes["\\"+string(r)] = func(buf c.Buffer[rune, int]) (node.Node, error) {
			return node.NodeForTable(NewUnicodeTableFor(x)), nil
		}
	}

	for _, r := range "^-]\\" { // spec symbols for classes
		x := r

		p.inClassPrefixes["\\"+string(r)] = func(buf c.Buffer[rune, int]) (node.Table, error) {
			return NewUnicodeTableFor(x), nil
		}
	}

	for _, configure := range options {
		configure(p)
	}

	p.prefixParsers = newParserBranches(p.prefixes)
	p.inClassPrefixParsers = newParserBranches(p.inClassPrefixes)

	p.alternationSep = c.Eq[rune, int]('|')

	// parse alternation
	alternation := func(buf c.Buffer[rune, int]) (node.Alternation, error) {
		variant, err := p.parseNestedExpression(buf)
		if err != nil {
			return nil, err
		}

		variants := make([]node.Node, 0, 1)
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

		return node.NewAlternation(variants), nil
	}

	// parse node
	parseNode := p.parseOptionalQuantifier(
		TryAll(
			p.parseNotCapturedGroup(alternation),
			p.parseNamedGroup(alternation),
			p.parseLookAhead(alternation),
			p.parseNegativeLookAhead(alternation),
			p.parseLookBehind(alternation),
			p.parseNegativeLookBehind(alternation),
			p.parseGroup(alternation),
			p.parseInvalidQuantifier(),
			p.parseNodeByPrefixes('|'),
			p.parseNodeByCustomParsers('|'),
			p.parseCharacterClasses('|'),
			p.parseCharacter('|'),
		),
	)

	// parse node of nested expression
	parseNestedNode := p.parseOptionalQuantifier(
		TryAll(
			p.parseNotCapturedGroup(alternation),
			p.parseNamedGroup(alternation),
			p.parseLookAhead(alternation),
			p.parseNegativeLookAhead(alternation),
			p.parseLookBehind(alternation),
			p.parseNegativeLookBehind(alternation),
			p.parseGroup(alternation),
			p.parseInvalidQuantifier(),
			p.parseNodeByPrefixes('|', ')'),
			p.parseNodeByCustomParsers(),
			p.parseCharacterClasses(')'),
			p.parseCharacter('|', ')'),
		),
	)

	p.parseExpression = func(buf c.Buffer[rune, int]) (node.Node, error) {
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

	p.parseNestedExpression = func(buf c.Buffer[rune, int]) (node.Node, error) {
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

func (p *CustomParser) Parse(str string) (node.Node, error) {
	buffer := buf.NewRunesBuffer(str)

	newNode, err := p.parseAlternationOrExpression(buffer)
	if err != nil {
		return nil, err
	}

	newNode.Traverse(func(x node.Node) {
		if len(x.GetNestedNodes()) == 0 { // its leaf
			x.AddExpression(str)
		}
	})

	return newNode, nil
}

func (p *CustomParser) parseAlternationOrExpression(
	buf c.Buffer[rune, int],
) (node.Node, error) {
	expression, err := p.parseExpression(buf)
	if err != nil {
		return nil, err
	}
	if buf.IsEOF() {
		return expression, nil
	}

	variants := make([]node.Node, 0, 1)
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

	return node.NewAlternation(variants), nil
}

func (p *CustomParser) parseCharacterClasses(
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	except = append(except, ']')

	parseTable := c.Choice[rune, int, node.Table](
		c.Try(p.parseRange(except...)),
		c.Try(p.parseCustomTable(except...)),
		c.Try(p.parseCharacterTable(except...)),
		c.Try(p.parseTableByCustomParsers(except...)),
	)

	return TryAll(
		p.parseNegatedCharacterClass(parseTable),
		p.parseCharacterClass(parseTable),
	)
}

func (p *CustomParser) parseInvalidQuantifier() c.Combinator[rune, int, node.Node] {
	invalidChars := map[rune]struct{}{
		'?': {},
		'*': {},
		'+': {},
	}

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
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
	expression c.Combinator[rune, int, node.Node],
) c.Combinator[rune, int, node.Node] {
	any := c.Any[rune, int]()
	digit := c.Try(Number())
	comma := c.Try(c.Eq[rune, int](','))
	rightBrace := c.Eq[rune, int]('}')

	parse := c.Try(
		c.MapAs(
			map[rune]c.Combinator[rune, int, *node.Quantity]{
				'?': func(buf c.Buffer[rune, int]) (*node.Quantity, error) {
					return node.NewQuantity(0, 1), nil
				},
				'+': func(buf c.Buffer[rune, int]) (*node.Quantity, error) {
					return node.NewEndlessQuantity(1), nil
				},
				'*': func(buf c.Buffer[rune, int]) (*node.Quantity, error) {
					return node.NewEndlessQuantity(0), nil
				},
				'{': c.Choice(
					c.Try(func(buf c.Buffer[rune, int]) (*node.Quantity, error) { // {1,1}
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

						return node.NewQuantity(from, to), nil
					}),
					c.Try(func(buf c.Buffer[rune, int]) (*node.Quantity, error) { // {,1}
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

						return node.NewQuantity(0, to), nil
					}),
					c.Try(func(buf c.Buffer[rune, int]) (*node.Quantity, error) { // {1,}
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

						return node.NewEndlessQuantity(from), nil
					}),
					func(buf c.Buffer[rune, int]) (*node.Quantity, error) { // {1}
						from, err := digit(buf)
						if err != nil {
							return nil, err
						}

						_, err = rightBrace(buf)
						if err != nil {
							return nil, err
						}

						return node.NewQuantity(from, from), nil
					},
				),
			},
			any,
		),
	)

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		x, err := expression(buf)
		if err != nil {
			return nil, err
		}

		q, err := parse(buf)
		if err != nil {
			return x, nil
		}

		return node.NewQuantifier(q, x), nil
	}
}

func (p *CustomParser) parseCharacter(
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parse := c.NoneOf[rune, int](except...)

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		x, err := parse(buf)
		if err != nil {
			return nil, err
		}

		return node.NodeForTable(NewUnicodeTableFor(x)), nil
	}
}

// TODO : parse group by prefix '(' too?
func (p *CustomParser) parseGroup(
	parse c.Combinator[rune, int, node.Alternation],
) c.Combinator[rune, int, node.Node] {
	return Parens(
		func(buf c.Buffer[rune, int]) (node.Node, error) {
			value, err := parse(buf)
			if err != nil {
				return nil, err
			}

			return node.NewGroup(value), nil
		},
	)
}

// TODO : parse not captured group by prefix '(' too?
func (p *CustomParser) parseNotCapturedGroup(
	parse c.Combinator[rune, int, node.Alternation],
) c.Combinator[rune, int, node.Node] {
	before := SkipString("?:")

	return Parens(
		func(buf c.Buffer[rune, int]) (node.Node, error) {
			_, err := before(buf)
			if err != nil {
				return nil, err
			}

			value, err := parse(buf)
			if err != nil {
				return nil, err
			}

			return node.NewNotCapturedGroup(value), nil
		},
	)
}

// TODO : parse named group by prefix '(' too?
func (p *CustomParser) parseNamedGroup(
	parse c.Combinator[rune, int, node.Alternation], except ...rune,
) c.Combinator[rune, int, node.Node] {
	groupName := c.Skip(
		c.Eq[rune, int]('?'),
		Angles(
			c.Some(
				0,
				c.Try(c.NoneOf[rune, int](append(except, '>')...)),
			),
		),
	)

	return Parens(
		func(buf c.Buffer[rune, int]) (node.Node, error) {
			name, err := groupName(buf)
			if err != nil {
				return nil, err
			}

			variants, err := parse(buf)
			if err != nil {
				return nil, err
			}

			return node.NewNamedGroup(string(name), variants), nil
		},
	)
}

func (p *CustomParser) parseLookAhead(
	parse c.Combinator[rune, int, node.Alternation], except ...rune,
) c.Combinator[rune, int, node.Node] {
	before := SkipString("?=")

	return Parens(
		func(buf c.Buffer[rune, int]) (node.Node, error) {
			_, err := before(buf)
			if err != nil {
				return nil, err
			}

			value, err := parse(buf)
			if err != nil {
				return nil, err
			}

			return node.NewLookAhead(value), nil
		},
	)
}

func (p *CustomParser) parseNegativeLookAhead(
	parse c.Combinator[rune, int, node.Alternation], except ...rune,
) c.Combinator[rune, int, node.Node] {
	before := SkipString("?!")

	return Parens(
		func(buf c.Buffer[rune, int]) (node.Node, error) {
			_, err := before(buf)
			if err != nil {
				return nil, err
			}

			value, err := parse(buf)
			if err != nil {
				return nil, err
			}

			return node.NewNegativeLookAhead(value), nil
		},
	)
}

func (p *CustomParser) parseLookBehind(
	parse c.Combinator[rune, int, node.Alternation], except ...rune,
) c.Combinator[rune, int, node.Node] {
	before := SkipString("?<=")

	return Parens(
		func(buf c.Buffer[rune, int]) (node.Node, error) {
			_, err := before(buf)
			if err != nil {
				return nil, err
			}

			value, err := parse(buf)
			if err != nil {
				return nil, err
			}

			n, err := node.NewLookBehind(value)
			if err != nil {
				// TODO : return explanation from parser
				// handle not inly NothingMatched error
				panic(err)
			}

			return n, nil
		},
	)
}

func (p *CustomParser) parseNegativeLookBehind(
	parse c.Combinator[rune, int, node.Alternation], except ...rune,
) c.Combinator[rune, int, node.Node] {
	before := SkipString("?<!")

	return Parens(
		func(buf c.Buffer[rune, int]) (node.Node, error) {
			_, err := before(buf)
			if err != nil {
				return nil, err
			}

			value, err := parse(buf)
			if err != nil {
				return nil, err
			}

			n, err := node.NewNegativeLookBehind(value)
			if err != nil {
				// TODO : return explanation from parser
				// handle not inly NothingMatched error
				panic(err)
			}

			return n, nil
		},
	)
}

// TODO : parse character class by prefix '[' too?
func (p *CustomParser) parseCharacterClass(
	parseTable c.Combinator[rune, int, node.Table],
) c.Combinator[rune, int, node.Node] {
	parse := Squares(c.Some(1, parseTable))

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		tables, err := parse(buf)
		if err != nil {
			return nil, err
		}

		table := MergeUnicodeTables(tables...)
		return node.NodeForTable(table), nil
	}
}

// TODO : parse negated character class by prefix '[' too?
func (p *CustomParser) parseNegatedCharacterClass(
	parseTable c.Combinator[rune, int, node.Table],
) c.Combinator[rune, int, node.Node] {
	parse := Squares(
		c.Skip(
			c.Eq[rune, int]('^'),
			c.Some(1, parseTable),
		),
	)

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		tables, err := parse(buf)
		if err != nil {
			return nil, err
		}

		table := MergeUnicodeTables(tables...).Invert()
		return node.NodeForTable(table), nil
	}
}

// TODO : what about ranges like \u{100}-\u{200} ?
func (p *CustomParser) parseRange(
	except ...rune,
) c.Combinator[rune, int, node.Table] {
	item := c.NoneOf[rune, int](except...)
	sep := c.Eq[rune, int]('-')

	return func(buf c.Buffer[rune, int]) (node.Table, error) {
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

		return NewUnicodeTableFor(runes...), nil
	}
}

func (p *CustomParser) parseCharacterTable(
	except ...rune,
) c.Combinator[rune, int, node.Table] {
	parse := c.NoneOf[rune, int](except...)

	return func(buf c.Buffer[rune, int]) (node.Table, error) {
		c, err := parse(buf)
		if err != nil {
			return nil, err
		}

		return NewUnicodeTableFor(c), nil
	}
}

func (p *CustomParser) parseNodeByPrefixes(
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parse := c.NoneOf[rune, int](except...)

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		current := p.prefixParsers.Children

		for len(current) > 0 {
			r, err := parse(buf)
			if err != nil {
				return nil, c.NothingMatched
			}

			next, exists := current[r]
			if !exists {
				return nil, c.NothingMatched
			}

			if next.parser != nil {
				return next.parser(buf)
			}

			current = next.Children
		}

		return nil, c.NothingMatched
	}
}

func (p *CustomParser) parseNodeByCustomParsers(
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parsers := make([]c.Combinator[rune, int, node.Node], 0, len(p.parsers))
	for _, comb := range p.parsers {
		parsers = append(parsers, comb(except...))
	}

	return TryAll[node.Node](parsers...)
}

func (p *CustomParser) parseTableByCustomParsers(
	except ...rune,
) c.Combinator[rune, int, node.Table] {
	parsers := make([]c.Combinator[rune, int, node.Table], 0, len(p.inClassParsers))
	for _, comb := range p.inClassParsers {
		parsers = append(parsers, comb(except...))
	}

	return TryAll[node.Table](parsers...)
}

func (p *CustomParser) parseCustomTable(
	except ...rune,
) c.Combinator[rune, int, node.Table] {
	parse := c.NoneOf[rune, int](except...)

	return func(buf c.Buffer[rune, int]) (node.Table, error) {
		current := p.inClassPrefixParsers.Children

		for len(current) > 0 {
			r, err := parse(buf)
			if err != nil {
				return nil, c.NothingMatched
			}

			next, exists := current[r]
			if !exists {
				return nil, c.NothingMatched
			}

			if next.parser != nil {
				return next.parser(buf)
			}

			current = next.Children
		}

		return nil, c.NothingMatched
	}
}
