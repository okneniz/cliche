package node

import (
	"github.com/okneniz/cliche/quantity"
	"github.com/okneniz/cliche/structs"
)

type Node interface {
	GetKey() string
	GetExpressions() structs.Set[string]
	AddExpression(string)
	GetNestedNodes() map[string]Node
	IsLeaf() bool

	Visit(Scanner, Input, int, int, Callback)

	// TODO : works only for fixed chain with one end node?
	// don't work for tree?

	// TODO : it's improtant for group too have chain in Value instead tree
	// make special type for this case?

	// TODO : what about alternation of chains?
	// iterate over all sizes in asserts (lookaheads / lookbehinds)

	// TODO : trie have few size / heights (leafs)
	// TODO : node with different sizes must have chain value instead trie to simpliy this moment
	Size() (int, bool)

	// TODO : what about anchors, is it endless or zero sized?

	Copy() Node
}

type Callback func(x Node, from int, to int, empty bool)

type Alternation interface {
	Node

	GetVariants() []Node

	VisitAlternation(
		scanner Scanner,
		input Input,
		from, to int,
		match AlternationCallback,
	)

	CopyAlternation() Alternation
}

type Container interface {
	Node

	GetValue() Node
}

type AlternationCallback func(x Node, from int, to int, empty bool) (stop bool)

type Table interface {
	Include(rune) bool
	Invert(rune) Table
	Empty() bool
	String() string
}

type Scanner interface {
	Position() int
	Rewind(pos int)

	MatchGroup(from int, to int)
	GroupsPosition() int
	GetGroup(idx int) (quantity.Interface, bool)
	RewindGroups(pos int)

	MatchNamedGroup(name string, from int, to int)
	NamedGroupsPosition() int
	GetNamedGroup(name string) (quantity.Interface, bool)
	RewindNamedGroups(pos int)

	MarkAsHole(from int, to int)
	HolesPosition() int
	RewindHoles(pos int)

	OptionsInclude(opt ScanOption) bool
	OptionsEnable(opt ScanOption)
	OptionsDisable(opt ScanOption)
	OptionsPosition() int
	RewindOptions(pos int)
}

type ScanOption uint

const (
	ScanOptionCaseInsensetive    ScanOption = 1
	ScanOptionMultiline          ScanOption = 2
	ScanOptionDisableNamedGroups ScanOption = 3
)

type Input interface {
	ReadAt(int) rune
	Size() int
	Position() int
}

type Output interface {
	Yield(
		n Node,
		subString string,
		sp quantity.Interface,
		groups []quantity.Interface,
		namedGroups map[string]quantity.Interface,
	)

	LastPosOf(n Node) (int, bool)
	String() string
}

func Traverse(n Node, f func(Node) bool) {
	if stop := f(n); stop {
		return
	}

	for _, nestedNode := range n.GetNestedNodes() {
		Traverse(nestedNode, f)
	}
}

func nextFor(pos int, empty bool) int {
	if empty {
		return pos
	}

	return pos + 1
}
