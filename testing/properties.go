package testing

import (
	"fmt"
	"math/rand/v2"
	"strings"

	ohsnap "github.com/okneniz/oh-snap"
)

// Simplest regexp generator
type TestRegexp struct {
	Expression    string
	MatchedString string
}

type arbitraryRegexp struct {
	rand *rand.Rand

	maxDeep int

	arbType ohsnap.Arbitrary[int]
	arbSize ohsnap.Arbitrary[int]
	arbRune ohsnap.Arbitrary[rune]
}

func ArbitraryRegexp(rnd *rand.Rand, maxDeep, maxSize int) ohsnap.Arbitrary[TestRegexp] {
	arbRune := ohsnap.OneOf(
		rnd,
		[]ohsnap.Arbitrary[rune]{
			ohsnap.OneOfValue(rnd, []rune("abcdefghijklmnopqrstuvwxyz")...),
			ohsnap.OneOfValue(rnd, []rune("01234567890")...),
		},
	)

	return &arbitraryRegexp{
		rand:    rnd,
		maxDeep: maxDeep,
		arbType: ohsnap.ArbitraryInt(rnd, 1, 1000),
		arbSize: ohsnap.ArbitraryInt(rnd, 1, maxSize),
		arbRune: arbRune,
	}
}

func (a *arbitraryRegexp) Generate() TestRegexp {
	size := a.arbSize.Generate()

	reg := TestRegexp{}

	for i := 0; i < size; i++ {
		subExp := a.generateExp(1)

		reg.Expression += subExp.Expression
		reg.MatchedString += subExp.MatchedString
	}

	return reg
}

func (a *arbitraryRegexp) generateExp(deep int) TestRegexp {
	if deep >= a.maxDeep {
		return a.generateString()
	}

	var reg TestRegexp

	switch a.arbType.Generate() % 3 {
	case 0:
		reg = a.generateString()
	case 1:
		reg = a.generateAlt(deep + 1)
	case 2:
		reg = a.generateClass()
	default:
		reg = a.generateGroup(deep + 1)
	}

	return a.optionalQuantifier(reg)
}

func (a *arbitraryRegexp) optionalQuantifier(reg TestRegexp) TestRegexp {
	if (a.arbType.Generate() % 10) < 7 {
		return reg
	}

	size := 1 + a.arbType.Generate()%5

	switch a.arbType.Generate() % 5 {
	case 0: // {1}
		matchedString := strings.Repeat(reg.MatchedString, size)

		return TestRegexp{
			Expression:    fmt.Sprintf("(%s){%d}", reg.Expression, size),
			MatchedString: matchedString,
		}
	case 1: // {,1}
		repeats := 1 + rand.IntN(size)
		matchedString := strings.Repeat(reg.MatchedString, repeats)

		return TestRegexp{
			Expression:    fmt.Sprintf("(%s){,%d}", reg.Expression, size),
			MatchedString: matchedString,
		}
	case 2: // {1,}
		repeats := size + rand.IntN(size)
		matchedString := strings.Repeat(reg.MatchedString, repeats)

		return TestRegexp{
			Expression:    fmt.Sprintf("(%s){%d,}", reg.Expression, size),
			MatchedString: matchedString,
		}
	case 3: // +
		repeats := 1 + rand.IntN(size)
		matchedString := strings.Repeat(reg.MatchedString, repeats)

		return TestRegexp{
			Expression:    fmt.Sprintf("(%s)+", reg.Expression),
			MatchedString: matchedString,
		}
	case 4: // *
		repeats := 1 + rand.IntN(size)
		matchedString := strings.Repeat(reg.MatchedString, repeats)

		return TestRegexp{
			Expression:    fmt.Sprintf("(%s)*", reg.Expression),
			MatchedString: matchedString,
		}
	case 5: // ?
		// skip
	}

	return reg
}

func (a *arbitraryRegexp) generateChar() rune {
	return a.arbRune.Generate()
}

func (a *arbitraryRegexp) generateString() TestRegexp {
	size := a.arbSize.Generate()

	runes := make([]rune, size)
	for i := 0; i < size; i++ {
		runes[i] = a.generateChar()
	}

	str := string(runes)

	return TestRegexp{
		MatchedString: str,
		Expression:    str,
	}
}

func (a *arbitraryRegexp) generateAlt(deep int) TestRegexp {
	deep++
	size := a.arbSize.Generate()

	matchedIndex := rand.IntN(size)
	matchedString := ""

	variants := make([]string, size)
	for i := 0; i < size; i++ {
		exp := a.generateExp(deep)
		variants[i] = exp.Expression

		if i == matchedIndex {
			matchedString = exp.MatchedString
		}
	}

	expression := strings.Join(variants, "|")

	return TestRegexp{
		MatchedString: matchedString,
		Expression:    fmt.Sprintf("(%s)", expression),
	}
}

func (a *arbitraryRegexp) generateGroup(deep int) TestRegexp {
	exp := a.generateExp(deep + 1)

	return TestRegexp{
		MatchedString: exp.MatchedString,
		Expression:    fmt.Sprintf("(%s)", exp.Expression),
	}
}

func (a *arbitraryRegexp) generateClass() TestRegexp {
	size := a.arbSize.Generate()

	matchedString := ""
	matchedIndex := rand.IntN(size)

	expression := make([]string, size)
	for i := 0; i < size; i++ {
		item := a.generateClassItem()
		expression[i] = item.Expression

		if i == matchedIndex {
			matchedString = item.MatchedString
		}
	}

	return TestRegexp{
		MatchedString: matchedString,
		Expression:    fmt.Sprintf("[%s]", strings.Join(expression, "")),
	}
}

func (a *arbitraryRegexp) generateClassItem() TestRegexp {
	switch a.arbType.Generate() % 2 {
	case 0:
		return a.generateCharTable()
	case 1:
		return a.generateRange()
	default:
		panic("unknown item type")
	}
}

func (a *arbitraryRegexp) generateCharTable() TestRegexp {
	char := a.generateChar()
	str := string(char)

	return TestRegexp{
		MatchedString: str,
		Expression:    str,
	}
}

func (a *arbitraryRegexp) generateRange() TestRegexp {
	from := a.generateChar()
	to := a.generateChar()

	if from > to {
		x := from
		from = to
		to = x
	}

	matchedString := string(from) // TODO : choose random from range

	return TestRegexp{
		Expression:    fmt.Sprintf("%s-%s", string(from), string(to)),
		MatchedString: matchedString,
	}
}

func (*arbitraryRegexp) Shrink(value TestRegexp) []TestRegexp {
	return nil
}
