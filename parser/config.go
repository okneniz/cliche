package parser

import (
	"github.com/okneniz/cliche/node"
)

type Config struct {
	nonClass *NonClassConfig
	group    *GroupConfig
	class    *ClassConfig
	quntity  *QuantityConfig
}

func NewConfig() *Config {
	cfg := new(Config)

	cfg.nonClass = new(NonClassConfig)
	cfg.nonClass.items = NewParserScope[node.Node]()

	cfg.group = new(GroupConfig)
	cfg.group.prefixes = make(map[string]GroupParserBuilder[node.Node], 0)
	cfg.group.parsers = make([]GroupParserBuilder[node.Node], 0)

	cfg.class = new(ClassConfig)
	cfg.class.runes = NewParserScope[rune]()
	cfg.class.items = NewParserScope[node.Table]()

	cfg.quntity = new(QuantityConfig)
	cfg.quntity.items = NewParserScope[*node.Quantity]()

	return cfg
}

func (cfg *Config) Groups() *GroupConfig {
	return cfg.group
}

func (cfg *Config) Class() *ClassConfig {
	return cfg.class
}

func (cfg *Config) NonClass() *NonClassConfig {
	return cfg.nonClass
}

func (cfg *Config) Quntifier() *QuantityConfig {
	return cfg.quntity
}
