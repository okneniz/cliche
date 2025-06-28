package quantity

import (
	"math"
	"testing"
)

func TestQuantity_Include(t *testing.T) {
	t.Parallel()

	type example struct {
		name   string
		q      *Quantity
		input  int
		output bool
	}

	tests := []example{
		{
			name:   "{0,1} or '?' include 0",
			q:      New(0, 1),
			input:  0,
			output: true,
		},
		{
			name:   "{0,1} or '?' include 1",
			q:      New(0, 1),
			input:  1,
			output: true,
		},
		{
			name:   "{0,1} or '?' exclude 2",
			q:      New(0, 1),
			input:  2,
			output: false,
		},
		{
			name:   "{1,} or '+' include 0",
			q:      NewEndlessQuantity(1),
			input:  0,
			output: false,
		},
		{
			name:   "{1,} or '+' include 1",
			q:      NewEndlessQuantity(1),
			input:  1,
			output: true,
		},
		{
			name:   "{1,} or '+' include max int",
			q:      NewEndlessQuantity(1),
			input:  math.MaxInt,
			output: true,
		},
		{
			name:   "{0,} or '*' include 0",
			q:      NewEndlessQuantity(0),
			input:  0,
			output: true,
		},
		{
			name:   "{0,} or '*' include 1",
			q:      NewEndlessQuantity(0),
			input:  1,
			output: true,
		},
		{
			name:   "{0,} or '*' include max int",
			q:      NewEndlessQuantity(0),
			input:  math.MaxInt,
			output: true,
		},
		{
			name:   "{2,} exclude 0",
			q:      NewEndlessQuantity(2),
			input:  0,
			output: false,
		},
		{
			name:   "{2,} exclude 1",
			q:      NewEndlessQuantity(2),
			input:  1,
			output: false,
		},
		{
			name:   "{2,} include 2",
			q:      NewEndlessQuantity(2),
			input:  2,
			output: true,
		},
		{
			name:   "{2} exclude 3",
			q:      New(2, 2),
			input:  3,
			output: false,
		},
		{
			name:   "{0,2} include 0",
			q:      New(0, 2),
			input:  0,
			output: true,
		},
		{
			name:   "{0,2} include 1",
			q:      New(0, 2),
			input:  1,
			output: true,
		},
		{
			name:   "{0,2} include 2",
			q:      New(0, 2),
			input:  2,
			output: true,
		},
		{
			name:   "{0,2} exclude 3",
			q:      New(0, 2),
			input:  3,
			output: false,
		},
		{
			name:   "{2,2} exclude 0",
			q:      New(2, 2),
			input:  0,
			output: false,
		},
		{
			name:   "{2,2} exclude 1",
			q:      New(2, 2),
			input:  1,
			output: false,
		},
		{
			name:   "{2,2} include 2",
			q:      New(2, 2),
			input:  2,
			output: true,
		},
		{
			name:   "{2,2} exclude 3",
			q:      New(2, 2),
			input:  3,
			output: false,
		},

		{
			name:   "{2,3} exclude 0",
			q:      New(2, 3),
			input:  0,
			output: false,
		},
		{
			name:   "{2,3} exclude 1",
			q:      New(2, 3),
			input:  1,
			output: false,
		},
		{
			name:   "{2,3} include 2",
			q:      New(2, 3),
			input:  2,
			output: true,
		},
		{
			name:   "{2,3} include 3",
			q:      New(2, 3),
			input:  3,
			output: true,
		},
		{
			name:   "{2,3} exclude 4",
			q:      New(2, 3),
			input:  4,
			output: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.q.Include(test.input)

			t.Log("input", test.input)

			if result != test.output {
				t.Fatalf("expected %v, actual %v", test.output, result)
			}
		})
	}
}
