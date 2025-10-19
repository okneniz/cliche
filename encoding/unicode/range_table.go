package unicode

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/okneniz/cliche/node"
)

type rangeTable struct {
	tbl *unicode.RangeTable
}

func newRangeTable(tbl *unicode.RangeTable) node.Table {
	return &rangeTable{
		tbl: tbl,
	}
}

func (t *rangeTable) Include(x rune) bool {
	return unicode.In(x, t.tbl)
}

func (t *rangeTable) Invert() node.Table {
	runes := make([]rune, 0)

	for x := rune(1); x <= unicode.MaxRune; x++ {
		if !unicode.In(x, t.tbl) {
			runes = append(runes, x)
		}
	}

	return NewTable(runes...)
}

func (t *rangeTable) Empty() bool {
	return len(t.tbl.R16) == 0 && len(t.tbl.R32) == 0
}

func (t *rangeTable) String() string {
	b := new(strings.Builder)

	b.WriteString("[")

	comma := false

	if len(t.tbl.R16) > 0 {
		b.WriteString("R16(")

		for i, r := range t.tbl.R16 {
			b.WriteString(fmt.Sprintf("%d", r.Lo))
			b.WriteString("-")
			b.WriteString(fmt.Sprintf("%d", r.Hi))

			if r.Stride != 1 {
				b.WriteString("-")
				b.WriteString(fmt.Sprintf("%d", r.Stride))
			}

			if i != len(t.tbl.R16)-1 {
				b.WriteString(",")
			}
		}

		b.WriteString(")")
		comma = true
	}

	if len(t.tbl.R32) > 0 {
		if comma {
			b.WriteString(",")
		}

		b.WriteString("R32(")

		for i, r := range t.tbl.R32 {
			b.WriteString(fmt.Sprintf("%d", r.Lo))
			b.WriteString("-")
			b.WriteString(fmt.Sprintf("%d", r.Hi))

			if r.Stride != 1 {
				b.WriteString("-")
				b.WriteString(fmt.Sprintf("%d", r.Stride))
			}

			if i != len(t.tbl.R32)-1 {
				b.WriteString(",")
			}
		}

		b.WriteString(")")
	}

	b.WriteString("]")

	return b.String()
}
