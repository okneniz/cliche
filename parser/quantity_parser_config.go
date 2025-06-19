package parser

import (
	"fmt"

	"github.com/okneniz/cliche/node"
)

type QuantityParserConfig struct {
	items *ParserScope[*node.Quantity]
}

func (cfg *QuantityParserConfig) Items() *ParserScope[*node.Quantity] {
	return cfg.items
}

func (cfg *QuantityParserConfig) String() string {
	return fmt.Sprintf("%T{%v}", cfg, cfg.items)
}
