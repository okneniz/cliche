package parser

import (
	"golang.org/x/exp/slices"

	"github.com/okneniz/cliche/buf"
	"github.com/okneniz/cliche/encoding/unicode"
	"github.com/okneniz/cliche/node"
	c "github.com/okneniz/parsec/common"
)

type CustomParser struct {
	config *ParserConfig
	parse  c.Combinator[rune, int, node.Alternation]
}

type Option[T any] func(T)

// var RequireDebug fmt.Stringer
//
// func debug(message string, args ...interface{}) {
// 	if RequireDebug != nil {
// 		fmt.Printf(message, args...)
// 		fmt.Println(RequireDebug.String())
// 	}
// }

func New(opts ...Option[*ParserConfig]) *CustomParser {
	p := new(CustomParser)

	cfg := NewConfig()
	for _, apply := range opts {
		apply(cfg)
	}
	p.config = cfg

	p.parse = p.alternationParser('|')

	return p
}

func (p *CustomParser) Parse(str string) (node.Node, error) {
	buffer := buf.NewRunesBuffer(str)

	alt, err := p.parse(buffer)
	if err != nil {
		return nil, err
	}

	var newNode node.Node
	newNode = alt

	// TODO : move it to special component?
	variants := alt.GetVariants()
	if len(variants) == 1 {
		newNode = variants[0]
	}

	// TODO: add specail error class with merge method
	// (required to pretty errors (expected: char or class ... explanation))
	// and merge expectations in choice

	newNode.Traverse(func(x node.Node) {
		if len(x.GetNestedNodes()) == 0 {
			x.AddExpression(str)
		}
	})

	return newNode, nil
}

func (p *CustomParser) alternationParser(
	except ...rune,
) c.Combinator[rune, int, node.Alternation] {
	parseAny := c.NoneOf[rune, int](except...)
	parseClass := c.Try(p.classParser(except...))

	// TODO : simplify - remove c.Try
	parseNonClass := c.Choice(
		c.Try(p.config.nonClassConfig.items.parser(except...)),
		c.Try(func(buf c.Buffer[rune, int]) (node.Node, error) {
			x, err := parseAny(buf)
			if err != nil {
				return nil, err
			}

			return node.NewForTable(unicode.NewTable(x)), nil
		}),
	)

	var (
		parseGroup c.Combinator[rune, int, node.Node]
	)

	parseVariant := c.Try(p.chainParser(p.optionalQuantifierParser(
		func(buf c.Buffer[rune, int]) (node.Node, error) {
			group, err := parseGroup(buf)
			if err == nil {
				return group, nil
			}

			class, err := parseClass(buf)
			if err == nil {
				return class, nil
			}

			nonClass, err := parseNonClass(buf)
			if err == nil {
				return nonClass, nil
			}

			return nil, err
		},
		except...,
	)))

	parseSeparator := c.Eq[rune, int]('|')

	parseAlternation := func(buf c.Buffer[rune, int]) (node.Alternation, error) {
		variant, err := parseVariant(buf)
		if err != nil {
			return nil, err
		}

		variants := make([]node.Node, 0, 1)
		variants = append(variants, variant)

		for !buf.IsEOF() {
			pos := buf.Position()

			_, err = parseSeparator(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			variant, err = parseVariant(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			variants = append(variants, variant)
		}

		return node.NewAlternation(variants), nil
	}

	groupAlternation := parseAlternation
	if !slices.Contains(except, ')') {
		groupAlternation = p.alternationParser(append(except, ')')...)
	}

	parseGroup = c.Try(
		Parens(
			p.config.groupConfig.parser(
				groupAlternation,
				append(except, ')')...,
			),
		),
	)

	return parseAlternation
}

func (p *CustomParser) runeParser(
	scope *ParserScope[rune],
	except ...rune,
) c.Combinator[rune, int, rune] {
	parseRune := scope.parser(except...)
	parseAny := c.NoneOf[rune, int](except...)

	return c.Choice(
		c.Try(parseRune),
		parseAny,
	)
}

func (p *CustomParser) rangeOrCharParser(
	except ...rune,
) c.Combinator[rune, int, node.Table] {
	parseSeparator := c.Eq[rune, int]('-')

	parseRune := p.runeParser(
		p.config.classConfig.runes,
		except...,
	)

	return func(buf c.Buffer[rune, int]) (node.Table, error) {
		from, err := parseRune(buf)
		if err != nil {
			return nil, err
		}

		pos := buf.Position()

		_, err = parseSeparator(buf)
		if err != nil {
			buf.Seek(pos)
			return unicode.NewTable(from), nil
		}

		to, err := parseRune(buf)
		if err != nil {
			buf.Seek(pos)
			return unicode.NewTable(from), nil
		}

		// TODO : check bounds

		return unicode.NewTableByPredicate(func(x rune) bool {
			return from <= x && x <= to
		}), nil
	}
}

func (p *CustomParser) classParser(
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parseTable := p.classTableParser(except...)

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		table, err := parseTable(buf)
		if err != nil {
			return nil, err
		}

		return node.NewForTable(table), nil
	}
}

func (p *CustomParser) classTableParser(
	except ...rune,
) c.Combinator[rune, int, node.Table] {
	var (
		parseClass    c.Combinator[rune, int, node.Table]
		parseSubClass c.Combinator[rune, int, node.Table]
	)

	newExcept := []rune{']'} //append(except, ']')
	// TODO: remove try?
	parseClassItem := c.Try(p.config.classConfig.items.parser(newExcept...))
	parseClassChar := c.Try(p.rangeOrCharParser(newExcept...))

	// hack : implementation without parsec.Try, parsec.Choice
	// to avoid problems with references to func and recursive parsing
	parseTable := func(buf c.Buffer[rune, int]) (node.Table, error) {
		// range
		// item
		// subclass
		// char
		pos := buf.Position()

		classItem, err := parseClassItem(buf)
		if err == nil {
			return classItem, nil
		}

		buf.Seek(pos)

		subClass, err := parseSubClass(buf) // must be first?
		if err == nil {
			return subClass, nil
		}

		buf.Seek(pos)

		classChar, err := parseClassChar(buf)
		if err == nil {
			return classChar, nil
		}

		return nil, err
	}

	parseSequenceOfTables := c.Some(1, c.Try(parseTable))

	parsePositive := func(buf c.Buffer[rune, int]) (node.Table, error) {
		tables, err := parseSequenceOfTables(buf)
		if err != nil {
			return nil, err
		}

		return unicode.MergeTables(tables...), nil
	}

	parseNegative := c.Skip(
		c.Eq[rune, int]('^'),
		func(buf c.Buffer[rune, int]) (node.Table, error) {
			table, err := parsePositive(buf)
			if err != nil {
				return nil, err
			}

			return table.Invert(), nil
		},
	)

	parseClass = Squares(
		c.Choice(
			c.Try(parseNegative),
			parsePositive,
		),
	)

	if slices.Contains(except, ']') {
		parseSubClass = parseClass
	} else {
		parseSubClass = p.classTableParser(newExcept...)
	}

	return parseClass
}

func (p *CustomParser) optionalQuantifierParser(
	expression c.Combinator[rune, int, node.Node],
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parseQuantity := c.Try(p.config.quntityConfig.items.parser(except...))

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		exp, err := expression(buf)
		if err != nil {
			return nil, err
		}

		quantity, err := parseQuantity(buf)
		if err != nil {
			return exp, nil
		}

		x := node.NewQuantifier(quantity, exp)
		return x, nil
	}
}

func (p *CustomParser) chainParser(
	parse c.Combinator[rune, int, node.Node],
) c.Combinator[rune, int, node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		first, err := parse(buf)
		if err != nil {
			return nil, err
		}

		last := first

		for !buf.IsEOF() {
			pos := buf.Position()

			next, err := parse(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			last.GetNestedNodes()[next.GetKey()] = next
			last = next
		}

		return first, nil
	}
}
