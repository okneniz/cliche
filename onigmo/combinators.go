package onigmo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/parser"
	"github.com/okneniz/cliche/quantity"
	"github.com/okneniz/parsec"
)

// TODO : move onigmo specific node.Node to this package?
// lookahead / lookbehind
// conditions
// comments

func parseNamedReference(
	except ...rune,
) parsec.Combinator[rune, int, node.Node, parser.State] {
	parseName := parsec.Try(parseNameOfNamedReferences(except...))

	return func(
		state parser.State,
		buf parsec.Buffer[rune, int],
	) (node.Node, parsec.Error[int]) {
		pos := buf.Position()

		name, err := parseName(state, buf)
		if err != nil {
			return nil, parsec.NewParseError(
				pos,
				"expected named backreferences",
				err,
			)
		}

		return node.NewForNameReference(name), nil
	}
}

func parseBackReference(except ...rune) parsec.Combinator[rune, int, node.Node, parser.State] {
	parseIndex := parseIndexedReferences(true, except...)

	return func(state parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
		pos := buf.Position()

		index, err := parseIndex(state, buf)
		if err != nil {
			return nil, parsec.NewParseError(
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
) parsec.Combinator[rune, int, int64, parser.State] {
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
	parseSeqOfDigits, err := parsec.Quantifier(
		"expected indexed backreference",
		1, 2,
		parsec.OneOf[rune, int, parser.State](
			"expected sequence of digits",
			digits...,
		),
	)
	if err != nil {
		panic(err.Error()) // TODO : remove panic
	}

	parse := parseSeqOfDigits

	if escaped {
		parse = parsec.Skip(
			parsec.Eq[rune, int, parser.State](
				"expected '\\' as start of backreference",
				'\\',
			),
			parseSeqOfDigits,
		)
	}

	return func(state parser.State, buf parsec.Buffer[rune, int]) (int64, parsec.Error[int]) {
		pos := buf.Position()

		runes, err := parse(state, buf)
		if err != nil {
			return -1, err
		}

		str := strings.ToLower(string(runes))

		index, castErr := strconv.ParseInt(str, 10, 64)
		if castErr != nil {
			return -1, parsec.NewParseError(
				pos,
				"invalid index for backreference: "+castErr.Error(),
			)
		}

		return index, nil
	}
}

func parseNameOfNamedReferences(
	except ...rune,
) parsec.Combinator[rune, int, string, parser.State] {
	parseLeftAngle := parsec.Eq[rune, int, parser.State](
		"expected '<' as beginning of name",
		'<',
	)

	parseRightAngle := parsec.Eq[rune, int, parser.State](
		"expected '>' as ending of name",
		'>',
	)

	except = append(except, '>')

	parseNameOfBackreference := parsec.Some(
		1,
		"expected sequence of none '>' chars as name",
		parsec.Try(
			parsec.NoneOf[rune, int, parser.State](
				"expected none of '>'",
				except...,
			// '>',
			),
		),
	)

	return func(state parser.State, buf parsec.Buffer[rune, int]) (string, parsec.Error[int]) {
		_, err := parseLeftAngle(state, buf)
		if err != nil {
			return "", err
		}

		name, err := parseNameOfBackreference(state, buf)
		if err != nil {
			return "", err
		}

		_, err = parseRightAngle(state, buf)
		if err != nil {
			return "", err
		}

		return string(name), nil
	}
}

func parseHexNumber(from, to int) parser.ParserBuilder[int] {
	return func(_ ...rune) parsec.Combinator[rune, int, int, parser.State] {
		// TODO : don't ignore except, check it

		parseDigits, err := parsec.Quantifier(
			"expected hex number, for example 12f or 1B",
			from, to,
			parsec.OneOf[rune, int, parser.State](
				"expected at least one symbol of hex number, for example '0123456789abcdefABCDEF'",
				[]rune("0123456789abcdefABCDEF")...,
			),
		)
		if err != nil {
			panic(err.Error()) // TODO : remove panic
		}

		return parsec.ParseInt[rune, int, int](
			"expected hex number, for example 12f or 1B",
			16, 64,
			func(state parser.State, buf parsec.Buffer[rune, int]) (string, parsec.Error[int]) {
				digits, err := parseDigits(state, buf)
				if err != nil {
					return "", err
				}

				return strings.ToLower(string(digits)), nil
			},
		)
	}
}

func parseOctalCharNumber(size int) parser.ParserBuilder[int] {
	parseLeftBraces := parsec.Eq[rune, int, parser.State](
		"expected '{' as beginning of octal number",
		'{',
	)

	parseRightBraces := parsec.Eq[rune, int, parser.State](
		"expected '}' as ending of octal number",
		'}',
	)

	return func(_ ...rune) parsec.Combinator[rune, int, int, parser.State] {
		// TODO : don't ignore except

		allowed := []rune("01234567")
		parseDigits := parsec.Count(
			size,
			"expected at least one digit between 0 and 7 as octal number",
			parsec.OneOf[rune, int, parser.State](
				"expecte digit between 0 and 7",
				allowed...,
			),
		)

		parse := parsec.ParseInt[rune, int, int](
			"expected at least one digit between 0 and 7 as octal number",
			8, 64,
			func(state parser.State, buf parsec.Buffer[rune, int]) (string, parsec.Error[int]) {
				digits, err := parseDigits(state, buf)
				if err != nil {
					return "", err
				}

				return strings.ToLower(string(digits)), nil
			},
		)

		return func(state parser.State, buf parsec.Buffer[rune, int]) (int, parsec.Error[int]) {
			_, leftErr := parseLeftBraces(state, buf)
			if leftErr != nil {
				return -1, leftErr
			}

			num, runesErr := parse(state, buf)
			if runesErr != nil {
				return -1, runesErr
			}

			_, rightErr := parseRightBraces(state, buf)
			if rightErr != nil {
				return -1, rightErr
			}

			return num, nil
		}
	}
}

func makeOptionsParser(
	opts map[rune]node.ScanOption,
	except ...rune,
) parsec.Combinator[rune, int, []node.ScanOption, parser.State] {
	parseRune := parsec.NoneOf[rune, int, parser.State](
		"expected option flag",
		except...,
	)

	flags := make([]string, 0, len(opts))
	for k := range opts {
		flags = append(flags, string(k))
	}

	errMessage := fmt.Sprintf(
		"expected one of option flag:  %v",
		strings.Join(flags, ", "),
	)

	parse := parsec.Many(
		1,
		parsec.Try(parsec.Map(errMessage, opts, parseRune)),
	)

	return parse
}

// в parsec нужно добавлять ничего и менять buffer
// чтобы добавить state в комбинатор нужно прокинуть его
// в combinator builder как параметр

func parseGroup(
	parseAlternation parsec.Combinator[rune, int, node.Alternation, parser.State],
	except ...rune,
) parsec.Combinator[rune, int, node.Node, parser.State] {
	parseCapturedGroup := func(
		state parser.State,
		buf parsec.Buffer[rune, int],
	) (node.Node, parsec.Error[int]) {
		pos := buf.Position()

		alt, altErr := parseAlternation(state, buf)
		if altErr != nil {
			return nil, parsec.NewParseError(
				pos,
				"expected alternation expression in non captured group",
				altErr,
			)
		}

		return node.NewGroup(alt, state.GroupIndex), nil
	}

	parseCapturedGroup = parsec.Try(parseCapturedGroup)

	prefix := parsec.Try(parsec.Eq[rune, int, parser.State](
		"expected '?' as prefix of options for non captured group",
		'?',
	))

	// TODO : add prefix parseSomething
	comma := parsec.Try(parsec.Eq[rune, int, parser.State](
		"expected ':' as suffix of options for non captured group",
		':',
	))

	optsDict := map[rune]node.ScanOption{
		'i': node.ScanOptionCaseInsensitive,
		'm': node.ScanOptionMultiline,
	}

	parseOptions := parsec.Try(
		makeOptionsParser(optsDict, append(except, '-', ':')...),
	)

	parseDiableOptions := parsec.Try(
		parsec.Skip(
			parsec.Eq[rune, int, parser.State](
				"expected '-' as prefix for disabled options",
				'-',
			),
			parseOptions,
		),
	)

	parseNotCapturedGroupWithOptions := parsec.Try(
		func(state parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
			pos := buf.Position()

			_, err := prefix(state, buf)
			if err != nil {
				return nil, err
			}

			enable, enableOptErr := parseOptions(state, buf)
			disable, disableOptErr := parseDiableOptions(state, buf)

			if len(enable) == 0 && len(disable) == 0 {
				return nil, parsec.NewParseError(
					pos,
					"captured group without any options",
					enableOptErr,
					disableOptErr,
				)
			}

			_, err = comma(state, buf)
			if err != nil {
				return node.NewOptionsSwitcher(enable, disable), nil
			}

			alt, err := parseAlternation(state, buf)
			if err != nil {
				return nil, err
			}

			// TODO : move it to unify function
			//
			// (?i:foo) -> (?i)(?:foo)(?-i)

			beforeSwitcher := node.NewOptionsSwitcher(enable, disable)
			group := node.NewGroup(alt, state.GroupIndex)
			afterSwitcher := node.NewOptionsSwitcher(disable, enable)

			beforeSwitcher.GetNestedNodes()[group.GetKey()] = group
			group.GetNestedNodes()[afterSwitcher.GetKey()] = afterSwitcher

			return beforeSwitcher, nil
		},
	)

	return func(state parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
		pos := buf.Position()

		groupWithOptions, withOptsErr := parseNotCapturedGroupWithOptions(state, buf)
		if withOptsErr == nil {
			return groupWithOptions, nil
		}

		group, groupErr := parseCapturedGroup(state, buf)
		if groupErr == nil {
			return group, nil
		}

		return nil, parsec.NewParseError(
			pos,
			"expected not captured group",
			withOptsErr,
			groupErr,
		)
	}
}

func parseNotCapturedGroup(
	parseAlternation parsec.Combinator[rune, int, node.Alternation, parser.State],
	_ ...rune,
) parsec.Combinator[rune, int, node.Node, parser.State] {
	return func(state parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
		pos := buf.Position()

		alt, err := parseAlternation(state, buf)
		if err != nil {
			return nil, parsec.NewParseError(
				pos,
				"expected non captured group",
				err,
			)
		}

		return node.NewNotCapturedGroup(alt), nil
	}
}

func parseNamedGroup(
	parseAlternation parsec.Combinator[rune, int, node.Alternation, parser.State],
	except ...rune,
) parsec.Combinator[rune, int, node.Node, parser.State] {
	except = append(except, '>')

	parseAllowedForNameSymbols := parsec.NoneOf[rune, int, parser.State](
		fmt.Sprintf("expected any char exclude %v", except),
		except...,
	)

	parseEndOfName := parsec.Eq[rune, int, parser.State](
		"expected '>' as ending of name of named group",
		'>',
	)

	parseGroupName := parsec.SkipAfter(
		parseEndOfName,
		parsec.Some(
			10,
			"expected name of group",
			parsec.Try(parseAllowedForNameSymbols),
		),
	)

	return parsec.Try(func(state parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
		pos := buf.Position()

		name, err := parseGroupName(state, buf)
		if err != nil {
			return nil, err
		}

		alt, err := parseAlternation(state, buf)
		if err != nil {
			return nil, parsec.NewParseError(
				pos,
				"expected named group",
				err,
			)
		}

		return node.NewNamedGroup(string(name), alt), nil
	})
}

func parseAtomicGroup(
	parseAlternation parsec.Combinator[rune, int, node.Alternation, parser.State],
	_ ...rune,
) parsec.Combinator[rune, int, node.Node, parser.State] {
	return func(state parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
		pos := buf.Position()

		alt, err := parseAlternation(state, buf)
		if err != nil {
			return nil, parsec.NewParseError(
				pos,
				"expected atomic group",
				err,
			)
		}

		return node.NewAtomicGroup(alt), nil
	}
}

func parseLookAhead(
	parseAlternation parsec.Combinator[rune, int, node.Alternation, parser.State],
	_ ...rune,
) parsec.Combinator[rune, int, node.Node, parser.State] {
	return func(state parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
		pos := buf.Position()

		alt, err := parseAlternation(state, buf)
		if err != nil {
			return nil, parsec.NewParseError(
				pos,
				"expected lookahead expression",
				err,
			)
		}

		return node.NewLookAhead(alt), nil
	}
}

func parseNegativeLookAhead(
	parseAlternation parsec.Combinator[rune, int, node.Alternation, parser.State],
	_ ...rune,
) parsec.Combinator[rune, int, node.Node, parser.State] {
	return func(state parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
		pos := buf.Position()

		alt, err := parseAlternation(state, buf)
		if err != nil {
			return nil, parsec.NewParseError(
				pos,
				"expected negative lookahead expression",
				err,
			)
		}

		return node.NewNegativeLookAhead(alt), nil
	}
}

func parseLookBehind(
	parseAlternation parsec.Combinator[rune, int, node.Alternation, parser.State],
	_ ...rune,
) parsec.Combinator[rune, int, node.Node, parser.State] {
	return func(state parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
		pos := buf.Position()

		alt, err := parseAlternation(state, buf)
		if err != nil {
			return nil, parsec.NewParseError(
				pos,
				"expected lookbehind expression",
				err,
			)
		}

		// TODO : move validation to special step
		n, validationErr := node.NewLookBehind(alt)
		if validationErr != nil {
			return nil, parsec.NewParseError(
				pos,
				"expected lookbehind expression",
				parsec.NewParseError(
					pos,
					validationErr.Error(),
				),
			)
		}

		return n, nil
	}
}

func parseNegativeLookBehind(
	parseAlternation parsec.Combinator[rune, int, node.Alternation, parser.State],
	_ ...rune,
) parsec.Combinator[rune, int, node.Node, parser.State] {
	return func(state parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
		pos := buf.Position()

		alt, err := parseAlternation(state, buf)
		if err != nil {
			return nil, parsec.NewParseError(
				pos,
				"negative lookbehind",
				err,
			)
		}

		// TODO : move validation to special step
		n, validationErr := node.NewNegativeLookBehind(alt)
		if validationErr != nil {
			return nil, parsec.NewParseError(
				pos,
				"negative lookbehind",
				parsec.NewParseError(
					pos,
					validationErr.Error(),
				),
			)
		}

		return n, nil
	}
}

func parseCondition(
	parseAlternation parsec.Combinator[rune, int, node.Alternation, parser.State],
	except ...rune,
) parsec.Combinator[rune, int, node.Node, parser.State] {
	parseName := parsec.Try(parseNameOfNamedReferences(except...))
	parseIndex := parsec.Try(parseIndexedReferences(false, except...))

	parseBackReference := parsec.Try(func(
		state parser.State,
		buf parsec.Buffer[rune, int],
	) (*node.Predicate, parsec.Error[int]) {
		pos := buf.Position()

		index, err := parseIndex(state, buf)
		if err != nil {
			return nil, parsec.NewParseError(
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

	parseNameReference := parsec.Try(func(
		state parser.State,
		buf parsec.Buffer[rune, int],
	) (*node.Predicate, parsec.Error[int]) {
		pos := buf.Position()

		name, nameErr := parseName(state, buf)
		if nameErr != nil {
			return nil, parsec.NewParseError(
				pos,
				"expected named reference",
				nameErr,
			)
		}

		return node.NewPredicate(
			name,
			func(s node.Scanner) bool {
				_, matched := s.GetNamedGroup(name)
				return matched
			},
		), nil
	})

	parseReference := func(
		state parser.State,
		buf parsec.Buffer[rune, int],
	) (*node.Predicate, parsec.Error[int]) {
		pos := buf.Position()

		ref, backErr := parseBackReference(state, buf)
		if backErr == nil {
			return ref, nil
		}

		ref, nameErr := parseNameReference(state, buf)
		if nameErr == nil {
			return ref, nil
		}

		return nil, parsec.NewParseError(
			pos,
			"expected condition expression",
			backErr,
			nameErr,
		)
	}

	parseRightParens := parsec.Eq[rune, int, parser.State](
		"expected ')' as ending of condition",
		')',
	)

	return parsec.Try(func(state parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
		pos := buf.Position()

		cond, err := parseReference(state, buf)
		if err != nil {
			return nil, err
		}

		_, err = parseRightParens(state, buf)
		if err != nil {
			return nil, err
		}

		alt, altErr := parseAlternation(state, buf)
		if altErr != nil {
			return nil, parsec.NewParseError(
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
			return nil, parsec.NewParseError(
				pos,
				"invalid condition pattern",
			)
		}
	})
}

func parseQuantity() parser.ParserBuilder[*quantity.Quantity] {
	return func(except ...rune) parsec.Combinator[rune, int, *quantity.Quantity, parser.State] {
		number := parseNumber(except...)

		comma := parsec.Eq[rune, int, parser.State](
			"expected ',' as separator in quantifier",
			',',
		)

		rightBrace := parsec.Eq[rune, int, parser.State](
			"expected '}' as ending of quantifier",
			'}',
		)

		full := parsec.Try(func(state parser.State, buf parsec.Buffer[rune, int]) (*quantity.Quantity, parsec.Error[int]) { // {1,1}
			pos := buf.Position()

			from, err := number(state, buf)
			if err != nil {
				return nil, err
			}

			_, err = comma(state, buf)
			if err != nil {
				return nil, err
			}

			to, err := number(state, buf)
			if err != nil {
				return nil, err
			}

			_, err = rightBrace(state, buf)
			if err != nil {
				return nil, err
			}

			if from > to {
				// TODO : move to validation?
				return nil, parsec.NewParseError(
					pos,
					"invalid quantifier",
				)
			}

			return quantity.New(from, to), nil
		})

		fromZero := parsec.Try(func(state parser.State, buf parsec.Buffer[rune, int]) (*quantity.Quantity, parsec.Error[int]) { // {,1}
			_, commaErr := comma(state, buf)
			if commaErr != nil {
				return nil, commaErr
			}

			to, numErr := number(state, buf)
			if numErr != nil {
				return nil, numErr
			}

			_, braceErr := rightBrace(state, buf)
			if braceErr != nil {
				return nil, braceErr
			}

			return quantity.New(0, to), nil
		})

		endless := parsec.Try(func(state parser.State, buf parsec.Buffer[rune, int]) (*quantity.Quantity, parsec.Error[int]) { // {1,}
			from, numErr := number(state, buf)
			if numErr != nil {
				return nil, numErr
			}

			_, commaErr := comma(state, buf)
			if commaErr != nil {
				return nil, commaErr
			}

			_, braceErr := rightBrace(state, buf)
			if braceErr != nil {
				return nil, braceErr
			}

			return quantity.NewEndlessQuantity(from), nil
		})

		fixed := parsec.Try(func(state parser.State, buf parsec.Buffer[rune, int]) (*quantity.Quantity, parsec.Error[int]) { // {1}
			from, err := number(state, buf)
			if err != nil {
				return nil, err
			}

			_, err = rightBrace(state, buf)
			if err != nil {
				return nil, err
			}

			return quantity.New(from, from), nil
		})

		return func(state parser.State, buf parsec.Buffer[rune, int]) (*quantity.Quantity, parsec.Error[int]) {
			pos := buf.Position()

			q, fullErr := full(state, buf)
			if fullErr == nil {
				return q, nil
			}

			q, fromZeroErr := fromZero(state, buf)
			if fromZeroErr == nil {
				return q, nil
			}

			q, endlessErr := endless(state, buf)
			if endlessErr == nil {
				return q, nil
			}

			q, fixedErr := fixed(state, buf)
			if fixedErr == nil {
				return q, nil
			}

			return nil, parsec.NewParseError(
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
	_ parsec.Combinator[rune, int, node.Alternation, parser.State],
	except ...rune,
) parsec.Combinator[rune, int, node.Node, parser.State] {
	parse := parsec.Many(
		10,
		parsec.Try(
			parsec.NoneOf[rune, int, parser.State](
				"expected comment",
				except...,
			),
		),
	)

	return func(state parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
		runes, err := parse(state, buf)
		if err != nil {
			return nil, err
		}

		return node.NewComment(string(runes)), nil
	}
}

func parseNumber(_ ...rune) parsec.Combinator[rune, int, int, parser.State] {
	const zero = rune('0')

	// TODO : don't ignore except param

	digit := parsec.Try(parsec.OneOf[rune, int, parser.State](
		"expected digit",
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	))

	return func(state parser.State, buf parsec.Buffer[rune, int]) (int, parsec.Error[int]) {
		token, err := digit(state, buf)
		if err != nil {
			return 0, err
		}

		number := int(token - zero)

		for {
			token, err := digit(state, buf)
			if err != nil {
				break
			}

			number = number * 10
			number += int(token - zero)
		}

		return number, nil
	}
}
