package parser

import (
	"github.com/okneniz/cliche/node"
)

type NonClassScope struct {
	// escaped char (\u{123}, \A, \z)
	// escaped range of char (\d, \w, \p{Property})
	items *ScopeConfig[node.Node]
}

func (cfg *NonClassScope) Items() *ScopeConfig[node.Node] {
	return cfg.items
}
