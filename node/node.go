package node

import (
	"github.com/okneniz/cliche/span" // must not be deps?
	"github.com/okneniz/cliche/structs"
)

type Node interface {
	GetKey() string
	GetExpressions() structs.Set[string]
	AddExpression(string)
	GetNestedNodes() map[string]Node
	IsLeaf() bool

	Visit(Scanner, Input, int, int, Callback)

	// Add parent to travers
	// Should return bool to interupt traversing?
	Traverse(func(Node))

	// TODO : works only for fixed chain with one end node?
	// don't work for tree?

	// TODO : it's improtant for group too have chain in Value instead tree
	// make special type for this case?

	// TODO : what about alternation of chains?
	// Iterate over all sizes in asserts (lookaheads / lookbehinds)

	// TODO : trie have few size / heights (leafs)
	// TODO : node with different sizes must have chain value instead trie to simpliy this moment
	Size() (int, bool)

	// TODO : what about anchors, is it endless or zero sized?
}

type Callback func(x Node, from int, to int, empty bool)

type Alternation interface {
	Node

	VisitAlternation(
		scanner Scanner,
		input Input,
		from, to int,
		onMatch AlternationCallback,
	)
}

type AlternationCallback func(x Node, from int, to int, empty bool) (stop bool)

type Table interface {
	Include(rune) bool
	Invert() Table
	String() string
}

type Scanner interface {
	Match(n Node, from, to int, isLeaf, isEmpty bool)
	Position() int
	Rewind(pos int)

	MatchGroup(from int, to int)
	GroupsPosition() int
	GetGroup(idx int) (span.Interface, bool)
	RewindGroups(pos int)

	MatchNamedGroup(name string, from int, to int)
	NamedGroupsPosition() int
	GetNamedGroup(name string) (span.Interface, bool)
	RewindNamedGroups(pos int)

	MarkAsHole(from int, to int)
	HolesPosition() int
	RewindHoles(pos int)
}

type Input interface {
	ReadAt(int) rune
	Size() int
	Position() int
}

type Output interface {
	Yield(
		n Node,
		subString string,
		sp span.Interface,
		groups []span.Interface,
		namedGroups map[string]span.Interface,
	)

	LastPosOf(n Node) (int, bool)
	String() string
}
