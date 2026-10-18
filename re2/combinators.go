package re2

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/parser"
	"github.com/okneniz/cliche/quantity"
	"github.com/okneniz/parsec"
)

func parseHexNumber(from, to int) parser.ParserBuilder[int] {
	return func(_ ...rune) parsec.Combinator[rune, int, int, *parser.State] {
		// TODO : don't ignore except, check it

		parse, err := parsec.Quantifier(
			"expected hex number, for example 12f or 1B",
			from, to,
			parsec.OneOf[rune, int, *parser.State](
				"expected at least one symbol of hex number, for example '0123456789abcdefABCDEF'",
				[]rune("0123456789abcdefABCDEF")...,
			),
		)
		if err != nil {
			panic(err.Error()) // TODO : remove panic
		}

		return func(state *parser.State, buf parsec.Buffer[rune, int]) (int, parsec.Error[int]) {
			pos := buf.Position()

			runes, err := parse(state, buf)
			if err != nil {
				return -1, err
			}

			str := strings.ToLower(string(runes))

			num, castErr := strconv.ParseInt(str, 16, 64)
			if castErr != nil {
				return -1, parsec.NewParseError(
					pos,
					"invalid hex number: "+castErr.Error(),
				)
			}

			return int(num), nil
		}
	}
}

func parseOctalCharNumber(size int) parser.ParserBuilder[int] {
	parseLeftBraces := parsec.Eq[rune, int, *parser.State](
		"expected '{' as beginning of octal number",
		'{',
	)

	parseRightBraces := parsec.Eq[rune, int, *parser.State](
		"expected '}' as ending of octal number",
		'}',
	)

	return func(_ ...rune) parsec.Combinator[rune, int, int, *parser.State] {
		// TODO : don't ignore except

		allowed := []rune("01234567")
		parse := parsec.Count(
			size,
			"expected at least one digit between 0 and 7 as octal number",
			parsec.OneOf[rune, int, *parser.State](
				"expecte digit between 0 and 7",
				allowed...,
			),
		)

		return func(state *parser.State, buf parsec.Buffer[rune, int]) (int, parsec.Error[int]) {
			pos := buf.Position()

			_, leftErr := parseLeftBraces(state, buf)
			if leftErr != nil {
				return -1, leftErr
			}

			runes, runesErr := parse(state, buf)
			if runesErr != nil {
				return -1, runesErr
			}

			_, rightErr := parseRightBraces(state, buf)
			if rightErr != nil {
				return -1, rightErr
			}

			str := strings.ToLower(string(runes))

			num, castErr := strconv.ParseInt(str, 8, 64)
			if castErr != nil {
				return -1, parsec.NewParseError(
					pos,
					"invalid octal number: "+castErr.Error(),
				)
			}

			return int(num), nil
		}
	}
}

func makeOptionsParser(
	opts map[rune]node.ScanOption,
	except ...rune,
) parsec.Combinator[rune, int, []node.ScanOption, *parser.State] {
	parseRune := parsec.NoneOf[rune, int, *parser.State](
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

func parseGroup(
	parseAlternation parsec.Combinator[rune, int, node.Alternation, *parser.State],
	except ...rune,
) parsec.Combinator[rune, int, node.Node, *parser.State] {
	parseCapturedGroup := func(
		state *parser.State,
		buf parsec.Buffer[rune, int],
	) (node.Node, parsec.Error[int]) {
		pos := buf.Position()

		index := state.GroupIndex
		state.GroupIndex++

		alt, altErr := parseAlternation(state, buf)
		if altErr != nil {
			return nil, parsec.NewParseError(
				pos,
				"expected alternation expression in non captured group",
				altErr,
			)
		}

		return node.NewGroup(alt, index, ""), nil
	}

	parseCapturedGroup = parsec.Try(parseCapturedGroup)

	prefix := parsec.Try(parsec.Eq[rune, int, *parser.State](
		"expected '?' as prefix of options for non captured group",
		'?',
	))

	// TODO : add prefix parseSomething
	comma := parsec.Try(parsec.Eq[rune, int, *parser.State](
		"expected ':' as suffix of options for non captured group",
		':',
	))

	optsDict := map[rune]node.ScanOption{
		'i': node.ScanOptionCaseInsensitive,
		's': node.ScanOptionMultiline,
	}

	parseOptions := parsec.Try(
		makeOptionsParser(optsDict, append(except, '-', ':')...),
	)

	parseDiableOptions := parsec.Try(
		parsec.Skip(
			parsec.Eq[rune, int, *parser.State](
				"expected '-' as prefix for disabled options",
				'-',
			),
			parseOptions,
		),
	)

	parseNotCapturedGroupWithOptions := parsec.Try(
		func(state *parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
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

			index := state.GroupIndex
			state.GroupIndex++

			alt, err := parseAlternation(state, buf)
			if err != nil {
				return nil, err
			}

			// TODO : move it to unify function
			//
			// (?i:foo) -> (?i)(?:foo)(?-i)

			beforeSwitcher := node.NewOptionsSwitcher(enable, disable)
			group := node.NewGroup(alt, index, "")
			afterSwitcher := node.NewOptionsSwitcher(disable, enable)

			beforeSwitcher.GetNestedNodes()[group.GetKey()] = group
			group.GetNestedNodes()[afterSwitcher.GetKey()] = afterSwitcher

			return beforeSwitcher, nil
		},
	)

	return func(state *parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
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
	parseAlternation parsec.Combinator[rune, int, node.Alternation, *parser.State],
	_ ...rune,
) parsec.Combinator[rune, int, node.Node, *parser.State] {
	return func(state *parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
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
	parseAlternation parsec.Combinator[rune, int, node.Alternation, *parser.State],
	except ...rune,
) parsec.Combinator[rune, int, node.Node, *parser.State] {
	except = append(except, '>')

	parseAllowedForNameSymbols := parsec.NoneOf[rune, int, *parser.State](
		fmt.Sprintf("expected any char exclude %v", except),
		except...,
	)

	parseEndOfName := parsec.Eq[rune, int, *parser.State](
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

	return parsec.Try(func(state *parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
		pos := buf.Position()

		name, err := parseGroupName(state, buf)
		if err != nil {
			return nil, err
		}

		index := state.GroupIndex
		state.GroupIndex++

		alt, err := parseAlternation(state, buf)
		if err != nil {
			return nil, parsec.NewParseError(
				pos,
				"expected named group",
				err,
			)
		}

		return node.NewGroup(alt, index, string(name)), nil
	})
}

func parseAtomicGroup(
	parseAlternation parsec.Combinator[rune, int, node.Alternation, *parser.State],
	_ ...rune,
) parsec.Combinator[rune, int, node.Node, *parser.State] {
	return func(state *parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
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

func parseQuantity() parser.ParserBuilder[*quantity.Quantity] {
	return func(except ...rune) parsec.Combinator[rune, int, *quantity.Quantity, *parser.State] {
		number := parseNumber(except...)

		comma := parsec.Eq[rune, int, *parser.State](
			"expected ',' as separator in quantifier",
			',',
		)

		rightBrace := parsec.Eq[rune, int, *parser.State](
			"expected '}' as ending of quantifier",
			'}',
		)

		full := func(state *parser.State, buf parsec.Buffer[rune, int]) (*quantity.Quantity, parsec.Error[int]) { // {1,1}
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
		}

		full = parsec.Try(full)

		fromZero := func(state *parser.State, buf parsec.Buffer[rune, int]) (*quantity.Quantity, parsec.Error[int]) { // {,1}
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
		}

		fromZero = parsec.Try(fromZero)

		endless := func(state *parser.State, buf parsec.Buffer[rune, int]) (*quantity.Quantity, parsec.Error[int]) { // {1,}
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
		}

		endless = parsec.Try(endless)

		fixed := parsec.Try(func(state *parser.State, buf parsec.Buffer[rune, int]) (*quantity.Quantity, parsec.Error[int]) { // {1}
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

		return func(state *parser.State, buf parsec.Buffer[rune, int]) (*quantity.Quantity, parsec.Error[int]) {
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
	_ parsec.Combinator[rune, int, node.Alternation, *parser.State],
	except ...rune,
) parsec.Combinator[rune, int, node.Node, *parser.State] {
	parse := parsec.Many(
		10,
		parsec.Try(
			parsec.NoneOf[rune, int, *parser.State](
				"expected comment",
				except...,
			),
		),
	)

	return func(state *parser.State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
		runes, err := parse(state, buf)
		if err != nil {
			return nil, err
		}

		return node.NewComment(string(runes)), nil
	}
}

func parseNumber(_ ...rune) parsec.Combinator[rune, int, int, *parser.State] {
	const zero = rune('0')

	// TODO : don't ignore except param

	digit := parsec.Try(parsec.OneOf[rune, int, *parser.State](
		"expected digit",
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	))

	return func(state *parser.State, buf parsec.Buffer[rune, int]) (int, parsec.Error[int]) {
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
