package parser

import (
	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/quantity"
)

type Config struct {
	nonClass *NonClassScope
	group    *GroupScope
	class    *ClassScope
	quntity  *QuantityScope
}

func NewConfig() *Config {
	cfg := new(Config)

	cfg.nonClass = new(NonClassScope)
	cfg.nonClass.items = NewScopeConfig[node.Node]()

	cfg.group = new(GroupScope)
	cfg.group.prefixes = make(map[string]GroupParserBuilder[node.Node], 0)
	cfg.group.parsers = make([]GroupParserBuilder[node.Node], 0)

	cfg.class = new(ClassScope)
	cfg.class.runes = NewScopeConfig[rune]()
	cfg.class.items = NewScopeConfig[node.Table]()

	cfg.quntity = new(QuantityScope)
	cfg.quntity.items = NewScopeConfig[*quantity.Quantity]()

	return cfg
}

func (cfg *Config) Groups() *GroupScope {
	return cfg.group
}

func (cfg *Config) Class() *ClassScope {
	return cfg.class
}

func (cfg *Config) NonClass() *NonClassScope {
	return cfg.nonClass
}

func (cfg *Config) Quntifier() *QuantityScope {
	return cfg.quntity
}
