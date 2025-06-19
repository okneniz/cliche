package parser

import (
	"github.com/okneniz/cliche/node"
)

type ParserConfig struct {
	nonClassConfig *NonClassParserConfig
	groupConfig    *GroupParserConfig
	classConfig    *ClassParserConfig
	quntityConfig  *QuantityParserConfig
}

func NewConfig() *ParserConfig {
	cfg := new(ParserConfig)

	cfg.nonClassConfig = new(NonClassParserConfig)
	cfg.nonClassConfig.items = NewParserScope[node.Node]()

	cfg.groupConfig = new(GroupParserConfig)
	cfg.groupConfig.prefixes = make(map[string]GroupParserBuilder[node.Node], 0)
	cfg.groupConfig.parsers = make([]GroupParserBuilder[node.Node], 0)

	cfg.classConfig = new(ClassParserConfig)
	cfg.classConfig.runes = NewParserScope[rune]()
	cfg.classConfig.items = NewParserScope[node.Table]()

	cfg.quntityConfig = new(QuantityParserConfig)
	cfg.quntityConfig.items = NewParserScope[*node.Quantity]()

	return cfg
}

func (cfg *ParserConfig) Groups() *GroupParserConfig {
	return cfg.groupConfig
}

func (cfg *ParserConfig) Class() *ClassParserConfig {
	return cfg.classConfig
}

func (cfg *ParserConfig) NonClass() *NonClassParserConfig {
	return cfg.nonClassConfig
}

func (cfg *ParserConfig) Quntifier() *QuantityParserConfig {
	return cfg.quntityConfig
}
