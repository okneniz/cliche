package unicode

import (
	"unicode"

	"github.com/okneniz/cliche/node"
)

var (
	empty = emptyTable{}
)

type emptyTable struct{}

func (t emptyTable) Include(x rune) bool {
	return false
}

func (t emptyTable) Invert() node.Table {
	runes := make([]rune, 0)

	for x := rune(1); x <= unicode.MaxRune; x++ {
		runes = append(runes, x)
	}

	return NewTable(runes...)
}

func (t emptyTable) Empty() bool {
	return true
}

func (t emptyTable) String() string {
	return "[]"
}
