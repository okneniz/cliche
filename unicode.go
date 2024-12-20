package cliche

import (
	"unicode"

	"golang.org/x/text/unicode/rangetable"
)

func isWord(x rune) bool {
	return x == '_' || unicode.IsLetter(x) || unicode.IsDigit(x)
}

func isHex(x rune) bool {
	return x >= '0' && x <= '9' ||
		x >= 'a' && x <= 'f' ||
		x >= 'A' && x <= 'F'
}

func negatiateTable(table *unicode.RangeTable) *unicode.RangeTable {
	runes := make([]rune, 0)

	for x := rune(1); x <= unicode.MaxRune; x++ {
		if !unicode.In(x, table) {
			runes = append(runes, x)
		}
	}

	return rangetable.New(runes...)
}
