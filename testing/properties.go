package testing

import (
	"math/rand/v2"
	"unicode"

	ohsnap "github.com/okneniz/oh-snap"
)

type TestRegexp struct {
	Expression    string
	MatchedString string
}

type arbitraryRegexp struct {
	rand *rand.Rand

	min, max int

	arbType ohsnap.Arbitrary[int]
	arbRune ohsnap.Arbitrary[rune]
}

func ArbitraryRegexp(rnd *rand.Rand) ohsnap.Arbitrary[TestRegexp] {
	arbRune := ohsnap.OneOf(
		rnd,
		[]ohsnap.Arbitrary[rune]{
			ohsnap.RuneFromTable(rnd, unicode.L),
			ohsnap.RuneFromTable(rnd, unicode.Digit),
			ohsnap.RuneFromTable(rnd, unicode.Space),
		},
	)

	return &arbitraryRegexp{
		rand:    rnd,
		min:     1,
		max:     10,
		arbType: ohsnap.ArbitraryInt(rnd, 1, 2),
		arbRune: arbRune,
	}
}

func (a *arbitraryRegexp) Generate() TestRegexp {
	size := a.rand.IntN(a.max-a.min+1) + int(a.min)

	reg := TestRegexp{}

	for i := 0; i < size; i++ {
		subExp := a.generateExp()

		reg.Expression += subExp.Expression
		reg.MatchedString += subExp.MatchedString
	}

	return reg
}

func (a *arbitraryRegexp) generateExp() TestRegexp {
	// switch a.arbType.Generate() {
	// case 1:
	char := a.generateChar()

	return TestRegexp{
		MatchedString: string(char),
		Expression:    string(char),
	}
	// default:
	// 	panic("unknown type")
	// }
}

func (a *arbitraryRegexp) generateChar() rune {
	return a.arbRune.Generate()
}

func (*arbitraryRegexp) Shrink(value TestRegexp) []TestRegexp {
	return nil
}
