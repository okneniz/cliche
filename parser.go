package cliche

import (
	"errors"
	"unicode"

	c "github.com/okneniz/parsec/common"
	"golang.org/x/text/unicode/rangetable"
)

var (
	DefaultParser          = NewParser(DefaultOptions...)
	InvalidQuantifierError = errors.New("target of repeat operator is not specified")
)

type Parser interface {
	Parse(c.Buffer[rune, int]) (Node, error)
}

type Option[T any] func(T)

type CustomParser struct {
	parseExpression       c.Combinator[rune, int, Node]
	parseNestedExpression c.Combinator[rune, int, Node]
	alternationSep        c.Combinator[rune, int, rune]

	// TODO : move it Config / Builder?

	prefixes      map[string]c.Combinator[rune, int, Node]
	prefixParsers *branch[Node]

	inClassPrefixes      map[string]c.Combinator[rune, int, *unicode.RangeTable]
	inClassPrefixParsers *branch[*unicode.RangeTable]

	parsers        []func(except ...rune) c.Combinator[rune, int, Node]
	inClassParsers []func(except ...rune) c.Combinator[rune, int, *unicode.RangeTable]
}

func NewParser(options ...Option[*CustomParser]) *CustomParser {
	p := new(CustomParser)
	p.prefixes = make(map[string]c.Combinator[rune, int, Node])
	p.inClassPrefixes = make(map[string]c.Combinator[rune, int, *unicode.RangeTable])

	// TODO : remove close parens and bracket from escaped chars?
	for _, r := range ".?+*^$[]{}()" { // spec symbols for expression
		x := r

		p.prefixes["\\"+string(r)] = func(buf c.Buffer[rune, int]) (Node, error) {
			return nodeForChar(x), nil
		}
	}

	for _, r := range "^-]\\" { // spec symbols for classes
		x := r

		p.inClassPrefixes["\\"+string(r)] = func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
			return rangetable.New(x), nil
		}
	}

	p.prefixes["."] = func(buf c.Buffer[rune, int]) (Node, error) {
		return newDot(), nil
	}

	p.prefixes["^"] = func(buf c.Buffer[rune, int]) (Node, error) {
		x := startOfLine{
			nestedNode: newNestedNode(),
		}

		return &x, nil
	}

	p.prefixes["$"] = func(buf c.Buffer[rune, int]) (Node, error) {
		x := endOfLine{
			nestedNode: newNestedNode(),
		}

		return &x, nil
	}

	p.prefixes["\\A"] = func(buf c.Buffer[rune, int]) (Node, error) {
		x := startOfString{
			nestedNode: newNestedNode(),
		}

		return &x, nil
	}

	p.prefixes["\\z"] = func(buf c.Buffer[rune, int]) (Node, error) {
		x := endOfString{
			nestedNode: newNestedNode(),
		}

		return &x, nil
	}

	for _, configure := range options {
		configure(p)
	}

	p.prefixParsers = newParserBranches(p.prefixes)
	p.inClassPrefixParsers = newParserBranches(p.inClassPrefixes)

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
		tryAll(
			p.parseNotCapturedGroup(alternation),
			p.parseNamedGroup(alternation),
			p.parseLookAhead(alternation),
			p.parseNegativeLookAhead(alternation),
			p.parseLookBehind(alternation),
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
		tryAll(
			p.parseNotCapturedGroup(alternation),
			p.parseNamedGroup(alternation),
			p.parseLookAhead(alternation),
			p.parseNegativeLookAhead(alternation),
			p.parseLookBehind(alternation),
			p.parseGroup(alternation),
			p.parseInvalidQuantifier(),
			p.parseNodeByPrefixes('|', ')'),
			p.parseNodeByCustomParsers(),
			p.parseCharacterClasses(')'),
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

func (p *CustomParser) parseCharacterClasses(
	except ...rune,
) c.Combinator[rune, int, Node] {
	except = append(except, ']')

	parseTable := c.Choice[rune, int, *unicode.RangeTable](
		c.Try(p.parseRange(except...)),
		c.Try(p.parseCustomTable(except...)),
		c.Try(p.parseCharacterTable(except...)),
		c.Try(p.parseTableByCustomParsers(except...)),
	)

	return tryAll(
		p.parseNegatedCharacterClass(parseTable),
		p.parseCharacterClass(parseTable),
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

func (p *CustomParser) parseCharacter(
	except ...rune,
) c.Combinator[rune, int, Node] {
	parse := c.NoneOf[rune, int](except...)

	return func(buf c.Buffer[rune, int]) (Node, error) {
		x, err := parse(buf)
		if err != nil {
			return nil, err
		}

		return nodeForChar(x), nil
	}
}

// TODO : parse group by prefix '(' too?
func (p *CustomParser) parseGroup(
	parse c.Combinator[rune, int, *alternation],
) c.Combinator[rune, int, Node] {
	return parens(
		func(buf c.Buffer[rune, int]) (Node, error) {
			value, err := parse(buf)
			if err != nil {
				return nil, err
			}

			return newGroup(value), nil
		},
	)
}

// TODO : parse not captured group by prefix '(' too?
func (p *CustomParser) parseNotCapturedGroup(
	parse c.Combinator[rune, int, *alternation],
) c.Combinator[rune, int, Node] {
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

			return newNotCapturedGroup(value), nil
		},
	)
}

// TODO : parse named group by prefix '(' too?
func (p *CustomParser) parseNamedGroup(
	parse c.Combinator[rune, int, *alternation], except ...rune,
) c.Combinator[rune, int, Node] {
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

			return newNamedGroup(string(name), variants), nil
		},
	)
}

func (p *CustomParser) parseLookAhead(
	parse c.Combinator[rune, int, *alternation], except ...rune,
) c.Combinator[rune, int, Node] {
	before := SkipString("?=")

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

			return newLookAhead(value), nil
		},
	)
}

func (p *CustomParser) parseNegativeLookAhead(
	parse c.Combinator[rune, int, *alternation], except ...rune,
) c.Combinator[rune, int, Node] {
	before := SkipString("?!")

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

			return newNegativeLookAhead(value), nil
		},
	)
}

func (p *CustomParser) parseLookBehind(
	parse c.Combinator[rune, int, *alternation], except ...rune,
) c.Combinator[rune, int, Node] {
	before := SkipString("?<=")

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

			n, err := newLookBehind(value)
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
	parseTable c.Combinator[rune, int, *unicode.RangeTable],
) c.Combinator[rune, int, Node] {
	parse := squares(c.Some(1, parseTable))

	return func(buf c.Buffer[rune, int]) (Node, error) {
		tables, err := parse(buf)
		if err != nil {
			return nil, err
		}

		table := rangetable.Merge(tables...)
		return nodeForTable(table), nil
	}
}

// TODO : parse negated character class by prefix '[' too?
func (p *CustomParser) parseNegatedCharacterClass(
	parseTable c.Combinator[rune, int, *unicode.RangeTable],
) c.Combinator[rune, int, Node] {
	parse := squares(
		c.Skip(
			c.Eq[rune, int]('^'),
			c.Some(1, parseTable),
		),
	)

	return func(buf c.Buffer[rune, int]) (Node, error) {
		tables, err := parse(buf)
		if err != nil {
			return nil, err
		}

		table := negatiateTable(rangetable.Merge(tables...))
		return nodeForTable(table), nil
	}
}

// TODO : what about ranges like \u{100}-\u{200} ?
func (p *CustomParser) parseRange(
	except ...rune,
) c.Combinator[rune, int, *unicode.RangeTable] {
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

func (p *CustomParser) parseCharacterTable(
	except ...rune,
) c.Combinator[rune, int, *unicode.RangeTable] {
	parse := c.NoneOf[rune, int](except...)

	return func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
		c, err := parse(buf)
		if err != nil {
			return nil, err
		}

		return rangetable.New(c), nil
	}
}

func (p *CustomParser) parseNodeByPrefixes(
	except ...rune,
) c.Combinator[rune, int, Node] {
	parse := c.NoneOf[rune, int](except...)

	return func(buf c.Buffer[rune, int]) (Node, error) {
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
) c.Combinator[rune, int, Node] {
	parsers := make([]c.Combinator[rune, int, Node], 0, len(p.parsers))
	for _, comb := range p.parsers {
		parsers = append(parsers, comb(except...))
	}

	return tryAll[Node](parsers...)
}

func (p *CustomParser) parseTableByCustomParsers(
	except ...rune,
) c.Combinator[rune, int, *unicode.RangeTable] {
	parsers := make([]c.Combinator[rune, int, *unicode.RangeTable], 0, len(p.inClassParsers))
	for _, comb := range p.inClassParsers {
		parsers = append(parsers, comb(except...))
	}

	return tryAll[*unicode.RangeTable](parsers...)
}

func (p *CustomParser) parseCustomTable(
	except ...rune,
) c.Combinator[rune, int, *unicode.RangeTable] {
	parse := c.NoneOf[rune, int](except...)

	return func(buf c.Buffer[rune, int]) (*unicode.RangeTable, error) {
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
