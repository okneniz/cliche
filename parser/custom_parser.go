package parser

import (
	"golang.org/x/exp/slices"

	"github.com/okneniz/cliche/buf"
	"github.com/okneniz/cliche/node"
	c "github.com/okneniz/parsec/common"
)

type CustomParser struct {
	config *Config
	parse  Parser[node.Alternation]
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

	newNode.Traverse(func(x node.Node) {
		if len(x.GetNestedNodes()) == 0 {
			x.AddExpression(str)
		}
	})

	return newNode, nil
}

func (p *CustomParser) makeAlternationParser(
	except ...rune,
) Parser[node.Alternation] {
	parseClass := p.config.class.makeParser()
	parseNonClass := p.config.nonClass.makeParser(except...)

	var (
		parseGroup Parser[node.Node]
	)

	parseNode := p.makeOptionalQuantifierParser(
		func(buf c.Buffer[rune, int]) (node.Node, Error) {
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

			return nil, MergeErrors(
				groupErr,
				classErr,
				nonClassErr,
			)
		},
		except...,
	)

	parseVariant := p.makeChainParser(parseNode)
	parseSeparator := Eq('|')

	parseVariants := func(
		buf c.Buffer[rune, int],
	) ([]node.Node, Error) {
		pos := buf.Position()

		variant, err := parseVariant(buf)
		if err != nil {
			buf.Seek(pos)
			return nil, err
		}

		pos = buf.Position()

		list := make([]node.Node, 1)
		list[0] = variant

		for {
			_, sepErr := parseSeparator(buf)
			if sepErr != nil {
				buf.Seek(pos)
				break
			}

			variant, err := parseVariant(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			pos = buf.Position()

			list = append(list, variant)
		}

		return list, nil
	}

	parseAlternation := func(
		buf c.Buffer[rune, int],
	) (node.Alternation, Error) {
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

	leftParens := Eq('(')
	rightParens := Eq(')')

	parseGroup = func(
		buf c.Buffer[rune, int],
	) (node.Node, Error) {
		pos := buf.Position()

		_, leftErr := leftParens(buf)
		if leftErr != nil {
			buf.Seek(pos)
			return nil, leftErr
		}

		value, gErr := parseGroupValue(buf)
		if gErr != nil {
			buf.Seek(pos)
			return nil, gErr
		}

		_, rightErr := rightParens(buf)
		if rightErr != nil {
			buf.Seek(pos)
			return nil, rightErr
		}

		return value, nil
	}

	return parseAlternation
}

func (p *CustomParser) makeOptionalQuantifierParser(
	expression Parser[node.Node],
	except ...rune,
) Parser[node.Node] {
	parseQuantity := p.config.quantity.makeParser(except...)

	return func(buf c.Buffer[rune, int]) (node.Node, Error) {
		exp, err := expression(buf)
		if err != nil {
			return nil, err
		}

		quantity, qErr := parseQuantity(buf)
		if qErr != nil {
			return exp, nil
		}

		return node.NewQuantifier(quantity, exp), nil
	}
}

func (p *CustomParser) makeChainParser(parse Parser[node.Node]) Parser[node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, Error) {
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
