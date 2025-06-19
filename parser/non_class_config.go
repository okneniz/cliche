package parser

import (
	"github.com/okneniz/cliche/node"
)

type NonClassConfig struct {
	// escaped char (\u{123}, \A, \z)
	// escaped range of char (\d, \w, \p{Property})
	items *ParserScope[node.Node]
}

func (cfg *NonClassConfig) Items() *ParserScope[node.Node] {
	return cfg.items
}
