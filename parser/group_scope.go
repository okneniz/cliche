package parser

import (
	"github.com/okneniz/cliche/node"
	"github.com/okneniz/parsec"
)

type GroupParserBuilder[T any] func(
	parseAlternation parsec.Combinator[rune, int, node.Alternation, *State],
	except ...rune,
) parsec.Combinator[rune, int, T, *State]

type GroupScope struct {
	prefixes map[string]GroupParserBuilder[node.Node]
	parsers  []GroupParserBuilder[node.Node] // alternation wrapped to node
}

func (cfg *GroupScope) Parse(
	builders ...GroupParserBuilder[node.Node],
) *GroupScope {
	cfg.parsers = append(cfg.parsers, builders...)
	return cfg
}

func (cfg *GroupScope) ParsePrefix(
	prefix string, builder GroupParserBuilder[node.Node],
) *GroupScope {
	cfg.prefixes[prefix] = builder
	return cfg
}

func (cfg *GroupScope) makeParser(
	parseAlternation parsec.Combinator[rune, int, node.Alternation, *State],
	except ...rune,
) parsec.Combinator[rune, int, node.Node, *State] {
	parseScopeByPrefix := makeGroupsParserTree(
		parseAlternation,
		cfg.prefixes,
		except...,
	)

	// TODO : вот где-то тут получается должен быть counter для груп
	// сейчас непонятно стоит ли заморачиваться с thread safe и прочим
	//
	// если состояние каждый раз новое при вызове Parse то наверно невозможно его как-то
	// вообще разделить между потоками
	//
	// пользовательские комбинаторы будут получаеть счетчик как параметр
	// без возможности что-то сделать еще

	// получается при парсинге нужно запоминать состояние на случай если c.Try не смог распарсить
	// чтобы счетчик не рос на каждую попытку

	// архитектурыный вопрос - стоит ли делать состояние только тут или оно понадобится еще и
	// в других местах

	// стоит ли давать пользователю возможность делать statefull комбинаторы

	// еще вопрос - стоит ли усложнять состояние
	// что-то кроме груп вообще будет?
	//
	// достаточно одной цифры?
	//
	// или нужна структура с интерфейсом и тд

	parsers := make([]parsec.Combinator[rune, int, node.Node, *State], 0, len(cfg.parsers)+1)
	parsers = append(parsers, parsec.Try(parseScopeByPrefix))

	for _, buildParser := range cfg.parsers {
		parser := buildParser(parseAlternation, except...)
		parsers = append(parsers, parsec.Try(parser))
	}

	return parsec.Choice(
		"expected group",
		parsers...,
	)
}
