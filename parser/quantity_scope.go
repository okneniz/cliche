package parser

import (
	"fmt"

	"github.com/okneniz/cliche/quantity"
)

type QuantityScope struct {
	items *ScopeConfig[*quantity.Quantity]
}

func (cfg *QuantityScope) Items() *ScopeConfig[*quantity.Quantity] {
	return cfg.items
}

func (cfg *QuantityScope) String() string {
	return fmt.Sprintf("%T{%v}", cfg, cfg.items)
}
