package cliche

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/rangetable"
)

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

	for x := rune(0); x <= unicode.MaxRune; x++ {
		if p(x) {
			runes = append(runes, x)
		}
	}

	return rangetable.New(runes...)
}

func rangeTableKey(table *unicode.RangeTable) string {
	if getCharsCount(table) == 1 { // TODO : not really good
		return string(getAllChars(table))
	}

	b := new(strings.Builder)

	b.WriteString("[")

	comma := false

	if len(table.R16) > 0 {
		b.WriteString("R16(")

		for i, r := range table.R16 {
			b.WriteString(fmt.Sprintf("%d", r.Lo))
			b.WriteString("-")
			b.WriteString(fmt.Sprintf("%d", r.Hi))

			if r.Stride != 1 {
				b.WriteString("-")
				b.WriteString(fmt.Sprintf("%d", r.Stride))
			}

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

			if r.Stride != 1 {
				b.WriteString("-")
				b.WriteString(fmt.Sprintf("%d", r.Stride))
			}

			if i != len(table.R32)-1 {
				b.WriteString(",")
			}
		}

		b.WriteString(")")
	}

	b.WriteString("]")

	return b.String()
}

func getCharsCount(table *unicode.RangeTable) int {
	var charsSum int

	for _, r := range table.R16 {
		charsSum += charsInR16(r)
	}
	for _, r := range table.R32 {
		charsSum += charsInR32(r)
	}

	return charsSum
}

func charsInR16(r unicode.Range16) int {
	return int((r.Hi-r.Lo)/r.Stride + 1)
}

func charsInR32(r unicode.Range32) int {
	return int((r.Hi-r.Lo)/r.Stride + 1)
}

func getAllChars(table *unicode.RangeTable) []rune {
	res := make([]rune, 0)

	for _, r := range table.R16 {
		for c := r.Lo; c <= r.Hi; c += r.Stride {
			res = append(res, rune(c))
		}
	}
	for _, r := range table.R32 {
		for c := r.Lo; c <= r.Hi; c += r.Stride {
			res = append(res, rune(c))
		}
	}

	return res
}
