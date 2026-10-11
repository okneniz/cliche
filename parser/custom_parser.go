package parser

import (
	"slices"

	"github.com/okneniz/cliche/buf"
	"github.com/okneniz/cliche/node"
	"github.com/okneniz/parsec"
)

type State struct {
	GroupIndex int
}

type CustomParser struct {
	config *Config
	parse  parsec.Combinator[rune, int, node.Alternation, State]
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
	state := State{}

	alt, err := p.parse(state, buffer)
	if err != nil {
		return nil, err
	}

	// если добавить state в parsec то можно удалить вообще?
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
) parsec.Combinator[rune, int, node.Alternation, State] {
	parseBar := parsec.Eq[rune, int, State]("expected '|' as alternation separator", '|')
	parseLeftParens := parsec.Eq[rune, int, State]("expected left parens as beginning of group", '(')
	parseRightParens := parsec.Eq[rune, int, State]("expected right parent as ending of group", ')')

	parseClass := parsec.Try(p.config.class.makeParser())
	parseNonClass := parsec.Try(p.config.nonClass.makeParser(
		"expected non class item",
		except...,
	))

	var (
		parseGroup parsec.Combinator[rune, int, node.Node, State]
	)

	// can't use parsec.Choice because parseGroup is var
	// and not initiated at this moment to pass as function param
	parseNode := p.makeOptionalQuantifierParser(
		func(state State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
			start := buf.Position()

			group, groupErr := parseGroup(state, buf)
			if groupErr == nil {
				return group, nil
			}

			class, classErr := parseClass(state, buf)
			if classErr == nil {
				return class, nil
			}

			nonClass, nonClassErr := parseNonClass(state, buf)
			if nonClassErr == nil {
				return nonClass, nil
			}

			return nil, parsec.NewParseError(
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

	parseVariants := parsec.SepBy1(
		1,
		"expected alternation variant",
		parseVariant,
		parseBar,
	)

	parseAlternation := func(
		state State,
		buf parsec.Buffer[rune, int],
	) (node.Alternation, parsec.Error[int]) {
		variants, err := parseVariants(state, buf)
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

	parseGroup = parsec.Try(
		parsec.Between(
			parseLeftParens,
			parseGroupValue,
			parseRightParens,
		),
	)

	return parseAlternation
}

func (p *CustomParser) makeOptionalQuantifierParser(
	parseExpression parsec.Combinator[rune, int, node.Node, State],
	except ...rune,
) parsec.Combinator[rune, int, node.Node, State] {
	parseQuantity := parsec.Try(p.config.quantity.makeParser(
		"expected quantifier",
		except...,
	))

	return func(state State, buf parsec.Buffer[rune, int]) (node.Node, parsec.Error[int]) {
		expression, err := parseExpression(state, buf)
		if err != nil {
			return nil, err
		}

		quantity, qErr := parseQuantity(state, buf)
		if qErr != nil {
			return expression, nil
		}

		return node.NewQuantifier(quantity, expression), nil
	}
}

func (p *CustomParser) makeChainParser(
	parse parsec.Combinator[rune, int, node.Node, State],
) parsec.Combinator[rune, int, node.Node, State] {
	makeChain := parsec.Const[rune, int, parsec.BinaryOp[node.Node], State](
		func(current, next node.Node) node.Node {
			current.GetNestedNodes()[next.GetKey()] = next
			return current
		},
	)

	return parsec.Chainr1(parse, makeChain)
}
