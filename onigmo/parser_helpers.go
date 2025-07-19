package onigmo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/parser"
	"github.com/okneniz/cliche/quantity"
	c "github.com/okneniz/parsec/common"
)

// TODO : move onigmo specific node.Node to this package?
// lookahead / lookbehind
// conditions
// comments

func parseNameReference(
	except ...rune,
) parser.Parser[node.Node] {
	parse := parser.Angles(
		parser.Some(
			"named backreference",
			parser.Try(parser.NoneOf('>')),
		),
	)

	return func(
		buf c.Buffer[rune, int],
	) (node.Node, *parser.ParsingError) {
		name, err := parse(buf)
		if err != nil {
			return nil, err
		}

		return node.NewForNameReference(string(name)), nil
	}
}

func parseBackReference(except ...rune) parser.Parser[node.Node] {
	digits := []rune("0123456789")

	if len(except) > 0 {
		exceptM := make(map[rune]struct{}, len(except))
		for _, c := range except {
			exceptM[c] = struct{}{}
		}

		for _, c := range digits {
			if _, exists := exceptM[c]; exists {
				// TODO : helper for it?
				panic("exceptions include digit " + string(c))
			}
		}
	}

	// is it possible to have back reference more than nine?
	// for example \13 or \99 ?
	parse := parser.Skip(
		parser.Eq('\\'),
		parser.Quantifier(1, 2, parser.OneOf(digits...)),
	)

	return func(
		buf c.Buffer[rune, int],
	) (node.Node, *parser.ParsingError) {
		pos := buf.Position()

		runes, err := parse(buf)
		if err != nil {
			return nil, parser.Expected("backreference", pos, err)
		}

		str := strings.ToLower(string(runes))

		index, castError := strconv.ParseInt(str, 16, 64)
		if castError != nil {
			return nil, parser.Expected("backreference", pos, castError)
		}

		return node.NodeForReference(int(index)), nil
	}
}

func parseHexNumber(from, to int) parser.ParserBuilder[int] {
	return func(_ ...rune) parser.Parser[int] {
		// TODO : don't ignore except

		parse := parser.Quantifier(
			from,
			to,
			parser.OneOf([]rune("0123456789abcdefABCDEF")...),
		)

		return func(buf c.Buffer[rune, int]) (int, *parser.ParsingError) {
			pos := buf.Position()

			runes, err := parse(buf)
			if err != nil {
				return -1, parser.Expected("hex number", pos, err)
			}

			str := strings.ToLower(string(runes))

			num, castErr := strconv.ParseInt(str, 16, 64)
			if err != nil {
				return -1, parser.Expected("hex number", pos, castErr)
			}

			return int(num), nil
		}
	}
}

func parseOctalCharNumber(size int) parser.ParserBuilder[int] {
	leftBraces := c.Eq[rune, int]('{')
	rightBraces := c.Eq[rune, int]('}')

	return func(_ ...rune) parser.Parser[int] {
		// TODO : don't ignore except

		allowed := []rune("01234567")
		parse := c.Count(size, c.OneOf[rune, int](allowed...))

		return func(buf c.Buffer[rune, int]) (int, *parser.ParsingError) {
			pos := buf.Position()

			_, err := leftBraces(buf)
			if err != nil {
				return -1, parser.Expected("octal number", pos, err)
			}

			runes, err := parse(buf)
			if err != nil {
				return -1, parser.Expected("octal number", pos, err)
			}

			_, err = rightBraces(buf)
			if err != nil {
				return -1, parser.Expected("octal number", pos, err)
			}

			str := strings.ToLower(string(runes))

			num, castErr := strconv.ParseInt(str, 8, 64)
			if err != nil {
				return -1, parser.Expected("octal number", pos, castErr)
			}

			return int(num), nil
		}
	}
}

func parseGroup(
	parseAlternation parser.Parser[node.Alternation],
	_ ...rune,
) parser.Parser[node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, *parser.ParsingError) {
		pos := buf.Position()

		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, parser.Expected("group", pos, err)
		}

		return node.NewGroup(alt), nil
	}
}

func parseNotCapturedGroup(
	parseAlternation parser.Parser[node.Alternation],
	_ ...rune,
) parser.Parser[node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, *parser.ParsingError) {
		pos := buf.Position()

		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, parser.Expected("group", pos, err)
		}

		return node.NewNotCapturedGroup(alt), nil
	}
}

func parseNamedGroup(
	parseAlternation parser.Parser[node.Alternation],
	except ...rune,
) parser.Parser[node.Node] {
	endOfName := c.Eq[rune, int]('>')
	allowedForNamedSymbols := c.NoneOf[rune, int](append(except, '>')...)

	parseGroupName := c.SkipAfter(
		endOfName,
		c.Some(0, c.Try(allowedForNamedSymbols)),
	)

	return func(buf c.Buffer[rune, int]) (node.Node, *parser.ParsingError) {
		pos := buf.Position()

		name, err := parseGroupName(buf)
		if err != nil {
			return nil, parser.Expected("name of group", pos, err)
		}

		fmt.Println("name of group parsed", name, buf)

		pos = buf.Position()

		alt, altErr := parseAlternation(buf)
		if altErr != nil {
			fmt.Println("parsing alternation for named group failed:", altErr)
			return nil, parser.Expected("named group", pos, altErr)
		}

		return node.NewNamedGroup(string(name), alt), nil
	}
}

func parseAtomicGroup(
	parseAlternation parser.Parser[node.Alternation],
	_ ...rune,
) parser.Parser[node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, *parser.ParsingError) {
		pos := buf.Position()

		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, parser.Expected("atomic group", pos, err)
		}

		return node.NewAtomicGroup(alt), nil
	}
}

func parseLookAhead(
	parseAlternation parser.Parser[node.Alternation],
	_ ...rune,
) parser.Parser[node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, *parser.ParsingError) {
		pos := buf.Position()

		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, parser.Expected("lookahead", pos, err)
		}

		return node.NewLookAhead(alt), nil
	}
}

func parseNegativeLookAhead(
	parseAlternation parser.Parser[node.Alternation],
	_ ...rune,
) parser.Parser[node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, *parser.ParsingError) {
		pos := buf.Position()

		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, parser.Expected("negative lookahead", pos, err)
		}

		return node.NewNegativeLookAhead(alt), nil
	}
}

func parseLookBehind(
	parseAlternation parser.Parser[node.Alternation],
	_ ...rune,
) parser.Parser[node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, *parser.ParsingError) {
		pos := buf.Position()

		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, parser.Expected("lookbehind", pos, err)
		}

		n, validationErr := node.NewLookBehind(alt)
		if err != nil {
			return nil, parser.Expected("lookbehind", pos, validationErr)
		}

		return n, nil
	}
}

func parseNegativeLookBehind(
	parseAlternation parser.Parser[node.Alternation],
	_ ...rune,
) parser.Parser[node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, *parser.ParsingError) {
		pos := buf.Position()

		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, parser.Expected("negative lookbehind", pos, err)
		}

		n, validationErr := node.NewNegativeLookBehind(alt)
		if err != nil {
			return nil, parser.Expected("negative lookbehind", pos, validationErr)
		}

		return n, nil
	}
}

// (?('test')c|d)
func parseCondition(
	parseAlternation parser.Parser[node.Alternation],
	_ ...rune,
) parser.Parser[node.Node] {
	// TODO : don't ignore except

	digits := []rune("0123456789")
	backReference := parser.Quantifier(1, 2, parser.OneOf(digits...))
	nameReference := parser.Angles(
		parser.Some(
			"backreference name",
			parser.Try(parser.NoneOf('>')),
		),
	)

	parseBackReference := func(
		buf c.Buffer[rune, int],
	) (*node.Predicate, *parser.ParsingError) {
		pos := buf.Position()

		runes, err := backReference(buf)
		if err != nil {
			return nil, err
		}

		str := strings.ToLower(string(runes))

		index, castErr := strconv.ParseInt(str, 16, 64)
		if err != nil {
			return nil, parser.Expected("digit", pos, castErr)
		}

		return node.NewPredicate(
			fmt.Sprintf("%d", index), // TODO: use strconv instead
			func(s node.Scanner) bool {
				_, matched := s.GetGroup(int(index))
				return matched
			},
		), nil
	}

	parseNameReference := func(
		buf c.Buffer[rune, int],
	) (*node.Predicate, *parser.ParsingError) {
		pos := buf.Position()

		name, err := nameReference(buf)
		if err != nil {
			return nil, parser.Expected("named reference", pos, err)
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

	reference := func(
		buf c.Buffer[rune, int],
	) (*node.Predicate, *parser.ParsingError) {
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

		return nil, parser.Expected("condition backreferences", pos, err)
	}

	condition := parser.Parens(reference)
	before := parser.Eq('?')

	return func(buf c.Buffer[rune, int]) (node.Node, *parser.ParsingError) {
		pos := buf.Position()

		_, err := before(buf)
		if err != nil {
			return nil, parser.Expected("condition", pos, err)
		}

		cond, condErr := condition(buf)
		if err != nil {
			return nil, parser.Expected("condition branch", pos, condErr)
		}

		alt, altErr := parseAlternation(buf)
		if err != nil {
			return nil, parser.Expected("condition branch", pos, altErr)
		}

		variants := alt.GetVariants()

		switch len(variants) {
		case 1:
			return node.NewGuard(cond, variants[0]), nil
		case 2:
			return node.NewCondition(cond, variants[0], variants[1]), nil
		default:
			return nil, parser.Expected(
				"condition",
				pos,
				fmt.Errorf("invalid condition pattern"),
			)
		}
	}
}

func parseQuantity() parser.ParserBuilder[*quantity.Quantity] {
	return func(except ...rune) parser.Parser[*quantity.Quantity] {
		number := parseNumber(except...)
		comma := parser.Eq(',')
		rightBrace := parser.Eq('}')

		full := func(buf c.Buffer[rune, int]) (*quantity.Quantity, *parser.ParsingError) { // {1,1}
			pos := buf.Position()

			from, err := number(buf)
			if err != nil {
				return nil, err
			}

			_, err = comma(buf)
			if err != nil {
				return nil, err
			}

			to, err := number(buf)
			if err != nil {
				return nil, err
			}

			if from > to {
				// TODO : move out of parsing, to validation?
				return nil, parser.Expected("quantity", pos, fmt.Errorf("invalid bounds"))
			}

			_, err = rightBrace(buf)
			if err != nil {
				return nil, err
			}

			return quantity.New(from, to), nil
		}

		fromZero := func(buf c.Buffer[rune, int]) (*quantity.Quantity, *parser.ParsingError) { // {,1}
			_, err := comma(buf)
			if err != nil {
				return nil, err
			}

			to, err := number(buf)
			if err != nil {
				return nil, err
			}

			_, err = rightBrace(buf)
			if err != nil {
				return nil, err
			}

			return quantity.New(0, to), nil
		}

		endless := func(buf c.Buffer[rune, int]) (*quantity.Quantity, *parser.ParsingError) { // {1,}
			from, err := number(buf)
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
		}

		fixed := func(buf c.Buffer[rune, int]) (*quantity.Quantity, *parser.ParsingError) { // {1}
			fmt.Println("parse fixed quantity:", buf)

			from, err := number(buf)
			if err != nil {
				return nil, err
			}

			fmt.Println("parsed number:", from)

			_, err = rightBrace(buf)
			if err != nil {
				fmt.Println("right brace?", err)
				return nil, err
			}

			return quantity.New(from, from), nil
		}

		return func(buf c.Buffer[rune, int]) (*quantity.Quantity, *parser.ParsingError) {
			pos := buf.Position()

			q, fullErr := full(buf)
			if fullErr == nil {
				return q, nil
			}

			buf.Seek(pos)

			q, fromZeroErr := fromZero(buf)
			if fromZeroErr == nil {
				return q, nil
			}

			buf.Seek(pos)

			q, endlessErr := endless(buf)
			if endlessErr == nil {
				return q, nil
			}

			buf.Seek(pos)

			q, fixedErr := fixed(buf)
			if fixedErr == nil {
				return q, nil
			}

			buf.Seek(pos)

			return nil, parser.MergeErrors(
				fullErr,
				fromZeroErr,
				endlessErr,
				fixedErr,
			)
		}
	}
}

func parseComment(
	_ parser.Parser[node.Alternation],
	except ...rune,
) parser.Parser[node.Node] {
	parse := c.Many(10, c.Try(c.NoneOf[rune, int](except...)))

	return func(buf c.Buffer[rune, int]) (node.Node, *parser.ParsingError) {
		pos := buf.Position()

		runes, err := parse(buf)
		if err != nil {
			return nil, parser.Expected("comment", pos, err)
		}

		return node.NewComment(string(runes)), nil
	}
}

func parseNumber(_ ...rune) parser.Parser[int] {
	const zero = rune('0')

	digit := parser.OneOf('0', '1', '2', '3', '4', '5', '6', '7', '8', '9')

	return func(buf c.Buffer[rune, int]) (int, *parser.ParsingError) {
		pos := buf.Position()

		token, err := digit(buf)
		if err != nil {
			buf.Seek(pos)
			return 0, parser.Expected("digit", pos, err)
		}

		number := int(token - zero)

		for {
			pos = buf.Position()

			token, err = digit(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			number = number * 10
			number += int(token - zero)
		}

		return number, nil
	}
}
