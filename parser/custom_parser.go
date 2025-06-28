package parser

import (
	"golang.org/x/exp/slices"

	"github.com/okneniz/cliche/buf"
	"github.com/okneniz/cliche/node"
	c "github.com/okneniz/parsec/common"
)

type CustomParser struct {
	config *Config
	parse  c.Combinator[rune, int, node.Alternation]
}

type Option[T any] func(T)

func New(opts ...Option[*Config]) *CustomParser {
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

	// TODO : move it to special component (alterer)
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
	parseClass := p.config.class.makeParser()
	parseNonClass := p.config.nonClass.makeParser(except...)

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
	parseVariants := c.SepBy1(0, parseVariant, parseSeparator)

	parseAlternation := func(buf c.Buffer[rune, int]) (node.Alternation, error) {
		variants, err := parseVariants(buf)
		if err != nil {
			return nil, err
		}

		return node.NewAlternation(variants), nil
	}

	groupAlternation := parseAlternation
	if !slices.Contains(except, ')') {
		groupAlternation = p.alternationParser(append(except, ')')...)
	}

	parseGroup = c.Try(
		Parens(
			p.config.group.makeParser(
				groupAlternation,
				append(except, ')')...,
			),
		),
	)

	return parseAlternation
}

func (p *CustomParser) optionalQuantifierParser(
	expression c.Combinator[rune, int, node.Node],
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parseQuantity := c.Try(p.config.quntity.items.makeParser(except...))

	return func(buf c.Buffer[rune, int]) (node.Node, error) {
		exp, err := expression(buf)
		if err != nil {
			return nil, err
		}

		quantity, err := parseQuantity(buf)
		if err != nil {
			return exp, nil
		}

		return node.NewQuantifier(quantity, exp), nil
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
