package parser

import (
	"fmt"

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

func (p *CustomParser) makeAlternationParser(
	except ...rune,
) Parser[node.Alternation] {
	parseClass := p.config.class.makeParser()
	parseNonClass := p.config.nonClass.makeParser(except...)

	var (
		parseGroup Parser[node.Node]
	)

	parseNode := p.makeOptionalQuantifierParser(
		func(buf c.Buffer[rune, int]) (node.Node, *ParsingError) {
			group, groupErr := parseGroup(buf)
			if groupErr == nil {
				fmt.Println("parsed node group", group.GetKey(), group.GetExpressions())
				return group, nil
			}

			fmt.Println("parsing node group failed:", groupErr)

			class, classErr := parseClass(buf)
			if classErr == nil {
				fmt.Println("parsed node class", class.GetKey(), class.GetExpressions())
				return class, nil
			}

			fmt.Println("parsing node class failed:", classErr)

			nonClass, nonClassErr := parseNonClass(buf)
			if nonClassErr == nil {
				fmt.Println("parsed node non class", nonClass.GetKey(), nonClass.GetExpressions())
				return nonClass, nil
			}

			fmt.Println("parsing node non class failed:", nonClassErr)

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
	) ([]node.Node, *ParsingError) {
		pos := buf.Position()

		variant, err := parseVariant(buf)
		if err != nil {
			buf.Seek(pos)
			return nil, err
		}

		pos = buf.Position()
		fmt.Println("parsed variant", variant.GetKey(), variant.GetExpressions())

		list := make([]node.Node, 1)
		list[0] = variant

		i := 1

		fmt.Println("try to parse more than one variant")
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

			i += 1
			fmt.Println("parsed next variant", i, variant.GetKey(), variant.GetExpressions())
			pos = buf.Position()

			list = append(list, variant)
		}

		return list, nil
	}

	parseAlternation := func(
		buf c.Buffer[rune, int],
	) (node.Alternation, *ParsingError) {
		variants, err := parseVariants(buf)
		if err != nil {
			fmt.Println("parsing alternation failed", err)
			return nil, err
		}

		fmt.Println("parsed alternation with", len(variants), "variants", buf)

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
	) (node.Node, *ParsingError) {
		pos := buf.Position()

		_, err := leftParens(buf)
		if err != nil {
			fmt.Println("parsing left parens of group failed", err)
			buf.Seek(pos)
			return nil, err
		}

		fmt.Println("left parens parsed")

		value, gErr := parseGroupValue(buf)
		if gErr != nil {
			buf.Seek(pos)
			return nil, gErr
		}

		fmt.Println("group value parsed", value, err, buf)

		_, err = rightParens(buf)
		if err != nil {
			fmt.Println("parsing right parens of group failed", err)
			buf.Seek(pos)
			return nil, err
		}

		fmt.Println("group parsed", value.GetKey(), value.GetExpressions())

		return value, nil
	}

	return parseAlternation
}

func (p *CustomParser) makeOptionalQuantifierParser(
	expression Parser[node.Node],
	except ...rune,
) Parser[node.Node] {
	parseQuantity := p.config.quantity.makeParser(except...)

	return func(buf c.Buffer[rune, int]) (node.Node, *ParsingError) {
		exp, err := expression(buf)
		if err != nil {
			return nil, err
		}

		fmt.Println("try to parse quantity", buf)

		quantity, qErr := parseQuantity(buf)
		if qErr != nil {
			fmt.Println("quantificator parsing failed:", qErr)
			return exp, nil
		}

		fmt.Println("success", quantity)

		return node.NewQuantifier(quantity, exp), nil
	}
}

func (p *CustomParser) makeChainParser(parse Parser[node.Node]) Parser[node.Node] {
	return func(buf c.Buffer[rune, int]) (node.Node, *ParsingError) {
		first, err := parse(buf)
		if err != nil {
			return nil, err
		}

		fmt.Println("in chain", first.GetKey(), first.GetExpressions())

		last := first

		for !buf.IsEOF() {
			pos := buf.Position()
			fmt.Println("parse chain at", buf.Position())

			next, err := parse(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			last.GetNestedNodes()[next.GetKey()] = next
			last = next
		}

		fmt.Println("return chain", buf, last.GetKey(), last.GetExpressions())

		return first, nil
	}
}
