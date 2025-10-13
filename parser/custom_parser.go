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
	p.parse = p.makeAlternationParser('|')

	return p
}

func (p *CustomParser) Parse(str string) (node.Alternation, error) {
	buffer := buf.NewRunesBuffer(str)

	alt, err := p.parse(buffer)
	if err != nil {
		return nil, err
	}

	node.Traverse(alt, func(x node.Node) bool {
		if len(x.GetNestedNodes()) == 0 {
			x.AddExpression(str)
		}

		return false
	})

	return alt, nil
}

func (p *CustomParser) makeAlternationParser(
	except ...rune,
) c.Combinator[rune, int, node.Alternation] {
	parseBar := c.Eq[rune, int]("expected '|' as alternation separator", '|')
	parseLeftParens := c.Eq[rune, int]("expected left parens as begining of group", '(')
	parseRightParens := c.Eq[rune, int]("expected right parent as ending of group", ')')

	parseClass := c.Try(p.config.class.makeParser())
	parseNonClass := c.Try(p.config.nonClass.makeParser(
		"expected non class item",
		except...,
	))

	var (
		parseGroup c.Combinator[rune, int, node.Node]
	)

	// can't use parsec/common.Choice because parseGroup is var
	// and not initiated at this moment to pass as function param
	parseNode := p.makeOptionalQuantifierParser(
		func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
			start := buf.Position()

			group, groupErr := parseGroup(buf)
			if groupErr == nil {
				return group, nil
			}

			class, classErr := parseClass(buf)
			if classErr == nil {
				return class, nil
			}

			nonClass, nonClassErr := parseNonClass(buf)
			if nonClassErr == nil {
				return nonClass, nil
			}

			return nil, c.NewParseError(
				start,
				"expected item for alternation",
				groupErr,
				classErr,
				nonClassErr,
			)
		},
		except...,
	)

	parseVariant := p.makeChainParser(parseNode)

	parseVariants := c.SepBy1(
		1,
		"expected alternation variant",
		parseVariant,
		parseBar,
	)

	parseAlternation := func(
		buf c.Buffer[rune, int],
	) (node.Alternation, c.Error[int]) {
		variants, err := parseVariants(buf)
		if err != nil {
			return nil, err
		}

		return node.NewAlternation(variants), nil
	}

	groupAlternation := parseAlternation
	if !slices.Contains(except, ')') {
		groupAlternation = p.makeAlternationParser(append(except, ')')...)
	}

	parseGroupValue := p.config.group.makeParser(
		groupAlternation,
		append(except, ')')...,
	)

	parseGroup = c.Try(
		c.Between(
			parseLeftParens,
			parseGroupValue,
			parseRightParens,
		),
	)

	return parseAlternation
}

func (p *CustomParser) makeOptionalQuantifierParser(
	parseExpression c.Combinator[rune, int, node.Node],
	except ...rune,
) c.Combinator[rune, int, node.Node] {
	parseQuantity := c.Try(p.config.quantity.makeParser(
		"expected quantifier",
		except...,
	))

	return func(buf c.Buffer[rune, int]) (node.Node, c.Error[int]) {
		expression, err := parseExpression(buf)
		if err != nil {
			return nil, err
		}

		quantity, qErr := parseQuantity(buf)
		if qErr != nil {
			return expression, nil
		}

		return node.NewQuantifier(quantity, expression), nil
	}
}

func (p *CustomParser) makeChainParser(
	parse c.Combinator[rune, int, node.Node],
) c.Combinator[rune, int, node.Node] {
	makeChain := c.Const[rune, int, c.BinaryOp[node.Node]](
		func(current, next node.Node) node.Node {
			current.GetNestedNodes()[next.GetKey()] = next
			return current
		},
	)

	return c.Chainr1(parse, makeChain)
}
