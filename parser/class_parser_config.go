package parser

import (
	"fmt"

	"github.com/okneniz/cliche/node"
)

// as in NonClass, but some char have another meaning (for example $, ^)
// don't have anchors
type ClassParserConfig struct {
	// \u{00E0}, \A, \z
	runes *ParserScope[rune]
	items *ParserScope[node.Table]
}

func (cfg *ClassParserConfig) Runes() *ParserScope[rune] {
	return cfg.runes
}

func (cfg *ClassParserConfig) Items() *ParserScope[node.Table] {
	return cfg.items
}

func (cfg *ClassParserConfig) String() string {
	return fmt.Sprintf("%T{%v}", cfg, cfg.items)
}
