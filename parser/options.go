package parser

import (
	"github.com/okneniz/cliche/encoding/unicode"
	"github.com/okneniz/cliche/node"
	c "github.com/okneniz/parsec/common"
)

func WithBracket(
	name string, predicate func(rune) bool,
) Option[*CustomParser] {
	// TODO : validate name to avoid conflicts with default spec symbols ".?+*^$[]{}()"

	table := unicode.NewUnicodeTableByPredicate(predicate)
	negatiatedTable := table.Invert()

	parseNode := func(buf c.Buffer[rune, int]) (node.Node, error) {
		return node.NewForTable(table), nil
	}

	parseNegatedNode := func(buf c.Buffer[rune, int]) (node.Node, error) {
		return node.NewForTable(negatiatedTable), nil
	}

	parseTable := func(buf c.Buffer[rune, int]) (node.Table, error) {
		return table, nil
	}

	parseNegatedTable := func(buf c.Buffer[rune, int]) (node.Table, error) {
		return negatiatedTable, nil
	}

	return func(parser *CustomParser) {
		parser.prefixes["[[:"+name+":]]"] = parseNode
		parser.prefixes["[[:^"+name+":]]"] = parseNegatedNode

		parser.inClassPrefixes["[[:"+name+":]]"] = parseTable
		parser.inClassPrefixes["[[:^"+name+":]]"] = parseNegatedTable
	}
}

// TODO : pass exceptions too
func WithEscapedMetaChar(
	name string, predicate func(rune) bool,
) Option[*CustomParser] {
	// TODO : validate char

	table := unicode.NewUnicodeTableByPredicate(predicate)
	parse := func(buf c.Buffer[rune, int]) (node.Node, error) {
		return node.NewForTable(table), nil
	}
	parseTable := func(buf c.Buffer[rune, int]) (node.Table, error) {
		return table, nil
	}

	return func(parser *CustomParser) {
		parser.prefixes["\\"+name] = parse
		parser.inClassPrefixes["\\"+name] = parseTable
	}
}

// TODO : pass exceptions too
func WithPrefix(
	name string, parse c.Combinator[rune, int, node.Node],
) Option[*CustomParser] {
	return func(parser *CustomParser) {
		// TODO : validate name
		parser.prefixes[name] = parse
	}
}

// TODO : pass exceptions too
func WithInClassPrefix(
	name string, parse c.Combinator[rune, int, node.Table],
) Option[*CustomParser] {
	return func(parser *CustomParser) {
		// TODO : validate name
		parser.inClassPrefixes["\\"+name] = parse
	}
}

// TODO : pass exceptions too
func WithParser(
	p func(except ...rune) c.Combinator[rune, int, node.Node],
) Option[*CustomParser] {
	return func(parser *CustomParser) {
		parser.parsers = append(parser.parsers, p)
	}
}

// TODO : pass exceptions too
func WithInClassParser(
	p func(except ...rune) c.Combinator[rune, int, node.Table],
) Option[*CustomParser] {
	return func(parser *CustomParser) {
		parser.inClassParsers = append(parser.inClassParsers, p)
	}
}
