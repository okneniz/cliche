package onigmo

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/parser"
	"github.com/okneniz/cliche/quantity"
	c "github.com/okneniz/parsec/common"
)

// TODO : move onigmo speciic node.Node to this package?
// lookahead / lookbehind
// conditions
// comments

func braces[T any](makeParser parser.ParserBuilder[T]) parser.ParserBuilder[T] {
	return func(except ...rune) c.Combinator[rune, int, T] {
		parse := parser.Braces(makeParser(except...))

		return func(buf c.Buffer[rune, int]) (T, error) {
			x, err := parse(buf)
			if err != nil {
				var def T
				return def, err
			}

			return x, nil
		}
	}
}

func parseNameReference(except ...rune) c.Combinator[rune, int, node.Node] {
	parse := parser.Angles(
		c.Some(
			0,
			c.Try(c.NoneOf[rune, int]('>')),
		),
	)

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		name, err := parse(buf)
		if err != nil {
			return nil, err
		}

		return node.NewForNameReference(string(name)), nil
	}
}

func parseBackReference(except ...rune) c.Combinator[rune, int, node.Node] {
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
	parse := c.Skip(
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

func parseHexNumber(from, to int) parser.ParserBuilder[int] {
	return func(except ...rune) c.Combinator[rune, int, int] {
		// TODO : don't ignore except

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
}

func parseOctal(size int) parser.ParserBuilder[int] {
	return func(except ...rune) c.Combinator[rune, int, int] {
		allowed := []rune("01234567")
		parse := c.Count(size, c.OneOf[rune, int](allowed...))

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
}

func parseGroup(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	_ ...rune,
) c.Combinator[rune, int, node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, err
		}

		return node.NewGroup(alt), nil
	}
}

func parseNotCapturedGroup(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	_ ...rune,
) c.Combinator[rune, int, node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, err
		}

		return node.NewNotCapturedGroup(alt), nil
	}
}

func parseNamedGroup(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	endOfName := c.Eq[rune, int]('>')
	allowedForNamedSymbols := c.NoneOf[rune, int](append(except, '>')...)

	parseGroupName := c.SkipAfter(
		endOfName,
		c.Some(0, c.Try(allowedForNamedSymbols)),
	)

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		name, err := parseGroupName(buf)
		if err != nil {
			return nil, err
		}

		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, err
		}

		return node.NewNamedGroup(string(name), alt), nil
	}
}

func parseAtomicGroup(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	_ ...rune,
) c.Combinator[rune, int, node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, err
		}

		return node.NewAtomicGroup(alt), nil
	}
}

func parseLookAhead(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	_ ...rune,
) c.Combinator[rune, int, node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, err
		}

		return node.NewLookAhead(alt), nil
	}
}

func parseNegativeLookAhead(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	_ ...rune,
) c.Combinator[rune, int, node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, err
		}

		return node.NewNegativeLookAhead(alt), nil
	}
}

func parseLookBehind(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	_ ...rune,
) c.Combinator[rune, int, node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, err
		}

		n, err := node.NewLookBehind(alt)
		if err != nil {
			// TODO : return explanation from parser
			// handle not only NothingMatched error
			panic(err)
		}

		return n, nil
	}
}

func parseNegativeLookBehind(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	_ ...rune,
) c.Combinator[rune, int, node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, err
		}

		n, err := node.NewNegativeLookBehind(alt)
		if err != nil {
			// TODO : return explanation from parser
			// handle not only NothingMatched error
			panic(err)
		}

		return n, nil
	}
}

// (?('test')c|d)
func parseCondition(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	_ ...rune,
) c.Combinator[rune, int, node.Node] {
	// TODO : don't ignore except

	digits := []rune("0123456789")
	backReference := parser.Quantifier(1, 2, c.OneOf[rune, int](digits...))
	nameReference := parser.Angles(c.Some(0, c.Try(c.NoneOf[rune, int]('>'))))

	parseBackReference := func(buf c.Buffer[rune, int]) (*node.Predicate, error) {
		runes, err := backReference(buf)
		if err != nil {
			return nil, err
		}

		str := strings.ToLower(string(runes))

		index, err := strconv.ParseInt(str, 16, 64)
		if err != nil {
			return nil, err
		}

		return node.NewPredicate(
			fmt.Sprintf("%d", index), // TODO: use strconv instead
			func(s node.Scanner) bool {
				_, matched := s.GetGroup(int(index))
				return matched
			},
		), nil
	}

	parseNameReference := func(buf c.Buffer[rune, int]) (*node.Predicate, error) {
		name, err := nameReference(buf)
		if err != nil {
			return nil, err
		}

		str := string(name)

		return node.NewPredicate(
			str,
			func(s node.Scanner) bool {
				_, matched := s.GetNamedGroup(str)
				return matched
			},
		), nil

	}

	reference := func(buf c.Buffer[rune, int]) (*node.Predicate, error) {
		pos := buf.Position()

		ref, err := parseBackReference(buf)
		if err == nil {
			return ref, nil
		}

		buf.Seek(pos)

		ref, err = parseNameReference(buf)
		if err == nil {
			return ref, nil
		}

		return nil, err
	}

	condition := parser.Parens(reference)
	before := c.Eq[rune, int]('?')

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		_, err := before(buf)
		if err != nil {
			return nil, err
		}

		cond, err := condition(buf)
		if err != nil {
			return nil, err
		}

		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, err
		}

		variants := alt.GetVariants()

		switch len(variants) {
		case 1:
			return node.NewGuard(cond, variants[0]), nil
		case 2:
			return node.NewCondition(cond, variants[0], variants[1]), nil
		}

		return nil, errors.New("invalid condition pattern")
	}
}

func parseQuanty(
	_ ...rune,
) c.Combinator[rune, int, *quantity.Quantity] {
	digit := c.Try(parseNumber())
	comma := c.Try(c.Eq[rune, int](','))
	rightBrace := c.Eq[rune, int]('}')

	return c.Choice(
		c.Try(func(buf c.Buffer[rune, int]) (*quantity.Quantity, error) { // {1,1}
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

			return quantity.NewQuantity(from, to), nil
		}),
		c.Try(func(buf c.Buffer[rune, int]) (*quantity.Quantity, error) { // {,1}
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

			return quantity.NewQuantity(0, to), nil
		}),
		c.Try(func(buf c.Buffer[rune, int]) (*quantity.Quantity, error) { // {1,}
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

			return quantity.NewEndlessQuantity(from), nil
		}),
		func(buf c.Buffer[rune, int]) (*quantity.Quantity, error) { // {1}
			from, err := digit(buf)
			if err != nil {
				return nil, err
			}

			_, err = rightBrace(buf)
			if err != nil {
				return nil, err
			}

			return quantity.NewQuantity(from, from), nil
		},
	)
}

func parseComment(
	_ c.Combinator[rune, int, node.Alternation],
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parse := c.Many(10, c.Try(c.NoneOf[rune, int](except...)))

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		runes, err := parse(buf)
		if err != nil {
			return nil, err
		}

		return node.NewComment(string(runes)), nil
	}
}

func parseNumber() c.Combinator[rune, int, int] {
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
