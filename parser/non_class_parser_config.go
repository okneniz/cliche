package parser

import (
	"github.com/okneniz/cliche/node"
)

type NonClassParserConfig struct {
	// escaped char (\u{123}, \A, \z)
	// escaped range of char (\d, \w, \p{Property})
	items *ParserScope[node.Node]
}

func (cfg *NonClassParserConfig) Items() *ParserScope[node.Node] {
	return cfg.items
}
