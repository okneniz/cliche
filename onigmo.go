package cliche

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	unicodeEncoding "github.com/okneniz/cliche/encoding/unicode"
	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/parser"
	c "github.com/okneniz/parsec/common"
)

var _ node.Table = new(unicodeEncoding.UnicodeTable)

// TODO : try to explain it in doc for contributors
//
// split to another groups of options
//
// common:
//   - chars (value)
//     - as is (a, b, 1)
//     - with prefix (\u{123}, \x017)
//   - escaped meta chars (range of value)
//   - classes (range of value)
//
// not in class:
//   - groups
//   - assertions (lookahead / lookbehind)
//   - alternative
//   - anchors: (match positions)
//   	- ^, $
//   	- \A, \z
// 	 - spec symbols - [(|
//   - quantifiers *+?
//   - meta chars - ^$.
//
// in class:
//   - bracket
//   - ranges from chars
// 	 - spec symbols - ^])

// node
//   - match node (chars, classes)
//     - return false if bounds out of ranges)
//	   - yield only not empty span
//   - position node (anchor)
//	   - yield only empty span
//   - special consuctions (group, alternation, assertions)
//     - capture internal sub expression

var (
	OnigmoParser = parser.New(func(cfg *parser.ParserConfig) {
		alnum := unicodeEncoding.NewTableByPredicate(func(x rune) bool { return unicode.IsLetter(x) || unicode.IsMark(x) || unicode.IsDigit(x) })
		alpha := unicodeEncoding.NewTableByPredicate(func(x rune) bool { return unicode.IsLetter(x) || unicode.IsMark(x) })
		ascii := unicodeEncoding.NewTableByPredicate(func(x rune) bool { return x < unicode.MaxASCII })
		blank := unicodeEncoding.NewTableByPredicate(func(x rune) bool { return x == ' ' || x == '\t' })
		digit := unicodeEncoding.NewTableByPredicate(unicode.IsDigit)
		lower := unicodeEncoding.NewTableByPredicate(unicode.IsLower)
		upper := unicodeEncoding.NewTableByPredicate(unicode.IsUpper)
		space := unicodeEncoding.NewTableByPredicate(unicode.IsSpace)
		cntrl := unicodeEncoding.NewTableByPredicate(unicode.IsControl)
		print := unicodeEncoding.NewTableByPredicate(unicode.IsPrint)
		graph := unicodeEncoding.NewTableByPredicate(func(x rune) bool { return unicode.IsGraphic(x) && !unicode.IsSpace(x) })
		punct := unicodeEncoding.NewTableByPredicate(unicode.IsPunct)
		xdigit := unicodeEncoding.NewTableByPredicate(isHex)
		word := unicodeEncoding.NewTableByPredicate(isWord)

		notAlnum := alnum.Invert()
		notAlpha := alpha.Invert()
		notAscii := ascii.Invert()
		notBlank := blank.Invert()
		notDigit := digit.Invert()
		notLower := lower.Invert()
		notUpper := upper.Invert()
		notSpace := space.Invert()
		notCntrl := cntrl.Invert()
		notPrint := print.Invert()
		notGraph := graph.Invert()
		notPunct := punct.Invert()
		notXdigit := xdigit.Invert()
		notWord := word.Invert()

		cfg.Class().
			Items().
			StringAsValue("[:alnum:]", alnum).
			StringAsValue("[:alpha:]", alpha).
			StringAsValue("[:ascii:]", ascii).
			StringAsValue("[:blank:]", blank).
			StringAsValue("[:digit:]", digit).
			StringAsValue("[:lower:]", lower).
			StringAsValue("[:upper:]", upper).
			StringAsValue("[:space:]", space).
			StringAsValue("[:cntrl:]", cntrl).
			StringAsValue("[:print:]", print).
			StringAsValue("[:graph:]", graph).
			StringAsValue("[:punct:]", punct).
			StringAsValue("[:xdigit:]", xdigit).
			StringAsValue("[:word:]", word).
			StringAsValue("[:^alnum:]", notAlnum).
			StringAsValue("[:^alpha:]", notAlpha).
			StringAsValue("[:^ascii:]", notAscii).
			StringAsValue("[:^blank:]", notBlank).
			StringAsValue("[:^digit:]", notDigit).
			StringAsValue("[:^lower:]", notLower).
			StringAsValue("[:^upper:]", notUpper).
			StringAsValue("[:^space:]", notSpace).
			StringAsValue("[:^cntrl:]", notCntrl).
			StringAsValue("[:^print:]", notPrint).
			StringAsValue("[:^graph:]", notGraph).
			StringAsValue("[:^punct:]", notPunct).
			StringAsValue("[:^xdigit:]", notXdigit).
			StringAsValue("[:^word:]", notWord).
			StringAsValue(`\d`, digit).
			StringAsValue(`\D`, notDigit).
			StringAsValue(`\w`, word).
			StringAsValue(`\W`, notWord).
			StringAsValue(`\s`, space).
			StringAsValue(`\S`, notSpace).
			StringAsValue(`\h`, xdigit).
			StringAsValue(`\H`, notXdigit)

		parseDigit := parser.NodeAsTable(parser.Const(digit))
		parseNotDigit := parser.NodeAsTable(parser.Const(notDigit))
		parseWord := parser.NodeAsTable(parser.Const(word))
		parseNotWord := parser.NodeAsTable(parser.Const(notWord))
		parseSpace := parser.NodeAsTable(parser.Const(space))
		parseNotSpace := parser.NodeAsTable(parser.Const(notSpace))
		parseXdigit := parser.NodeAsTable(parser.Const(xdigit))
		parseNotXdigit := parser.NodeAsTable(parser.Const(notXdigit))

		cfg.NonClass().
			Items().
			StringAsFunc(`\A`, node.NewStartOfString).
			StringAsFunc(`\z`, node.NewEndOfString).
			StringAsFunc(`\K`, node.NewKeep).
			StringAsFunc(`.`, node.NewDot).
			StringAsFunc(`^`, node.NewStartOfLine).
			StringAsFunc(`$`, node.NewEndOfLine).
			StringAsFunc(`\b`, node.NewWordBoundary).
			StringAsFunc(`\B`, node.NewNonWordBoundary).
			WithPrefix(`\d`, parseDigit).
			WithPrefix(`\D`, parseNotDigit).
			WithPrefix(`\w`, parseWord).
			WithPrefix(`\W`, parseNotWord).
			WithPrefix(`\s`, parseSpace).
			WithPrefix(`\S`, parseNotSpace).
			WithPrefix(`\h`, parseXdigit).
			WithPrefix(`\H`, parseNotXdigit)

		dot := unicodeEncoding.NewTableFor('.')
		question := unicodeEncoding.NewTableFor('?')
		plus := unicodeEncoding.NewTableFor('+')
		asterisk := unicodeEncoding.NewTableFor('*')
		circumFlexus := unicodeEncoding.NewTableFor('^')
		dollar := unicodeEncoding.NewTableFor('$')
		leftBracket := unicodeEncoding.NewTableFor('[')
		bar := unicodeEncoding.NewTableFor('|')

		parseDot := parser.NodeAsTable(parser.Const(dot))
		parseQuestion := parser.NodeAsTable(parser.Const(question))
		parsePlus := parser.NodeAsTable(parser.Const(plus))
		parseAsterisk := parser.NodeAsTable(parser.Const(asterisk))
		parseCircumFlexus := parser.NodeAsTable(parser.Const(circumFlexus))

		parseDollar := parser.NodeAsTable(parser.Const(dollar))
		parseLeftBracket := parser.NodeAsTable(parser.Const(leftBracket))
		parseBar := parser.NodeAsTable(parser.Const(bar))

		cfg.NonClass().
			Items().
			Parse(parseBackReference).
			WithPrefix(`\.`, parseDot).
			WithPrefix(`\?`, parseQuestion).
			WithPrefix(`\+`, parsePlus).
			WithPrefix(`\*`, parseAsterisk).
			WithPrefix(`\^`, parseCircumFlexus).
			WithPrefix(`\$`, parseDollar).
			WithPrefix(`\[`, parseLeftBracket).
			WithPrefix(`\|`, parseBar).
			WithPrefix(`\n`, parser.NodeAsTable(parser.Const(unicodeEncoding.NewTableFor('\n')))).
			WithPrefix(`\t`, parser.NodeAsTable(parser.Const(unicodeEncoding.NewTableFor('\t')))).
			WithPrefix(`\v`, parser.NodeAsTable(parser.Const(unicodeEncoding.NewTableFor('\u000B')))).
			WithPrefix(`\r`, parser.NodeAsTable(parser.Const(unicodeEncoding.NewTableFor('\r')))).
			WithPrefix(`\f`, parser.NodeAsTable(parser.Const(unicodeEncoding.NewTableFor('\f')))).
			WithPrefix(`\a`, parser.NodeAsTable(parser.Const(unicodeEncoding.NewTableFor('\a')))).
			WithPrefix(`\e`, parser.NodeAsTable(parser.Const(unicodeEncoding.NewTableFor('\u001b'))))

		cfg.Class().
			Items().
			WithPrefix(`\[`, parser.RuneAsTable(parser.Const('['))).
			WithPrefix(`\]`, parser.RuneAsTable(parser.Const(']'))).
			WithPrefix(`\n`, parser.RuneAsTable(parser.Const('\n'))).
			WithPrefix(`\t`, parser.RuneAsTable(parser.Const('\t'))).
			WithPrefix(`\v`, parser.RuneAsTable(parser.Const('\u000B'))).
			WithPrefix(`\r`, parser.RuneAsTable(parser.Const('\r'))).
			WithPrefix(`\b`, parser.RuneAsTable(parser.Const('\b'))).
			WithPrefix(`\f`, parser.RuneAsTable(parser.Const('\f'))).
			WithPrefix(`\a`, parser.RuneAsTable(parser.Const('\a'))).
			WithPrefix(`\e`, parser.RuneAsTable(parser.Const('\u001b')))

		// TODO : check size in different docs
		parseHexChar := parser.NumberAsRune(parseHexNumber(2, 2))
		parseHexCharTable := parser.RuneAsTable(parseHexChar)
		parseHexCharNode := parser.NodeAsTable(parseHexCharTable)

		parseOctalChar := parser.NumberAsRune(Braces(parseOctal(3)))
		parseOctalCharTable := parser.RuneAsTable(parseOctalChar)
		parseOctalCharNode := parser.NodeAsTable(parseOctalCharTable)

		parseUnicodeChar := parser.NumberAsRune(parseHexNumber(1, 4))
		parseUnicodeTable := parser.RuneAsTable(parseUnicodeChar)
		parseUnicodeNode := parser.NodeAsTable(parseUnicodeTable)

		cfg.Class().
			Runes().
			WithPrefix(`\x`, parseHexChar).
			WithPrefix(`\o`, parseOctalChar).
			WithPrefix(`\u`, parseUnicodeChar)

		cfg.Class().
			Items().
			WithPrefix(`\o`, parseOctalCharTable).
			WithPrefix(`\u`, parseUnicodeTable)

		cfg.NonClass().
			Items().
			WithPrefix(`\x`, parseHexCharNode).
			WithPrefix(`\o`, parseOctalCharNode).
			WithPrefix(`\u`, parseUnicodeNode).
			WithPrefix(`\k`, parseNameReference)

		cfg.Groups().
			Parse(parseCondition).
			Parse(parseGroup).
			ParsePrefix("?:", parseNotCapturedGroup).
			ParsePrefix("?<", parseNamedGroup).
			ParsePrefix("?>", parseAtomicGroup).
			ParsePrefix("?=", parseLookAhead).
			ParsePrefix("?!", parseNegativeLookAhead).
			ParsePrefix("?<=", parseLookBehind).
			ParsePrefix("?<!", parseNegativeLookBehind)

		// TODO : parseInvalidQuantifier
		configureProperty(cfg, unicode.Properties)
		configureProperty(cfg, unicode.Scripts)
		configureProperty(cfg, unicode.Categories)
	})
)

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

func Braces[T any](makeParser parser.ParserBuilder[T]) parser.ParserBuilder[T] {
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
	parse := c.Skip[rune, int](
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

func configureProperty(cfg *parser.ParserConfig, props map[string]*unicode.RangeTable) {
	for name, prop := range props {
		tbl := unicodeEncoding.NewTable(prop)

		cfg.NonClass().
			Items().
			WithPrefix(fmt.Sprintf("\\p{%s}", name), parser.NodeAsTable(parser.Const(tbl))).
			WithPrefix(fmt.Sprintf("\\p{^%s}", name), parser.NodeAsTable(parser.Const(tbl.Invert()))).
			WithPrefix(fmt.Sprintf("\\P{%s}", name), parser.NodeAsTable(parser.Const(tbl.Invert())))

		cfg.Class().
			Items().
			WithPrefix(fmt.Sprintf("\\p{%s}", name), parser.Const(tbl)).
			WithPrefix(fmt.Sprintf("\\p{^%s}", name), parser.Const(tbl.Invert())).
			WithPrefix(fmt.Sprintf("\\P{%s}", name), parser.Const(tbl.Invert()))
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

	reference := parser.TryAll(
		parseBackReference,
		parseNameReference,
	)

	condition := parser.Parens(reference)
	before := parser.SkipString("?")

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

func isWord(x rune) bool {
	return x == '_' || unicode.IsLetter(x) || unicode.IsDigit(x)
}

func isHex(x rune) bool {
	return x >= '0' && x <= '9' ||
		x >= 'a' && x <= 'f' ||
		x >= 'A' && x <= 'F'
}
