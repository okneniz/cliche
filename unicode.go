package cliche

import (
	"unicode"
)

func isWord(x rune) bool {
	return x == '_' || unicode.IsLetter(x) || unicode.IsDigit(x)
}

func isHex(x rune) bool {
	return x >= '0' && x <= '9' ||
		x >= 'a' && x <= 'f' ||
		x >= 'A' && x <= 'F'
}
