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
	Merge(Node) // remove, implement Merge(Node, Node) method in parser or tree

	// Add parent to travers
	// Should return bool to interupt traversing?
	Traverse(func(Node))

	// TODO : works only for fixed chain with one end node?
	// don't work for tree?

	// TODO : it's improtant for group too have chain in Value instead tree
	// make special type for this case?

	// TODO : what about alternation of chains?
	Size() (int, bool)

	// TODO : what about anchors, is it endless or zero sized?
}

type Table interface {
	Include(rune) bool
	Invert() Table
	String() string
}

// TODO : rename to Visitor?
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

type Callback func(x Node, from int, to int, empty bool)
