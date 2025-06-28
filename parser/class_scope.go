package parser

import (
	"github.com/okneniz/cliche/node"
)

// as in NonClass, but some char have another meaning (for example $, ^)
// don't have anchors
type ClassScope struct {
	// \u{00E0}, \A, \z
	runes *ScopeConfig[rune]
	items *ScopeConfig[node.Table]
}

func (cfg *ClassScope) Runes() *ScopeConfig[rune] {
	return cfg.runes
}

func (cfg *ClassScope) Items() *ScopeConfig[node.Table] {
	return cfg.items
}
