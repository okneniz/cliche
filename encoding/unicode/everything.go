package unicode

import "github.com/okneniz/cliche/node"

var (
	everything = everythingTable{}
)

type everythingTable struct{}

func (t everythingTable) Include(x rune) bool {
	return true
}

func (t everythingTable) Invert() node.Table {
	return empty
}

func (t everythingTable) Empty() bool {
	return false
}

func (t everythingTable) String() string {
	return "[ALL]"
}
