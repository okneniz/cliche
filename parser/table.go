package parser

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/okneniz/cliche/node"
	"golang.org/x/text/unicode/rangetable"
)

// TODO : move to special package

type UnicodeTable struct {
	tbl *unicode.RangeTable
}

var _ node.Table = new(UnicodeTable)

func NewUnicodeTable(tbl *unicode.RangeTable) *UnicodeTable {
	return &UnicodeTable{
		tbl: tbl,
	}
}

func MergeUnicodeTables(tbls ...node.Table) node.Table {
	runes := make([]rune, 0)

	for x := rune(0); x <= unicode.MaxRune; x++ {
		for _, tbl := range tbls {
			if tbl.Include(x) {
				runes = append(runes, x)
			}
		}
	}

	return NewUnicodeTableFor(runes...)
}

func NewUnicodeTableFor(items ...rune) *UnicodeTable {
	return NewUnicodeTable(rangetable.New(items...))
}

func NewUnicodeTableByPredicate(p func(rune) bool) *UnicodeTable {
	runes := make([]rune, 0)

	for x := rune(0); x <= unicode.MaxRune; x++ {
		if p(x) {
			runes = append(runes, x)
		}
	}

	return NewUnicodeTable(rangetable.New(runes...))
}

func (t *UnicodeTable) Include(x rune) bool {
	return unicode.In(x, t.tbl)
}

func (t *UnicodeTable) Invert() node.Table {
	runes := make([]rune, 0)

	for x := rune(1); x <= unicode.MaxRune; x++ {
		if !unicode.In(x, t.tbl) {
			runes = append(runes, x)
		}
	}

	return NewUnicodeTable(rangetable.New(runes...))
}

func (t *UnicodeTable) String() string {
	return rangeTableKey(t.tbl)
}

// конвертить в BitSet
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
