package onigmo

import (
	"fmt"
	"unicode"

	unicodeEncoding "github.com/okneniz/cliche/encoding/unicode"
	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/parser"
	"github.com/okneniz/cliche/quantity"
)

var (
	alnum  = unicodeEncoding.NewTableByPredicate(func(x rune) bool { return unicode.IsLetter(x) || unicode.IsMark(x) || unicode.IsDigit(x) })
	alpha  = unicodeEncoding.NewTableByPredicate(func(x rune) bool { return unicode.IsLetter(x) || unicode.IsMark(x) })
	ascii  = unicodeEncoding.NewTableByPredicate(func(x rune) bool { return x < unicode.MaxASCII })
	blank  = unicodeEncoding.NewTableByPredicate(func(x rune) bool { return x == ' ' || x == '\t' })
	digit  = unicodeEncoding.NewTableByPredicate(unicode.IsDigit)
	lower  = unicodeEncoding.NewTableByPredicate(unicode.IsLower)
	upper  = unicodeEncoding.NewTableByPredicate(unicode.IsUpper)
	space  = unicodeEncoding.NewTableByPredicate(unicode.IsSpace)
	cntrl  = unicodeEncoding.NewTableByPredicate(unicode.IsControl)
	print  = unicodeEncoding.NewTableByPredicate(unicode.IsPrint)
	graph  = unicodeEncoding.NewTableByPredicate(func(x rune) bool { return unicode.IsGraphic(x) && !unicode.IsSpace(x) })
	punct  = unicodeEncoding.NewTableByPredicate(unicode.IsPunct)
	xdigit = unicodeEncoding.NewTableByPredicate(isHex)
	word   = unicodeEncoding.NewTableByPredicate(isWord)

	notAlnum  = alnum.Invert()
	notAlpha  = alpha.Invert()
	notAscii  = ascii.Invert()
	notBlank  = blank.Invert()
	notDigit  = digit.Invert()
	notLower  = lower.Invert()
	notUpper  = upper.Invert()
	notSpace  = space.Invert()
	notCntrl  = cntrl.Invert()
	notPrint  = print.Invert()
	notGraph  = graph.Invert()
	notPunct  = punct.Invert()
	notXdigit = xdigit.Invert()
	notWord   = word.Invert()

	parseDigit     = parser.NodeAsTable(parser.Const(digit))
	parseNotDigit  = parser.NodeAsTable(parser.Const(notDigit))
	parseWord      = parser.NodeAsTable(parser.Const(word))
	parseNotWord   = parser.NodeAsTable(parser.Const(notWord))
	parseSpace     = parser.NodeAsTable(parser.Const(space))
	parseNotSpace  = parser.NodeAsTable(parser.Const(notSpace))
	parseXdigit    = parser.NodeAsTable(parser.Const(xdigit))
	parseNotXdigit = parser.NodeAsTable(parser.Const(notXdigit))

	dot          = unicodeEncoding.NewTable('.')
	question     = unicodeEncoding.NewTable('?')
	plus         = unicodeEncoding.NewTable('+')
	asterisk     = unicodeEncoding.NewTable('*')
	circumFlexus = unicodeEncoding.NewTable('^')
	dollar       = unicodeEncoding.NewTable('$')
	leftBracket  = unicodeEncoding.NewTable('[')
	bar          = unicodeEncoding.NewTable('|')

	parseDot          = parser.NodeAsTable(parser.Const(dot))
	parseQuestion     = parser.NodeAsTable(parser.Const(question))
	parsePlus         = parser.NodeAsTable(parser.Const(plus))
	parseAsterisk     = parser.NodeAsTable(parser.Const(asterisk))
	parseCircumFlexus = parser.NodeAsTable(parser.Const(circumFlexus))

	parseDollar      = parser.NodeAsTable(parser.Const(dollar))
	parseLeftBracket = parser.NodeAsTable(parser.Const(leftBracket))
	parseBar         = parser.NodeAsTable(parser.Const(bar))

	// TODO : check size in different docs
	parseHexChar      = parser.NumberAsRune(parseHexNumber(2, 2))
	parseHexCharTable = parser.RuneAsTable(parseHexChar)
	parseHexCharNode  = parser.NodeAsTable(parseHexCharTable)

	parseOctalChar      = parser.NumberAsRune(braces(parseOctal(3)))
	parseOctalCharTable = parser.RuneAsTable(parseOctalChar)
	parseOctalCharNode  = parser.NodeAsTable(parseOctalCharTable)

	parseUnicodeChar  = parser.NumberAsRune(parseHexNumber(1, 4))
	parseUnicodeTable = parser.RuneAsTable(parseUnicodeChar)
	parseUnicodeNode  = parser.NodeAsTable(parseUnicodeTable)

	OnigmoParser = parser.New(func(cfg *parser.Config) {
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

		cfg.NonClass().
			Items().
			StringAsFunc(`\A`, node.NewStartOfString).
			StringAsFunc(`\z`, node.NewEndOfString).
			StringAsFunc(`\Z`, node.NewEndOfStringAndNewLine).
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
			WithPrefix(`\n`, parser.NodeAsTable(parser.Const(unicodeEncoding.NewTable('\n')))).
			WithPrefix(`\t`, parser.NodeAsTable(parser.Const(unicodeEncoding.NewTable('\t')))).
			WithPrefix(`\v`, parser.NodeAsTable(parser.Const(unicodeEncoding.NewTable('\u000B')))).
			WithPrefix(`\r`, parser.NodeAsTable(parser.Const(unicodeEncoding.NewTable('\r')))).
			WithPrefix(`\f`, parser.NodeAsTable(parser.Const(unicodeEncoding.NewTable('\f')))).
			WithPrefix(`\a`, parser.NodeAsTable(parser.Const(unicodeEncoding.NewTable('\a')))).
			WithPrefix(`\e`, parser.NodeAsTable(parser.Const(unicodeEncoding.NewTable('\u001b'))))

		cfg.Class().
			Items().
			WithPrefix(`\[`, parser.RuneAsTable(parser.Const('['))).
			WithPrefix(`\]`, parser.RuneAsTable(parser.Const(']'))).
			WithPrefix(`\^`, parser.RuneAsTable(parser.Const('^'))).
			WithPrefix(`\n`, parser.RuneAsTable(parser.Const('\n'))).
			WithPrefix(`\t`, parser.RuneAsTable(parser.Const('\t'))).
			WithPrefix(`\v`, parser.RuneAsTable(parser.Const('\u000B'))).
			WithPrefix(`\r`, parser.RuneAsTable(parser.Const('\r'))).
			WithPrefix(`\b`, parser.RuneAsTable(parser.Const('\b'))).
			WithPrefix(`\f`, parser.RuneAsTable(parser.Const('\f'))).
			WithPrefix(`\a`, parser.RuneAsTable(parser.Const('\a'))).
			WithPrefix(`\e`, parser.RuneAsTable(parser.Const('\u001b')))

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
			ParsePrefix("?<!", parseNegativeLookBehind).
			ParsePrefix("?#", parseComment)

		// TODO : parseInvalidQuantifier
		configureProperty(cfg, unicode.Properties)
		configureProperty(cfg, unicode.Scripts)
		configureProperty(cfg, unicode.Categories)

		cfg.Quntifier().Items().StringAsValue("?", quantity.New(0, 1))
		cfg.Quntifier().Items().StringAsValue("+", quantity.NewEndlessQuantity(1))
		cfg.Quntifier().Items().StringAsValue("*", quantity.NewEndlessQuantity(0))
		cfg.Quntifier().Items().WithPrefix("{", parseQuanty)
	})
)

func configureProperty(cfg *parser.Config, props map[string]*unicode.RangeTable) {
	for name, prop := range props {
		tbl := unicodeEncoding.NewTableByPredicate(func(r rune) bool {
			return unicode.In(r, prop)
		})

		positive := parser.Const(tbl)
		negative := parser.Const(tbl.Invert())

		cfg.NonClass().
			Items().
			WithPrefix(fmt.Sprintf("\\p{%s}", name), parser.NodeAsTable(positive)).
			WithPrefix(fmt.Sprintf("\\p{^%s}", name), parser.NodeAsTable(negative)).
			WithPrefix(fmt.Sprintf("\\P{%s}", name), parser.NodeAsTable(negative))

		cfg.Class().
			Items().
			WithPrefix(fmt.Sprintf("\\p{%s}", name), positive).
			WithPrefix(fmt.Sprintf("\\p{^%s}", name), negative).
			WithPrefix(fmt.Sprintf("\\P{%s}", name), negative)
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
