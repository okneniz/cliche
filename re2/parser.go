package re2

import (
	"fmt"
	"unicode"

	unicodeEncoding "github.com/okneniz/cliche/encoding/unicode"
	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/parser"
	"github.com/okneniz/cliche/quantity"
)

// DOC - https://pkg.go.dev/regexp/syntax
//
// Why without backreferences:
// - https://swtch.com/~rsc/regexp/regexp3.html
// - https://en.wikipedia.org/wiki/ReDoS

var (
	asciiAlnum  = unicodeEncoding.NewTableByPredicate(unicode.MaxASCII, func(x rune) bool { return unicode.IsLetter(x) || unicode.IsMark(x) || unicode.IsDigit(x) })
	asciiAlpha  = unicodeEncoding.NewTableByPredicate(unicode.MaxASCII, func(x rune) bool { return unicode.IsLetter(x) || unicode.IsMark(x) })
	ascii       = unicodeEncoding.NewTableByPredicate(unicode.MaxASCII, isAscii)
	asciiBlank  = unicodeEncoding.NewTableByPredicate(unicode.MaxASCII, func(x rune) bool { return x == ' ' || x == '\t' })
	asciiDigit  = unicodeEncoding.NewTableByPredicate(unicode.MaxASCII, unicode.IsDigit)
	asciiLower  = unicodeEncoding.NewTableByPredicate(unicode.MaxASCII, unicode.IsLower)
	asciiUpper  = unicodeEncoding.NewTableByPredicate(unicode.MaxASCII, unicode.IsUpper)
	asciiSpace  = unicodeEncoding.NewTableByPredicate(unicode.MaxASCII, unicode.IsSpace)
	asciiCntrl  = unicodeEncoding.NewTableByPredicate(unicode.MaxASCII, unicode.IsControl)
	asciiPrint  = unicodeEncoding.NewTableByPredicate(unicode.MaxASCII, unicode.IsPrint)
	asciiGraph  = unicodeEncoding.NewTableByPredicate(unicode.MaxASCII, func(x rune) bool { return unicode.IsGraphic(x) && !unicode.IsSpace(x) })
	asciiPunct  = unicodeEncoding.NewTableByPredicate(unicode.MaxASCII, unicode.IsPunct)
	asciiXdigit = unicodeEncoding.NewTableByPredicate(unicode.MaxASCII, isHex)
	asciiWord   = unicodeEncoding.NewTableByPredicate(unicode.MaxASCII, isWord)

	// TODO : rename to AsciiNonAlnum
	//
	// [[:cntrl:]] - ascii cntrl symbols
	// but
	// [^[:cntrl:]] - non ascii cntrl symbols, includes unicode symbols
	//
	// re := regexp.MustCompile(`[[:^cntrl:]]+`)
	// result := re.FindAllStringSubmatch("foo é Bar\n1\t2", -1)
	// fmt.Println("RESULT", result)
	//
	// => RESULT [[foo é Bar] [1] [2]]

	notAsciiAlnum  = asciiAlnum.Invert(unicode.MaxRune)
	notAsciiAlpha  = asciiAlpha.Invert(unicode.MaxRune)
	notAscii       = ascii.Invert(unicode.MaxRune)
	notAsciiBlank  = asciiBlank.Invert(unicode.MaxRune)
	notAsciiDigit  = asciiDigit.Invert(unicode.MaxRune)
	notAsciiLower  = asciiLower.Invert(unicode.MaxRune)
	notAsciiUpper  = asciiUpper.Invert(unicode.MaxRune)
	notAsciiSpace  = asciiSpace.Invert(unicode.MaxRune)
	notAsciiCntrl  = asciiCntrl.Invert(unicode.MaxRune)
	notAsciiPrint  = asciiPrint.Invert(unicode.MaxRune)
	notAsciiGraph  = asciiGraph.Invert(unicode.MaxRune)
	notAsciiPunct  = asciiPunct.Invert(unicode.MaxRune)
	notAsciiXdigit = asciiXdigit.Invert(unicode.MaxRune)
	notAsciiWord   = asciiWord.Invert(unicode.MaxRune)

	parseAsciiDigit     = parser.TableAsClass(parser.Const(asciiDigit))
	parseAsciiNotDigit  = parser.TableAsClass(parser.Const(notAsciiDigit))
	parseAsciiWord      = parser.TableAsClass(parser.Const(asciiWord))
	parseAsciiNotWord   = parser.TableAsClass(parser.Const(notAsciiWord))
	parseAsciiSpace     = parser.TableAsClass(parser.Const(asciiSpace))
	parseAsciiNotSpace  = parser.TableAsClass(parser.Const(notAsciiSpace))
	parseAsciiXdigit    = parser.TableAsClass(parser.Const(asciiXdigit))
	parseAsciiNotXdigit = parser.TableAsClass(parser.Const(notAsciiXdigit))

	dot          = unicodeEncoding.NewTable('.')
	question     = unicodeEncoding.NewTable('?')
	plus         = unicodeEncoding.NewTable('+')
	asterisk     = unicodeEncoding.NewTable('*')
	circumFlexus = unicodeEncoding.NewTable('^')
	dollar       = unicodeEncoding.NewTable('$')
	leftBracket  = unicodeEncoding.NewTable('[')
	rightBracket = unicodeEncoding.NewTable(']')
	leftParens   = unicodeEncoding.NewTable('(')
	rightParens  = unicodeEncoding.NewTable(')')
	rightBraces  = unicodeEncoding.NewTable('}')

	bar = unicodeEncoding.NewTable('|')

	parseDot          = parser.TableAsClass(parser.Const(dot))
	parseQuestion     = parser.TableAsClass(parser.Const(question))
	parsePlus         = parser.TableAsClass(parser.Const(plus))
	parseAsterisk     = parser.TableAsClass(parser.Const(asterisk))
	parseCircumFlexus = parser.TableAsClass(parser.Const(circumFlexus))

	parseDollar       = parser.TableAsClass(parser.Const(dollar))
	parseLeftBracket  = parser.TableAsClass(parser.Const(leftBracket))
	parseRightBracket = parser.TableAsClass(parser.Const(rightBracket))
	parseLeftParens   = parser.TableAsClass(parser.Const(leftParens))
	parseRightParens  = parser.TableAsClass(parser.Const(rightParens))
	parseRightBraces  = parser.TableAsClass(parser.Const(rightBraces))
	parseBar          = parser.TableAsClass(parser.Const(bar))

	// TODO : check size in different docs
	parseHexChar      = parser.NumberAsRune(parseHexNumber(2, 2))
	parseHexCharTable = parser.RuneAsTable(parseHexChar)
	parseHexCharNode  = parser.TableAsClass(parseHexCharTable)

	parseOctalChar      = parser.NumberAsRune(parseOctalCharNumber(3))
	parseOctalCharTable = parser.RuneAsTable(parseOctalChar)
	parseOctalCharNode  = parser.TableAsClass(parseOctalCharTable)

	parseUnicodeChar  = parser.NumberAsRune(parseHexNumber(1, 4))
	parseUnicodeTable = parser.RuneAsTable(parseUnicodeChar)
	parseUnicodeNode  = parser.TableAsClass(parseUnicodeTable)

	Parser = parser.New(func(cfg *parser.Config) {
		cfg.Class().
			Items().
			StringAsValue("[:alnum:]", asciiAlnum).
			StringAsValue("[:alpha:]", asciiAlpha).
			StringAsValue("[:ascii:]", ascii).
			StringAsValue("[:blank:]", asciiBlank).
			StringAsValue("[:cntrl:]", asciiCntrl).
			StringAsValue("[:digit:]", asciiDigit).
			StringAsValue("[:graph:]", asciiGraph).
			StringAsValue("[:lower:]", asciiLower).
			StringAsValue("[:print:]", asciiPrint).
			StringAsValue("[:punct:]", asciiPunct).
			StringAsValue("[:space:]", asciiSpace).
			StringAsValue("[:upper:]", asciiUpper).
			StringAsValue("[:word:]", asciiWord).
			StringAsValue("[:xdigit:]", asciiXdigit).
			StringAsValue("[:^alnum:]", notAsciiAlnum).
			StringAsValue("[:^alpha:]", notAsciiAlpha).
			StringAsValue("[:^ascii:]", notAscii).
			StringAsValue("[:^blank:]", notAsciiBlank).
			StringAsValue("[:^cntrl:]", notAsciiCntrl).
			StringAsValue("[:^digit:]", notAsciiDigit).
			StringAsValue("[:^graph:]", notAsciiGraph).
			StringAsValue("[:^lower:]", notAsciiLower).
			StringAsValue("[:^print:]", notAsciiPrint).
			StringAsValue("[:^punct:]", notAsciiPunct).
			StringAsValue("[:^space:]", notAsciiSpace).
			StringAsValue("[:^upper:]", notAsciiUpper).
			StringAsValue("[:^word:]", notAsciiWord).
			StringAsValue("[:^xdigit:]", notAsciiXdigit).
			StringAsValue(`\d`, asciiDigit).
			StringAsValue(`\D`, notAsciiDigit).
			StringAsValue(`\w`, asciiWord).
			StringAsValue(`\W`, notAsciiWord).
			StringAsValue(`\s`, asciiSpace).
			StringAsValue(`\S`, notAsciiSpace).
			StringAsValue(`\h`, asciiDigit).
			StringAsValue(`\H`, notAsciiXdigit)

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
			WithPrefix(`\d`, parseAsciiDigit).
			WithPrefix(`\D`, parseAsciiNotDigit).
			WithPrefix(`\w`, parseAsciiWord).
			WithPrefix(`\W`, parseAsciiNotWord).
			WithPrefix(`\s`, parseAsciiSpace).
			WithPrefix(`\S`, parseAsciiNotSpace).
			WithPrefix(`\h`, parseAsciiXdigit).
			WithPrefix(`\H`, parseAsciiNotXdigit)

		cfg.NonClass().
			Items().
			WithPrefix(`\.`, parseDot).
			WithPrefix(`\?`, parseQuestion).
			WithPrefix(`\+`, parsePlus).
			WithPrefix(`\*`, parseAsterisk).
			WithPrefix(`\^`, parseCircumFlexus).
			WithPrefix(`\$`, parseDollar).
			WithPrefix(`\[`, parseLeftBracket).
			WithPrefix(`\]`, parseRightBracket).
			WithPrefix(`\(`, parseLeftParens).
			WithPrefix(`\)`, parseRightParens).
			WithPrefix(`\}`, parseRightBraces).
			WithPrefix(`\|`, parseBar).
			WithPrefix(`\n`, parser.TableAsClass(parser.Const(unicodeEncoding.NewTable('\n')))).
			WithPrefix(`\t`, parser.TableAsClass(parser.Const(unicodeEncoding.NewTable('\t')))).
			WithPrefix(`\v`, parser.TableAsClass(parser.Const(unicodeEncoding.NewTable('\u000B')))).
			WithPrefix(`\r`, parser.TableAsClass(parser.Const(unicodeEncoding.NewTable('\r')))).
			WithPrefix(`\f`, parser.TableAsClass(parser.Const(unicodeEncoding.NewTable('\f')))).
			WithPrefix(`\a`, parser.TableAsClass(parser.Const(unicodeEncoding.NewTable('\a')))).
			WithPrefix(`\e`, parser.TableAsClass(parser.Const(unicodeEncoding.NewTable('\u001b'))))

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
			WithPrefix(`\u`, parseUnicodeNode)

		cfg.Groups().
			Parse(parseGroup).
			ParsePrefix("?:", parseNotCapturedGroup).
			ParsePrefix("?<", parseNamedGroup).
			ParsePrefix("?P<", parseNamedGroup).
			ParsePrefix("?>", parseAtomicGroup).
			ParsePrefix("?#", parseComment)

		// TODO : check properties in re2 doc
		configureProperty(cfg, unicode.Properties)
		configureProperty(cfg, unicode.Scripts)
		configureProperty(cfg, unicode.Categories)

		cfg.Quantifier().Items().StringAsValue("?", quantity.New(0, 1))
		cfg.Quantifier().Items().StringAsValue("+", quantity.NewEndlessQuantity(1))
		cfg.Quantifier().Items().StringAsValue("*", quantity.NewEndlessQuantity(0))
		cfg.Quantifier().Items().WithPrefix("{", parseQuantity())
	})
)

func configureProperty(cfg *parser.Config, props map[string]*unicode.RangeTable) {
	for name, prop := range props {
		tbl := unicodeEncoding.NewTableByPredicate(unicode.MaxRune, func(r rune) bool {
			return unicode.In(r, prop)
		})

		positive := parser.Const(tbl)
		negative := parser.Const(tbl.Invert(unicode.MaxRune))

		cfg.NonClass().
			Items().
			WithPrefix(fmt.Sprintf("\\p{%s}", name), parser.TableAsClass(positive)).
			WithPrefix(fmt.Sprintf("\\p{^%s}", name), parser.TableAsClass(negative)).
			WithPrefix(fmt.Sprintf("\\P{%s}", name), parser.TableAsClass(negative))

		cfg.Class().
			Items().
			WithPrefix(fmt.Sprintf("\\p{%s}", name), positive).
			WithPrefix(fmt.Sprintf("\\p{^%s}", name), negative).
			WithPrefix(fmt.Sprintf("\\P{%s}", name), negative)
	}
}

// TODO : move to encoding package?

func isAscii(x rune) bool {
	return x <= unicode.MaxASCII
}

func isWord(x rune) bool {
	return x >= '0' && x <= '9' ||
		x >= 'a' && x <= 'z' ||
		x >= 'A' && x <= 'Z' ||
		x == '_'
}

func isHex(x rune) bool {
	return x >= '0' && x <= '9' ||
		x >= 'a' && x <= 'f' ||
		x >= 'A' && x <= 'F'
}
