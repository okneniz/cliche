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

	const iterations = 1_000_000

	seed := time.Now().UnixNano()
	t.Logf("seed: %v", seed)

	rnd := rand.New(rand.NewPCG(0, uint64(seed)))
	arb := tests.ArbitraryRegexp(rnd)

	ohsnap.Check(t, iterations, arb, func(reg tests.TestRegexp) bool {
		tr := cliche.New(onigmo.OnigmoParser)
		tr.Add(reg.Expression)
		matches := tr.Match(reg.MatchedString)

		if len(matches) == 0 {
			t.Log("expression", reg.Expression)
			t.Log("string", reg.MatchedString)
			return false
		}

		return true
	})

}
