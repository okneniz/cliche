package parser

import (
	"fmt"

	"github.com/okneniz/cliche/node"
)

type QuantityScope struct {
	items *ScopeConfig[*node.Quantity]
}

func (cfg *QuantityScope) Items() *ScopeConfig[*node.Quantity] {
	return cfg.items
}

func (cfg *QuantityScope) String() string {
	return fmt.Sprintf("%T{%v}", cfg, cfg.items)
}
