package parser

import (
	"fmt"

	"github.com/okneniz/cliche/node"
)

type QuantityConfig struct {
	items *ParserScope[*node.Quantity]
}

func (cfg *QuantityConfig) Items() *ParserScope[*node.Quantity] {
	return cfg.items
}

func (cfg *QuantityConfig) String() string {
	return fmt.Sprintf("%T{%v}", cfg, cfg.items)
}
