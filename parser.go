package regular

import (
	"errors"
	"fmt"

	c "github.com/okneniz/parsec/common"
)

type parser = c.Combinator[rune, int, node]

var (
	defaultParser = parseRegexp()
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

	// parse union
	union := func(buf c.Buffer[rune, int]) (*union, error) {
		variant, err := parseNestedExpression(buf)
		if err != nil {
			return nil, err
		}

		variants := make([]node, 0, 1)
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

		return newUnion(variants), nil
	}

	// parse node
	parseNode := parseOptionalQuantifier(
		choice(
			parseSet('|'),
			parseNotCapturedGroup(union),
			parseNamedGroup(union),
			parseGroup(union),
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
			parseSet('|', ')'),
			parseNotCapturedGroup(union),
			parseNamedGroup(union),
			parseGroup(union),
			parseInvalidQuantifier(),
			parseEscapedMetaCharacters(),
			parseMetaCharacters(),
			parseEscapedSpecSymbols(),
			parseCharacter('|', ')'),
		),
	)

	parseExpression = func(buf c.Buffer[rune, int]) (node, error) {
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

			last.getNestedNodes()[next.getKey()] = next
			last = next
		}

		return first, nil
	}

	parseNestedExpression = func(buf c.Buffer[rune, int]) (node, error) {
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

			last.getNestedNodes()[next.getKey()] = next
			last = next
		}

		return first, nil
	}

	// parse union or expression
	return func(buf c.Buffer[rune, int]) (node, error) {
		expression, err := parseExpression(buf)
		if err != nil {
			return nil, err
		}
		if buf.IsEOF() {
			return expression, nil
		}

		variants := make([]node, 0, 1)
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

		return newUnion(variants), nil
	}
}

func parseSet(except ...rune) parser {
	// TODO : without except?
	parseNode := choice(
		parseRange(append(except, ']')...),
		parseEscapedMetaCharacters(),
		parseEscapedSpecSymbols(),
		parseCharacter(append(except, ']')...),
	)

	return choice(
		parseNegativeSet(parseNode),
		parsePositiveSet(parseNode),
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

		cases[r] = func(buf c.Buffer[rune, int]) (node, error) {
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

	return func(buf c.Buffer[rune, int]) (node, error) {
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

	parseQuantifier := c.Try(
		c.MapAs(
			map[rune]c.Combinator[rune, int, quantifier]{
				'?': func(buf c.Buffer[rune, int]) (quantifier, error) {
					q := quantifier{}

					q.From = 0
					to := 1
					q.To = &to
					q.More = false

					return q, nil
				},
				'+': func(buf c.Buffer[rune, int]) (quantifier, error) {
					q := quantifier{}

					q.From = 1
					q.More = true

					return q, nil
				},
				'*': func(buf c.Buffer[rune, int]) (quantifier, error) {
					q := quantifier{}

					q.From = 0
					q.More = true

					return q, nil
				},
				'{': func(buf c.Buffer[rune, int]) (quantifier, error) {
					q := quantifier{}

					from, err := digit(buf)
					if err != nil {
						return q, err
					}
					q.From = from

					_, err = comma(buf)
					if err != nil {
						if from == 0 {
							// TODO : or better special parsing error?
							return q, c.NothingMatched
						}

						_, err = rightBrace(buf)
						if err != nil {
							return q, err
						}

						return q, nil
					}
					q.More = true

					to, err := digit(buf)
					if err != nil {
						_, err = rightBrace(buf)
						if err != nil {
							return q, err
						}

						return q, err
					}
					q.To = &to
					q.More = false

					if (from == 0 && to == 0) || (from > to) {
						// TODO : or better special parsing error?
						return q, c.NothingMatched
					}

					if from == to {
						q.To = nil
					}

					_, err = rightBrace(buf)
					if err != nil {
						return q, err
					}

					return q, nil
				},
			},
			any,
		),
	)

	return func(buf c.Buffer[rune, int]) (node, error) {
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

		return &q, nil
	}
}

func parseCharacter(except ...rune) parser {
	parse := c.NoneOf[rune, int](except...)

	return func(buf c.Buffer[rune, int]) (node, error) {
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
	n.Nested = make(index)
	return n
}

func parseMetaCharacters() parser {
	return c.MapAs(
		map[rune]c.Combinator[rune, int, node]{
			'.': func(buf c.Buffer[rune, int]) (node, error) {
				x := dot{
					nestedNode: newNestedNode(),
				}

				return &x, nil
			},
			'^': func(buf c.Buffer[rune, int]) (node, error) {
				x := startOfLine{
					nestedNode: newNestedNode(),
				}

				return &x, nil
			},
			'$': func(buf c.Buffer[rune, int]) (node, error) {
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
			map[rune]c.Combinator[rune, int, node]{
				'd': func(buf c.Buffer[rune, int]) (node, error) {
					x := digit{
						nestedNode: newNestedNode(),
					}

					return &x, nil
				},
				'D': func(buf c.Buffer[rune, int]) (node, error) {
					x := nonDigit{
						nestedNode: newNestedNode(),
					}

					return &x, nil
				},
				'w': func(buf c.Buffer[rune, int]) (node, error) {
					x := word{
						nestedNode: newNestedNode(),
					}

					return &x, nil
				},
				'W': func(buf c.Buffer[rune, int]) (node, error) {
					x := nonWord{
						nestedNode: newNestedNode(),
					}

					return &x, nil
				},
				's': func(buf c.Buffer[rune, int]) (node, error) {
					x := space{
						nestedNode: newNestedNode(),
					}

					return &x, nil
				},
				'S': func(buf c.Buffer[rune, int]) (node, error) {
					x := nonSpace{
						nestedNode: newNestedNode(),
					}

					return &x, nil
				},
				'A': func(buf c.Buffer[rune, int]) (node, error) {
					x := startOfString{
						nestedNode: newNestedNode(),
					}

					return &x, nil
				},
				'z': func(buf c.Buffer[rune, int]) (node, error) {
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

func parseGroup(parse c.Combinator[rune, int, *union]) parser {
	return parens(
		func(buf c.Buffer[rune, int]) (node, error) {
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

func parseNotCapturedGroup(parse c.Combinator[rune, int, *union]) parser {
	before := SkipString("?:")

	return parens(
		func(buf c.Buffer[rune, int]) (node, error) {
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

func parseNamedGroup(parse c.Combinator[rune, int, *union], except ...rune) parser {
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
		func(buf c.Buffer[rune, int]) (node, error) {
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

func parseNegativeSet(expression parser) parser {
	parse := squares(
		c.Skip(
			c.Eq[rune, int]('^'),
			c.Some(1, expression),
		),
	)

	return func(buf c.Buffer[rune, int]) (node, error) {
		set, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := negativeSet{
			Value:      set,
			nestedNode: newNestedNode(),
		}

		return &x, nil
	}
}

func parsePositiveSet(expression parser) parser {
	parse := squares(c.Some(1, expression))

	return func(buf c.Buffer[rune, int]) (node, error) {
		set, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := positiveSet{
			Value:      set,
			nestedNode: newNestedNode(),
		}

		return &x, nil
	}
}

func parseRange(except ...rune) parser {
	item := c.NoneOf[rune, int](except...)
	sep := c.Eq[rune, int]('-')

	return func(buf c.Buffer[rune, int]) (node, error) {
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
			From:       f,
			To:         t,
			nestedNode: newNestedNode(),
		}

		return &x, nil
	}
}


type simpleBuffer struct {
	data     []rune
	position int

	// data string
	// positions []int (stack of last positions)
}

// newBuffer - make buffer which can read text on input

var _ c.Buffer[rune, int] = &simpleBuffer{}

func newBuffer(str string) *simpleBuffer {
	b := new(simpleBuffer)
	b.data = []rune(str)
	b.position = 0
	return b
}

// Read - read next item, if greedy buffer keep position after reading.
func (b *simpleBuffer) Read(greedy bool) (rune, error) {
	if b.IsEOF() {
		return 0, c.EndOfFile
	}

	x := b.data[b.position]

	if greedy {
		b.position++
	}

	return x, nil
}

func (b *simpleBuffer) ReadAt(idx int) rune {
	return b.data[idx]
}

func (b *simpleBuffer) Size() int { // TODO : check for another runes
	return len(b.data)
}

func (b *simpleBuffer) String() string {
	return fmt.Sprintf("Buffer(%s, %d)", string(b.data), b.position)
}

func (b *simpleBuffer) Substring(from, to int) (string, error) {
	if from > to {
		return "", fmt.Errorf(
			"invalid bounds for substring: from=%d to=%d size=%d",
			from,
			to,
			len(b.data),
		)
	}

	if from < 0 || from >= len(b.data) || to >= len(b.data) {
		return "", fmt.Errorf(
			"out of bounds buffer: from=%d to=%d size=%d",
			from,
			to,
			len(b.data),
		)
	}

	return string(b.data[from : to+1]), nil
}

// Seek - change buffer position
func (b *simpleBuffer) Seek(x int) {
	b.position = x
}

// Position - return current buffer position
func (b *simpleBuffer) Position() int {
	return b.position
}

// IsEOF - true if buffer ended
func (b *simpleBuffer) IsEOF() bool {
	return b.position >= len(b.data)
}
