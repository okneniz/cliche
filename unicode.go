package cliche

import (
	"fmt"
	"strings"
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

func predicateToTable(p func(rune) bool) *unicode.RangeTable {
	runes := make([]rune, 0)

	for x := rune(1); x <= unicode.MaxRune; x++ {
		if p(x) {
			runes = append(runes, x)
		}
	}

	return rangetable.New(runes...)
}

func rangeTableKey(table *unicode.RangeTable) string {
	b := new(strings.Builder)

	b.WriteString("[")

	comma := false

	if len(table.R16) > 0 {
		b.WriteString("R16(")

		for i, r := range table.R16 {
			b.WriteString(fmt.Sprintf("%d", r.Lo))
			b.WriteString("-")
			b.WriteString(fmt.Sprintf("%d", r.Hi))
			b.WriteString("-")
			b.WriteString(fmt.Sprintf("%d", r.Stride))

			if i != len(table.R16)-1 {
				b.WriteString(",")
			}
		}

		b.WriteString(")")
		comma = true
	}

	if len(table.R32) > 0 {
		if comma {
			b.WriteString(",")
		}

		b.WriteString("R32(")

		for i, r := range table.R32 {
			b.WriteString(fmt.Sprintf("%d", r.Lo))
			b.WriteString("-")
			b.WriteString(fmt.Sprintf("%d", r.Hi))
			b.WriteString("-")
			b.WriteString(fmt.Sprintf("%d", r.Stride))

			if i != len(table.R32)-1 {
				b.WriteString(",")
			}
		}

		b.WriteString(")")
	}

	b.WriteString("]")

	return b.String()
}
