package parser

import (
	"fmt"

	"github.com/okneniz/cliche/node"
)

// as in NonClass, but some char have another meaning (for example $, ^)
// don't have anchors
type ClassConfig struct {
	// \u{00E0}, \A, \z
	runes *ParserScope[rune]
	items *ParserScope[node.Table]
}

func (cfg *ClassConfig) Runes() *ParserScope[rune] {
	return cfg.runes
}

func (cfg *ClassConfig) Items() *ParserScope[node.Table] {
	return cfg.items
}

func (cfg *ClassConfig) String() string {
	return fmt.Sprintf("%T{%v}", cfg, cfg.items)
}
