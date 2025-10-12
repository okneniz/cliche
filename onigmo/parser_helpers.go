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

func parseNamedReference(
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parseName := c.Try(parseNameOfNamedReferences(except...))

	return func(
		buf c.Buffer[rune, int],
	) (node.Node, c.Error[int]) {
		pos := buf.Position()

		name, err := parseName(buf)
		if err != nil {
			return nil, c.NewParseError(
				pos,
				"expected named backreferences",
				err,
			)
		}

		return node.NewForNameReference(string(name)), nil
	}
}

func parseBackReference(except ...rune) c.Combinator[rune, int, node.Node] {
	parseIndex := parseIndexedReferences(true, except...)

	return func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
		pos := buf.Position()

		index, err := parseIndex(buf)
		if err != nil {
			return nil, c.NewParseError(
				pos,
				"expected backreference",
				err,
			)
		}

		return node.NodeForReference(int(index)), nil
	}
}

func parseIndexedReferences(
	escaped bool,
	except ...rune,
) c.Combinator[rune, int, int64] {
	digits := []rune("0123456789")

	if len(except) > 0 {
		exceptM := make(map[rune]struct{}, len(except))
		for _, c := range except {
			exceptM[c] = struct{}{}
		}

		for _, c := range digits {
			if _, exists := exceptM[c]; exists {
				// TODO : helper for it
				panic("exceptions include digit " + string(c))
			}
		}
	}

	// is it possible to have back reference more than nine?
	// for example \13 or \99 ?
	parseSeqOfDigits, err := parser.Quantifier(
		"expected indexed backreference",
		1, 2,
		c.OneOf[rune, int](
			"expected sequence of digits",
			digits...,
		),
	)
	if err != nil {
		panic(err.Error()) // TODO : remove panic
	}

	parse := parseSeqOfDigits

	if escaped {
		parse = c.Skip(
			c.Eq[rune, int](
				"expected '\\' as start of backreference",
				'\\',
			),
			parseSeqOfDigits,
		)
	}

	return func(buf c.Buffer[rune, int]) (int64, c.Error[int]) {
		pos := buf.Position()

		runes, err := parse(buf)
		if err != nil {
			return -1, err
		}

		str := strings.ToLower(string(runes))

		index, castErr := strconv.ParseInt(str, 10, 64)
		if castErr != nil {
			return -1, c.NewParseError(
				pos,
				fmt.Sprintf(
					"invalid index for backreference: %s",
					castErr.Error(),
				),
			)
		}

		return index, nil
	}
}

func parseNameOfNamedReferences(
	except ...rune,
) c.Combinator[rune, int, string] {
	parseLeftAngle := c.Eq[rune, int](
		"expected '<' as begining of name",
		'<',
	)

	parseRightAngle := c.Eq[rune, int](
		"expected '>' as ending of name",
		'>',
	)

	except = append(except, '>')

	parseNameOfBackreference := c.Some(
		1,
		"expected sequence of none '>' chars as name",
		c.Try(
			c.NoneOf[rune, int](
				"expected none of '>'",
				except...,
			// '>',
			),
		),
	)

	return func(buf c.Buffer[rune, int]) (string, c.Error[int]) {
		_, err := parseLeftAngle(buf)
		if err != nil {
			return "", err
		}

		name, err := parseNameOfBackreference(buf)
		if err != nil {
			return "", err
		}

		_, err = parseRightAngle(buf)
		if err != nil {
			return "", err
		}

		return string(name), nil
	}
}

func parseHexNumber(from, to int) parser.ParserBuilder[int] {
	return func(_ ...rune) c.Combinator[rune, int, int] {
		// TODO : don't ignore except, check it

		parse, err := parser.Quantifier(
			"expected hex number, for example 12f or 1B",
			from, to,
			c.OneOf[rune, int](
				"expected at least one symbol of hex number, for example '0123456789abcdefABCDEF'",
				[]rune("0123456789abcdefABCDEF")...,
			),
		)
		if err != nil {
			panic(err.Error()) // TODO : remove panic
		}

		return func(buf c.Buffer[rune, int]) (int, c.Error[int]) {
			pos := buf.Position()

			runes, err := parse(buf)
			if err != nil {
				return -1, err
			}

			str := strings.ToLower(string(runes))

			num, castErr := strconv.ParseInt(str, 16, 64)
			if castErr != nil {
				return -1, c.NewParseError(
					pos,
					fmt.Sprintf(
						"invalid hex number: %s",
						castErr.Error(),
					),
				)
			}

			return int(num), nil
		}
	}
}

func parseOctalCharNumber(size int) parser.ParserBuilder[int] {
	parseLeftBraces := c.Eq[rune, int](
		"expected '{' as begining of octal number",
		'{',
	)

	parseRightBraces := c.Eq[rune, int](
		"expected '}' as ending of octal number",
		'}',
	)

	return func(_ ...rune) c.Combinator[rune, int, int] {
		// TODO : don't ignore except

		allowed := []rune("01234567")
		parse := c.Count(
			size,
			"expected at least one digit between 0 and 7 as octal number",
			c.OneOf[rune, int](
				"expecte digit between 0 and 7",
				allowed...,
			),
		)

		return func(buf c.Buffer[rune, int]) (int, c.Error[int]) {
			pos := buf.Position()

			_, leftErr := parseLeftBraces(buf)
			if leftErr != nil {
				return -1, leftErr
			}

			runes, runesErr := parse(buf)
			if runesErr != nil {
				return -1, runesErr
			}

			_, rightErr := parseRightBraces(buf)
			if rightErr != nil {
				return -1, rightErr
			}

			str := strings.ToLower(string(runes))

			num, castErr := strconv.ParseInt(str, 8, 64)
			if castErr != nil {
				return -1, c.NewParseError(
					pos,
					fmt.Sprintf(
						"invalid octal number: %s",
						castErr.Error(),
					),
				)
			}

			return int(num), nil
		}
	}
}

func makeOptionsParser(
	opts map[rune]node.ScanOption,
	except ...rune,
) c.Combinator[rune, int, []node.ScanOption] {
	parseRune := c.NoneOf[rune, int](
		"expected option flag",
		except...,
	)

	flags := make([]string, len(opts))
	for k := range opts {
		flags = append(flags, string(k))
	}

	errMessage := fmt.Sprintf(
		"expected one of option flag:  %v",
		strings.Join(flags, ", "),
	)

	parse := c.Many(
		1,
		c.Try(c.Map(errMessage, opts, parseRune)),
	)

	return parse
}

func parseGroup(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parseCapturedGroup := c.Try(func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
		pos := buf.Position()

		alt, altErr := parseAlternation(buf)
		if altErr != nil {
			return nil, c.NewParseError(
				pos,
				"expected alternation expression in non captured group",
				altErr,
			)
		}

		return node.NewGroup(alt), nil
	})

	prefix := c.Try(c.Eq[rune, int](
		"expected '?' as prefix of options for non captured group",
		'?',
	))

	// TODO : add prefix parseSomething
	comma := c.Try(c.Eq[rune, int](
		"expected ':' as suffix of options for non captured group",
		':',
	))

	optsDict := map[rune]node.ScanOption{
		'i': node.ScanOptionCaseInsensetive,
		'm': node.ScanOptionMultiline,
	}

	parseOptions := c.Try(
		makeOptionsParser(optsDict, append(except, '-', ':')...),
	)

	parseDiableOptions := c.Try(
		c.Skip(
			c.Eq[rune, int](
				"expected '-' as prefix for disabled options",
				'-',
			),
			parseOptions,
		),
	)

	parseNotCapturedGroupWithOptions := c.Try(func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
		pos := buf.Position()

		_, err := prefix(buf)
		if err != nil {
			return nil, err
		}

		enable, enableOptErr := parseOptions(buf)
		disable, disableOptErr := parseDiableOptions(buf)

		if len(enable) == 0 && len(disable) == 0 {
			return nil, c.NewParseError(
				pos,
				"captured group without any options",
				enableOptErr,
				disableOptErr,
			)
		}

		_, err = comma(buf)
		if err != nil {
			return node.NewOptionsSwitcher(enable, disable), nil
		}

		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, err
		}

		group := node.NewGroup(alt)

		// TODO : move it to alterer
		switcher := node.NewOptionsSwitcher(enable, disable)
		switcher.GetNestedNodes()[group.GetKey()] = group
		// TODO : add opposite switcher after?

		return switcher, nil
	})

	return func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
		pos := buf.Position()

		groupWithOptions, withOptsErr := parseNotCapturedGroupWithOptions(buf)
		if withOptsErr == nil {
			return groupWithOptions, nil
		}

		group, groupErr := parseCapturedGroup(buf)
		if groupErr == nil {
			return group, nil
		}

		return nil, c.NewParseError(
			pos,
			"expected not captured group",
			withOptsErr,
			groupErr,
		)
	}
}

func parseNotCapturedGroup(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	_ ...rune,
) c.Combinator[rune, int, node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
		pos := buf.Position()

		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, c.NewParseError(
				pos,
				"expected non captured group",
				err,
			)
		}

		return node.NewNotCapturedGroup(alt), nil
	}
}

func parseNamedGroup(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	except = append(except, '>')

	parseAllowedForNameSymbols := c.NoneOf[rune, int](
		fmt.Sprintf("expected any char exclude %v", except),
		except...,
	)

	parseEndOfName := c.Eq[rune, int](
		"expected '>' as ending of name of named group",
		'>',
	)

	parseGroupName := c.SkipAfter(
		parseEndOfName,
		c.Some(
			10,
			"expected name of group",
			c.Try(parseAllowedForNameSymbols),
		),
	)

	return c.Try(func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
		pos := buf.Position()

		name, err := parseGroupName(buf)
		if err != nil {
			return nil, err
		}

		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, c.NewParseError(
				pos,
				"expected named group",
				err,
			)
		}

		return node.NewNamedGroup(string(name), alt), nil
	})
}

func parseAtomicGroup(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	_ ...rune,
) c.Combinator[rune, int, node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
		pos := buf.Position()

		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, c.NewParseError(
				pos,
				"expected atomic group",
				err,
			)
		}

		return node.NewAtomicGroup(alt), nil
	}
}

func parseLookAhead(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	_ ...rune,
) c.Combinator[rune, int, node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
		pos := buf.Position()

		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, c.NewParseError(
				pos,
				"expected lookahead expression",
				err,
			)
		}

		return node.NewLookAhead(alt), nil
	}
}

func parseNegativeLookAhead(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	_ ...rune,
) c.Combinator[rune, int, node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
		pos := buf.Position()

		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, c.NewParseError(
				pos,
				"expected negative lookahead expression",
				err,
			)
		}

		return node.NewNegativeLookAhead(alt), nil
	}
}

func parseLookBehind(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	_ ...rune,
) c.Combinator[rune, int, node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
		pos := buf.Position()

		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, c.NewParseError(
				pos,
				"expected lookbehind expression",
				err,
			)
		}

		// TODO : move validation to specail step
		n, validationErr := node.NewLookBehind(alt)
		if validationErr != nil {
			return nil, c.NewParseError(
				pos,
				"expected lookbehind expression",
				c.NewParseError(
					pos,
					validationErr.Error(),
				),
			)
		}

		return n, nil
	}
}

func parseNegativeLookBehind(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	_ ...rune,
) c.Combinator[rune, int, node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
		pos := buf.Position()

		alt, err := parseAlternation(buf)
		if err != nil {
			return nil, c.NewParseError(
				pos,
				"negative lookbehind",
				err,
			)
		}

		// TODO : move validation to specail step
		n, validationErr := node.NewNegativeLookBehind(alt)
		if validationErr != nil {
			return nil, c.NewParseError(
				pos,
				"negative lookbehind",
				c.NewParseError(
					pos,
					validationErr.Error(),
				),
			)
		}

		return n, nil
	}
}

func parseCondition(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parseName := c.Try(parseNameOfNamedReferences(except...))
	parseIndex := c.Try(parseIndexedReferences(false, except...))

	parseBackReference := c.Try(func(
		buf c.Buffer[rune, int],
	) (*node.Predicate, c.Error[int]) {
		pos := buf.Position()

		index, err := parseIndex(buf)
		if err != nil {
			return nil, c.NewParseError(
				pos,
				"expected backreference",
				err,
			)
		}

		return node.NewPredicate(
			strconv.Itoa(int(index)),
			func(s node.Scanner) bool {
				_, matched := s.GetGroup(int(index))
				return matched
			},
		), nil
	})

	parseNameReference := c.Try(func(
		buf c.Buffer[rune, int],
	) (*node.Predicate, c.Error[int]) {
		pos := buf.Position()

		name, nameErr := parseName(buf)
		if nameErr != nil {
			return nil, c.NewParseError(
				pos,
				"expected named reference",
				nameErr,
			)
		}

		str := string(name)

		return node.NewPredicate(
			str,
			func(s node.Scanner) bool {
				_, matched := s.GetNamedGroup(str)
				return matched
			},
		), nil

	})

	parseReference := func(
		buf c.Buffer[rune, int],
	) (*node.Predicate, c.Error[int]) {
		pos := buf.Position()

		ref, backErr := parseBackReference(buf)
		if backErr == nil {
			return ref, nil
		}

		ref, nameErr := parseNameReference(buf)
		if nameErr == nil {
			return ref, nil
		}

		return nil, c.NewParseError(
			pos,
			"expected condition expression",
			backErr,
			nameErr,
		)
	}

	parseLeftParens := c.Eq[rune, int](
		"expected '(' as begining of condition",
		'(',
	)

	parseRightParens := c.Eq[rune, int](
		"expected ')' as ending of condition",
		')',
	)

	parsePrefix := c.Eq[rune, int](
		"expected '?' as begining of condition",
		'?',
	)

	return c.Try(func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
		pos := buf.Position()

		_, err := parsePrefix(buf)
		if err != nil {
			return nil, err
		}

		_, err = parseLeftParens(buf)
		if err != nil {
			return nil, err
		}

		cond, err := parseReference(buf)
		if err != nil {
			return nil, err
		}

		_, err = parseRightParens(buf)
		if err != nil {
			return nil, err
		}

		alt, altErr := parseAlternation(buf)
		if altErr != nil {
			return nil, c.NewParseError(
				pos,
				"expected condition branch",
				altErr,
			)
		}

		variants := alt.GetVariants()

		switch len(variants) {
		case 1:
			return node.NewGuard(cond, variants[0]), nil
		case 2:
			return node.NewCondition(cond, variants[0], variants[1]), nil
		default:
			return nil, c.NewParseError(
				pos,
				"invalid condition pattern",
			)
		}
	})
}

func parseQuantity() parser.ParserBuilder[*quantity.Quantity] {
	return func(except ...rune) c.Combinator[rune, int, *quantity.Quantity] {
		number := parseNumber(except...)

		comma := c.Eq[rune, int](
			"expected ',' as separator in quantifier",
			',',
		)

		rightBrace := c.Eq[rune, int](
			"expected '}' as ending of quantifier",
			'}',
		)

		full := c.Try(func(buf c.Buffer[rune, int]) (*quantity.Quantity, c.Error[int]) { // {1,1}
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

			_, err = rightBrace(buf)
			if err != nil {
				return nil, err
			}

			if from > to {
				// TODO : move to validation?
				return nil, c.NewParseError(
					pos,
					"invalid quantifier",
				)
			}

			return quantity.New(from, to), nil
		})

		fromZero := c.Try(func(buf c.Buffer[rune, int]) (*quantity.Quantity, c.Error[int]) { // {,1}
			_, commaErr := comma(buf)
			if commaErr != nil {
				return nil, commaErr
			}

			to, numErr := number(buf)
			if numErr != nil {
				return nil, numErr
			}

			_, braceErr := rightBrace(buf)
			if braceErr != nil {
				return nil, braceErr
			}

			return quantity.New(0, to), nil
		})

		endless := c.Try(func(buf c.Buffer[rune, int]) (*quantity.Quantity, c.Error[int]) { // {1,}
			from, numErr := number(buf)
			if numErr != nil {
				return nil, numErr
			}

			_, commaErr := comma(buf)
			if commaErr != nil {
				return nil, commaErr
			}

			_, braceErr := rightBrace(buf)
			if braceErr != nil {
				return nil, braceErr
			}

			return quantity.NewEndlessQuantity(from), nil
		})

		fixed := c.Try(func(buf c.Buffer[rune, int]) (*quantity.Quantity, c.Error[int]) { // {1}
			from, err := number(buf)
			if err != nil {
				return nil, err
			}

			_, err = rightBrace(buf)
			if err != nil {
				return nil, err
			}

			return quantity.New(from, from), nil
		})

		return func(buf c.Buffer[rune, int]) (*quantity.Quantity, c.Error[int]) {
			pos := buf.Position()

			q, fullErr := full(buf)
			if fullErr == nil {
				return q, nil
			}

			q, fromZeroErr := fromZero(buf)
			if fromZeroErr == nil {
				return q, nil
			}

			q, endlessErr := endless(buf)
			if endlessErr == nil {
				return q, nil
			}

			q, fixedErr := fixed(buf)
			if fixedErr == nil {
				return q, nil
			}

			return nil, c.NewParseError(
				pos,
				"expected quantifier",
				fullErr,
				fromZeroErr,
				endlessErr,
				fixedErr,
			)
		}
	}
}

func parseComment(
	_ c.Combinator[rune, int, node.Alternation],
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parse := c.Many(
		10,
		c.Try(
			c.NoneOf[rune, int](
				"expected comment",
				except...,
			),
		),
	)

	return func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
		runes, err := parse(buf)
		if err != nil {
			return nil, err
		}

		return node.NewComment(string(runes)), nil
	}
}

func parseNumber(_ ...rune) c.Combinator[rune, int, int] {
	const zero = rune('0')

	// TODO : don't ignore except param

	digit := c.Try(c.OneOf[rune, int](
		"expected digit",
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	))

	return func(buf c.Buffer[rune, int]) (int, c.Error[int]) {
		token, err := digit(buf)
		if err != nil {
			return 0, err
		}

		number := int(token - zero)

		for {
			token, err := digit(buf)
			if err != nil {
				break
			}

			number = number * 10
			number += int(token - zero)
		}

		return number, nil
	}
}
