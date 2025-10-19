package unicode

import (
	"fmt"
	"unicode"

	"github.com/okneniz/cliche/node"
)

type runeTable struct {
	r rune
}

func (t runeTable) Include(x rune) bool {
	return t.r == x
}

func (t runeTable) Invert() node.Table {
	runes := make([]rune, 0)

	for x := rune(1); x <= unicode.MaxRune; x++ {
		if t.r != x {
			runes = append(runes, x)
		}
	}

	return NewTable(runes...)
}

func (t runeTable) Empty() bool {
	return false
}

func (t runeTable) String() string {
	return fmt.Sprintf("[%d]", t.r)
}
