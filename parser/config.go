package parser

import (
	"github.com/okneniz/cliche/node"
)

type Config struct {
	nonClassConfig *NonClassConfig
	groupConfig    *GroupConfig
	classConfig    *ClassConfig
	quntityConfig  *QuantityConfig
}

func NewConfig() *Config {
	cfg := new(Config)

	cfg.nonClassConfig = new(NonClassConfig)
	cfg.nonClassConfig.items = NewParserScope[node.Node]()

	cfg.groupConfig = new(GroupConfig)
	cfg.groupConfig.prefixes = make(map[string]GroupParserBuilder[node.Node], 0)
	cfg.groupConfig.parsers = make([]GroupParserBuilder[node.Node], 0)

	cfg.classConfig = new(ClassConfig)
	cfg.classConfig.runes = NewParserScope[rune]()
	cfg.classConfig.items = NewParserScope[node.Table]()

	cfg.quntityConfig = new(QuantityConfig)
	cfg.quntityConfig.items = NewParserScope[*node.Quantity]()

	return cfg
}

func (cfg *Config) Groups() *GroupConfig {
	return cfg.groupConfig
}

func (cfg *Config) Class() *ClassConfig {
	return cfg.classConfig
}

func (cfg *Config) NonClass() *NonClassConfig {
	return cfg.nonClassConfig
}

func (cfg *Config) Quntifier() *QuantityConfig {
	return cfg.quntityConfig
}
