package parser

import (
	"github.com/okneniz/cliche/quantity"
)

type QuantityScope struct {
	items *ScopeConfig[*quantity.Quantity]
}

func (cfg *QuantityScope) Items() *ScopeConfig[*quantity.Quantity] {
	return cfg.items
}
