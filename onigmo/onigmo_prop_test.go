package onigmo_test

import (
	"math/rand/v2"
	"testing"
	"time"

	"github.com/okneniz/cliche"
	"github.com/okneniz/cliche/onigmo"
	tests "github.com/okneniz/cliche/testing"
	ohsnap "github.com/okneniz/oh-snap"
)

func TestOnigmoProperties(t *testing.T) {
	t.Parallel()

	seed := time.Now().UnixNano()
	t.Logf("seed: %v", seed)

	t.Run("one by one", func(t *testing.T) {
		t.Parallel()

		const iterations = 1000

		rnd := rand.New(rand.NewPCG(0, uint64(seed)))
		arb := tests.ArbitraryRegexp(rnd, 3, 7)

		ohsnap.Check(t, iterations, arb, func(reg tests.TestRegexp) bool {
			tr := cliche.New(onigmo.OnigmoParser)

			err := tr.Add(reg.Expression)
			if err != nil {
				t.Log("add expression", reg.Expression)
				t.Log("string", reg.MatchedString)
				t.Error(err)

				return false
			}

			// t.Log("expression", reg.Expression)
			// t.Log("string", reg.MatchedString)
			// t.Log("tree", tr.String())

			matches := tr.Match(reg.MatchedString)
			if len(matches) == 0 {
				t.Log("expression", reg.Expression)
				t.Log("string", reg.MatchedString)
				return false
			}

			return true
		})
	})

	t.Run("in batch", func(t *testing.T) {
		t.Parallel()

		const iterations = 100

		rnd := rand.New(rand.NewPCG(0, uint64(seed)))
		arbExp := tests.ArbitraryRegexp(rnd, 3, 7)
		arb := ohsnap.ArbitrarySlice(rnd, arbExp, 10, 50)

		ohsnap.Check(t, iterations, arb, func(regs []tests.TestRegexp) bool {
			tr := cliche.New(onigmo.OnigmoParser)

			for _, reg := range regs {
				err := tr.Add(reg.Expression)
				if err != nil {
					t.Log("add expression", reg.Expression)
					t.Log("string", reg.MatchedString)
					t.Error(err)

					return false
				}
			}

			for _, reg := range regs {
				// t.Log("expression", reg.Expression)
				// t.Log("string", reg.MatchedString)
				// t.Log("tree", tr.String())

				matches := tr.Match(reg.MatchedString)
				if len(matches) == 0 {
					t.Log("expression", reg.Expression)
					t.Log("string", reg.MatchedString)

					return false
				}
			}

			return true
		})
	})
}
