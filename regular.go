package regular

import (
	"errors"

	c "github.com/okneniz/parsec/common"
	s "github.com/okneniz/parsec/strings"
)

type node interface {
	// Key() rune
	// Push(node)
	// Scan()
	// ToSlice()
	// ToMap()
	// Leafs()
}

type root struct {
	children map[string]node
}

type group struct {
	variants [][]node
	leaf bool
	children map[string]node
}

type namedGroup struct {
	name string
	variants [][]node
	leaf bool
	children map[string]node
}

type notCapturedGroup struct {
	variants [][]node
	leaf bool
	children map[string]node
}

type char struct {
	value rune
	leaf bool
	children map[string]node
}

type dot struct {
	leaf bool
	children map[string]node
}

type digit struct {
	leaf bool
	children map[string]node
}

type nonDigit struct {
	leaf bool
	children map[string]node
}

type word struct {
	leaf bool
	children map[string]node
}

type nonWord struct {
	leaf bool
	children map[string]node
}

type space struct {
	leaf bool
	children map[string]node
}

type nonSpace struct {
	leaf bool
	children map[string]node
}

type startOfLine struct {
	leaf bool
	children map[string]node
}

type endOfLine struct {
	leaf bool
	children map[string]node
}

type startOfString struct {
	leaf bool
	children map[string]node
}

type endOfString struct {
	leaf bool
	children map[string]node
}

type rangeNode struct {
	from rune
	to rune
	children map[string]node
	leaf bool
}

type quantifier struct {
	from int
	to *int
	more bool
	expression node
	leaf bool
	children map[string]node
}

type positiveSet struct {
	value []node
	leaf bool
	children map[string]node
}

type negativeSet struct {
	value []node
	leaf bool
	children map[string]node
}

var (
	defaultParser = parseRegexp()
)

type Expression interface {
	// Match(string) (bool, error)
}

func Parse(data string) (Expression, error) {
	buf := s.Buffer([]rune(data))

	expression, err := defaultParser(buf)
	if err != nil {
		return nil, err
	}

	return expression, nil
}

type parser = c.Combinator[rune, s.Position, node]
type expressionParser = c.Combinator[rune, s.Position, []node]
type buffer = c.Buffer[rune, s.Position]

var none = struct{}{}

func parseRegexp(except ...rune) expressionParser {
	var (
		regexp expressionParser
		groups parser
	)

	if len(except) == 0 {
		groups = choice(
			parseNotCapturedGroup(regexp),
			parseNamedGroup(regexp),
			parseGroup(regexp),
		)
	} else {
		nestedRegexp := parseRegexp(append(except, ')', '|')...)

		groups = choice(
			parseNotCapturedGroup(nestedRegexp),
			parseNamedGroup(nestedRegexp),
			parseGroup(nestedRegexp),
		)
	}

	characters := choice(
		parseInvalidQuantifier(),
		parseEscapedMetacharacters(),
		parseDot(),
		parseDigit(),
		parseNonDigit(),
		parseWord(),
		parseNonWord(),
		parseSpace(),
		parseNonSpace(),
		parseStartOfLine(),
		parseEndOfLine(),
		parseStartOfString(),
		parseEndOfString(),
		parseCharacter(except...),
	)

	setsCombinatrors := choice( // where dot?
		parseRange(append(except, ']')...),
		parseEscapedMetacharacters(),
		parseDigit(),
		parseNonDigit(),
		parseWord(),
		parseNonWord(),
		parseSpace(),
		parseNonSpace(),
		parseStartOfLine(),
		parseEndOfLine(),
		parseStartOfString(),
		parseEndOfString(),
		parseCharacter(except...),
	)

	sets := choice(
		parsePositiveSet(setsCombinatrors),
		parseNegativeSet(setsCombinatrors),
	)

	parse := s.Some(
		0,
		parseOptionalQuantifier(
			choice(
				sets,
				groups,
				characters,
			),
		),
	)

	return parse
}

func choice(parsers ...parser) parser {
	attempts := make([]parser, len(parsers))

	for i, parse := range parsers {
		attempts[i] = c.Try(parse)
	}

	return c.Choice(attempts...)
}

var (
	InvalidQuantifierError = errors.New("target of repeat operator is not specified")
)

func parseEscapedMetacharacters() parser {
	chars := ".?+*^$[]{}()"
	parsers := make([]parser, len(chars))

	for i, c := range chars {
		parsers[i] = parseEscapedMetacharacter(c)
	}

	return choice(parsers...)
}

func parseEscapedMetacharacter(value rune) parser {
	str := string([]rune{'\\', value})
	parse := SkipString(str)

	return func(buf buffer) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := char{
			value: value,
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseInvalidQuantifier() parser {
	invalidChars := map[rune]struct{}{
		'?': {},
		'*': {},
		'+': {},
	}

	return func(buf buffer) (node, error) {
		x, err := buf.Read(true)
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
	digit := c.Try(s.Unsigned[int]())
	lookup := s.Satisfy(false, c.Anything[rune])
	skip := s.Any()

	parseQuantifier := c.Try(
		c.MapAs(
			map[rune]c.Combinator[rune, s.Position, quantifier]{
				'?': func(buf buffer) (quantifier, error) {
					q := quantifier{}

					_, err := skip(buf)
					if err != nil {
						return q, err
					}

					to := 1
					q.from = 0
					q.to = &to
					q.more = false

					return q, nil
				},
				'+': func(buf buffer) (quantifier, error) {
					q := quantifier{}

					_, err := skip(buf)
					if err != nil {
						return q, err
					}

					q.from = 1
					q.more = true

					return q, nil
				},
				'*': func(buf buffer) (quantifier, error) {
					q := quantifier{}

					_, err := skip(buf)
					if err != nil {
						return q, err
					}

					q.from = 0
					q.more = true

					return q, nil
				},
				'{': s.Braces(func(buf buffer) (quantifier, error) {
					q := quantifier{}

					from, err := digit(buf)
					if err != nil {
						return q, err
					}

					q.from = from

					x, err := lookup(buf)
					if err != nil {
						return q, nil
					}
					if x != ',' {
						return q, nil
					}
					_, err = skip(buf)
					if err != nil {
						return q, err
					}

					q.more = true

					to, err := digit(buf)
					if err != nil {
						return q, err
					}

					q.to = &to

					return q,  nil
				},
				),
			},
			lookup,
		),
	)

	return func(buf buffer) (node, error) {
		x, err := expression(buf)
		if err != nil {
			return nil, err
		}

		q, err := parseQuantifier(buf)
		if err != nil {
			return x, nil
		}

		q.expression = x
		q.leaf = buf.IsEOF()

		return q, nil
	}
}

func SkipString(data string) c.Combinator[rune, s.Position, struct{}] {
	return func(buffer c.Buffer[rune, s.Position]) (struct{}, error) {
		l := len(data)
		for _, x := range data {
			r, err := buffer.Read(true)
			if err != nil {
				return none, err
			}
			if x != r {
				return none, c.NothingMatched
			}
			l =- 1
		}

		if l != 0 {
			return none, c.NothingMatched
		}

		return none, nil
	}
}

func parseCharacter(except ...rune) parser {
	char := s.NoneOf(except...)

	return func(buf buffer) (node, error) {
		_, err := char(buf)
		if err != nil {
			return nil, err
		}

		x := dot{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseDot() parser {
	parse := s.Eq('.')

	return func(buf buffer) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := dot{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseDigit() parser {
	parse := SkipString("\\d")

	return func(buf buffer) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := digit{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseNonDigit() parser {
	parse := SkipString("\\D")

	return func(buf buffer) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := nonDigit{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseWord() parser {
	parse := SkipString("\\w")

	return func(buf buffer) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := word{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseNonWord() parser {
	parse := SkipString("\\w")

	return func(buf buffer) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := nonWord{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseSpace() parser {
	parse := SkipString("\\s")

	return func(buf buffer) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := space{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseNonSpace() parser {
	parse := SkipString("\\S")

	return func(buf buffer) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := nonSpace{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseStartOfLine() parser {
	parse := SkipString("\\^")

	return func(buf buffer) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := startOfLine{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseEndOfLine() parser {
	parse := SkipString("\\$")

	return func(buf buffer) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := endOfLine{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseStartOfString() parser {
	parse := SkipString("\\A")

	return func(buf buffer) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := startOfString{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseEndOfString() parser {
	parse := SkipString("\\z")

	return func(buf buffer) (node, error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := endOfString{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseGroup(expression expressionParser) parser {
	sep := s.Eq('|')
	union := s.Parens(s.SepBy1(0, expression, sep))

	return func(buf buffer) (node, error) {
		variants, err := union(buf)
		if err != nil {
			return nil, err
		}

		x := group{
			variants: variants,
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseNotCapturedGroup(expression expressionParser) parser {
	sep := s.Eq('|')
	union := s.SepBy1(0, expression, sep)
	before := SkipString("?:")

	return s.Parens(
		func(buf buffer) (node, error) {
			_, err := before(buf)
			if err != nil {
				return nil, err
			}

			variants, err := union(buf)
			if err != nil {
				return nil, err
			}

			x := notCapturedGroup{
				variants: variants,
				leaf: buf.IsEOF(),
			}

			return x, nil
		},
	)
}

func parseNamedGroup(expression expressionParser, except ...rune) parser {
	sep := s.Eq('|')
	union := s.SepBy1(1, expression, sep)
	groupName := s.Angles(
		s.Skip(
			s.Eq('?'),
			s.Many(0, s.NoneOf(append(except, '>')...)),
		),
	)

	return s.Parens(
		func(buf buffer) (node, error) {
			name, err := groupName(buf)
			if err != nil {
				return nil, err
			}

			variants, err := union(buf)
			if err != nil {
				return nil, err
			}

			x := namedGroup{
				name: string(name),
				variants: variants,
				leaf: buf.IsEOF(),
			}

			return x, nil
		},
	)
}

func parseNegativeSet(expression parser) parser {
	parse := s.Squares(
		s.Skip(
			s.Eq('^'),
			s.Some(1, expression),
		),
	)

	return func(buf buffer) (node, error) {
		set, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := negativeSet{
			value: set,
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parsePositiveSet(expression parser) parser {
	parse := s.Squares(s.Some(1, expression))

	return func(buf buffer) (node, error) {
		set, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := positiveSet{
			value: set,
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseRange(except ...rune) parser {
	item := s.NoneOf(except...)
	sep := s.Eq('-')

	return func(buf buffer) (node, error) {
		f, err := item(buf)
		if err != nil {
			return nil, err
		}

		_, err = sep(buf)
		if err != nil {
			return nil, err
		}

		t, err := item(buf)
		if err != nil {
			return nil, err
		}

		x := rangeNode{
			from: f,
			to: t,
		}

		return x, nil
	}
}
