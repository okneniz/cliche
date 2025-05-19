package unicode

import (
	"unicode"

	"github.com/okneniz/cliche/node" // how to remove it from deps?
	"golang.org/x/exp/slices"
	"golang.org/x/text/unicode/rangetable"
)

func NewTable(runes ...rune) node.Table {
	slices.Sort(runes)
	runes = slices.Compact(runes)

	switch {
	case len(runes) == 0:
		return empty
	case len(runes) == 1:
		return runeTable{
			r: runes[0],
		}
	case len(runes) >= unicode.MaxRune:
		return everything
	}

	return newRangeTable(rangetable.New(runes...))
}

func MergeTables(tbls ...node.Table) node.Table {
	runes := make([]rune, 0)

	for x := rune(0); x <= unicode.MaxRune; x++ {
		for _, tbl := range tbls {
			if tbl.Include(x) {
				runes = append(runes, x)
			}
		}
	}

	return NewTable(runes...)
}

func NewTableByPredicate(p func(rune) bool) node.Table {
	runes := make([]rune, 0)

	for x := rune(0); x <= unicode.MaxRune; x++ {
		if p(x) {
			runes = append(runes, x)
		}
	}

	return NewTable(runes...)
}
