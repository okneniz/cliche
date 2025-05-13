package parser

import (
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/okneniz/cliche/buf"
	"github.com/okneniz/cliche/encoding/unicode"
	"github.com/okneniz/cliche/node"
	c "github.com/okneniz/parsec/common"
)

// TODO : можно давать комбинаторам имена чтобы генерировать
// красивую ошибку с пояснением (что ожидается и что на самом деле)
//
// тогда комбинаторы должны быть или структурой или обернуты в структуру
//
// move prefixes parser to special builder
// to simplify Config to list of combinators
// (usefull for orders of combinators - apply in same orders as added)
//
// move it to parsec in future

// Что в итоге должно быть
//
// ### Basic syntax
//
//   - `\` escape (enable or disable meta character)
//   - `|` alternation
//   - `(...)` group
//   - `[...]` character class

// TODO : improve names (remove "Parser" prefixes)

type CustomParser struct {
	config *ParserConfig
	parse  c.Combinator[rune, int, node.Alternation]
}

type ParserConfig struct {
	nonClassConfig *NonClassParserConfig
	groupConfig    *GroupParserConfig
	classConfig    *ClassParserConfig
}

type Option[T any] func(T)

func NewConfig() *ParserConfig {
	cfg := new(ParserConfig)

	cfg.nonClassConfig = new(NonClassParserConfig)
	cfg.nonClassConfig.items = NewParserScope[node.Node]()

	cfg.groupConfig = new(GroupParserConfig)
	cfg.groupConfig.prefixes = make(map[string]GroupParserBuilder[node.Node], 0)
	cfg.groupConfig.parsers = make([]GroupParserBuilder[node.Node], 0)

	cfg.classConfig = new(ClassParserConfig)
	cfg.classConfig.runes = NewParserScope[rune]()
	cfg.classConfig.items = NewParserScope[node.Table]()

	return cfg
}

func (cfg *ParserConfig) Groups() *GroupParserConfig {
	return cfg.groupConfig
}

func (cfg *ParserConfig) Class() *ClassParserConfig {
	return cfg.classConfig
}

func (cfg *ParserConfig) NonClass() *NonClassParserConfig {
	return cfg.nonClassConfig
}

type ParserScope[T any] struct {
	prefixes map[string]ParserBuilder[T]
	parsers  []ParserBuilder[T]
}

func NewParserScope[T any]() *ParserScope[T] {
	scope := new(ParserScope[T])
	scope.prefixes = make(map[string]ParserBuilder[T], 0)
	scope.parsers = make([]ParserBuilder[T], 0)
	return scope
}

func (scope *ParserScope[T]) Parse(
	builders ...ParserBuilder[T],
) *ParserScope[T] {
	scope.parsers = append(scope.parsers, builders...)
	return scope
}

func (scope *ParserScope[T]) WithPrefix(
	prefix string, builder ParserBuilder[T],
) *ParserScope[T] {
	scope.prefixes[prefix] = builder
	return scope
}

func (scope *ParserScope[T]) StringAsValue(
	prefix string, value T,
) *ParserScope[T] {
	return scope.WithPrefix(prefix, Const(value))
}

func (scope *ParserScope[T]) StringAsFunc(
	prefix string, nodeBuilder func() T,
) *ParserScope[T] {
	return scope.WithPrefix(
		prefix,
		func(_ ...rune) c.Combinator[rune, int, T] {
			return func(_ c.Buffer[rune, int]) (T, error) {
				return nodeBuilder(), nil
			}
		},
	)
}

func (scope *ParserScope[T]) parser(except ...rune) c.Combinator[rune, int, T] {
	parseAny := c.Any[rune, int]() // to parse prefix rune by rune

	parseScopeByPrefix := NewParserTree(
		parseAny,
		scope.prefixes,
		except...,
	)

	parsers := make([]c.Combinator[rune, int, T], 0, len(scope.parsers)+1)
	parsers = append(parsers, c.Try(parseScopeByPrefix))

	for _, buildParser := range scope.parsers {
		parser := buildParser(except...)
		parsers = append(parsers, c.Try(parser))
	}

	return c.Choice(parsers...)
}

func (scope *ParserScope[T]) String() string {
	return fmt.Sprintf("%T{%v}", scope, scope.prefixes)
}

type NonClassParserConfig struct {
	// escaped char (\u{123}, \A, \z)
	// escaped range of char (\d, \w, \p{Property})
	items *ParserScope[node.Node]
}

func (cfg *NonClassParserConfig) Items() *ParserScope[node.Node] {
	return cfg.items
}

// as in NonClass, but some char have another meaning (for example $, ^)
// don't have anchors
type ClassParserConfig struct {
	// \u{00E0}, \A, \z
	runes *ParserScope[rune]
	items *ParserScope[node.Table]
}

func (cfg *ClassParserConfig) Runes() *ParserScope[rune] {
	return cfg.runes
}

func (cfg *ClassParserConfig) Items() *ParserScope[node.Table] {
	return cfg.items
}

func (cfg *ClassParserConfig) String() string {
	return fmt.Sprintf("%T{%v}", cfg, cfg.items)
}

// can use parser from common, need only for named group, captured group, etc
type GroupParserConfig struct {
	prefixes map[string]GroupParserBuilder[node.Node]
	parsers  []GroupParserBuilder[node.Node] // alternation wrapped to node
}

type GroupParserBuilder[T any] func(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	except ...rune,
) c.Combinator[rune, int, T]

func (cfg *GroupParserConfig) Parse(
	builders ...GroupParserBuilder[node.Node],
) *GroupParserConfig {
	cfg.parsers = append(cfg.parsers, builders...)
	return cfg
}

func (cfg *GroupParserConfig) ParsePrefix(
	prefix string, builder GroupParserBuilder[node.Node],
) *GroupParserConfig {
	cfg.prefixes[prefix] = builder
	return cfg
}

func (cfg *GroupParserConfig) parser(
	parseAlternation c.Combinator[rune, int, node.Alternation],
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parseAny := c.NoneOf[rune, int](except...) // to parse prefix rune by rune

	parseScopeByPrefix := NewGroupsParserTree(
		parseAny,
		parseAlternation,
		cfg.prefixes,
		except...,
	)

	parsers := make([]c.Combinator[rune, int, node.Node], 0, len(cfg.parsers)+1)
	parsers = append(parsers, c.Try(parseScopeByPrefix))

	for _, buildParser := range cfg.parsers {
		parser := buildParser(parseAlternation, except...)
		parsers = append(parsers, c.Try(parser))
	}

	return c.Choice(parsers...)
}

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

	variants := alt.GetVariants()
	if len(variants) == 1 {
		newNode = variants[0]
	}

	// NOTE: translate one type node to another -> \d{3} -> \d\d\d ?
	// NOTE: add specail error class with merge method
	// (required to pretty errors (expected: char or class ... explanation))
	// and merge expectations in choice

	newNode.Traverse(func(x node.Node) {
		if len(x.GetNestedNodes()) == 0 { // its leaf
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

			return node.NewForTable(unicode.NewTableFor(x)), nil
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
		fmt.Println("parse range", buf)

		from, err := parseRune(buf)
		if err != nil {
			fmt.Println("failed", err, buf)
			return nil, err
		}

		fmt.Println("parse range from : ", from, buf)
		pos := buf.Position()

		_, err = parseSeparator(buf)
		if err != nil {
			fmt.Println("failed", err, buf, from)
			buf.Seek(pos)
			return unicode.NewTableFor(from), nil
		}

		to, err := parseRune(buf)
		if err != nil {
			fmt.Println("failed", err, buf, to)
			buf.Seek(pos)
			return unicode.NewTableFor(from), nil
		}

		fmt.Println("parse range to : ", to, buf)

		// TODO : check bounds

		return unicode.NewTableByPredicate(func(x rune) bool {
			return from <= x && x <= to
		}), nil
	}
}

func (p *CustomParser) classParser(
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parseClass := p.classTableParser(except...)

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		table, err := parseClass(buf)
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
	parseQuantity := c.Try(p.quantityParser(except...))

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

func (p *CustomParser) quantityParser(
	except ...rune,
) c.Combinator[rune, int, *node.Quantity] {
	any := c.NoneOf[rune, int](except...)
	digit := c.Try(Number())
	comma := c.Try(c.Eq[rune, int](','))
	rightBrace := c.Eq[rune, int]('}')

	return c.MapAs(
		map[rune]c.Combinator[rune, int, *node.Quantity]{
			'?': func(buf c.Buffer[rune, int]) (*node.Quantity, error) {
				return node.NewQuantity(0, 1), nil
			},
			'+': func(buf c.Buffer[rune, int]) (*node.Quantity, error) {
				return node.NewEndlessQuantity(1), nil
			},
			'*': func(buf c.Buffer[rune, int]) (*node.Quantity, error) {
				return node.NewEndlessQuantity(0), nil
			},
			'{': c.Choice(
				c.Try(func(buf c.Buffer[rune, int]) (*node.Quantity, error) { // {1,1}
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

					return node.NewQuantity(from, to), nil
				}),
				c.Try(func(buf c.Buffer[rune, int]) (*node.Quantity, error) { // {,1}
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

					return node.NewQuantity(0, to), nil
				}),
				c.Try(func(buf c.Buffer[rune, int]) (*node.Quantity, error) { // {1,}
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

					return node.NewEndlessQuantity(from), nil
				}),
				func(buf c.Buffer[rune, int]) (*node.Quantity, error) { // {1}
					from, err := digit(buf)
					if err != nil {
						return nil, err
					}

					_, err = rightBrace(buf)
					if err != nil {
						return nil, err
					}

					return node.NewQuantity(from, from), nil
				},
			),
		},
		any,
	)
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
