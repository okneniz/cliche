package cliche

import (
	"math"
	"testing"
)

func TestQuantifierBounds(t *testing.T) {
	t.Parallel()

	type example struct {
		name   string
		q      quantifier
		input  int
		output bool
	}

	tests := []example{
		{
			name: "{0,1} or '?' include 0",
			q: quantifier{
				From: 0,
				To:   pointer(1),
				More: false,
			},
			input:  0,
			output: true,
		},
		{
			name: "{0,1} or '?' include 1",
			q: quantifier{
				From: 0,
				To:   pointer(1),
				More: false,
			},
			input:  1,
			output: true,
		},
		{
			name: "{0,1} or '?' exclude 2",
			q: quantifier{
				From: 0,
				To:   pointer(1),
				More: false,
			},
			input:  2,
			output: false,
		},
		{
			name: "{1} or '+' include 0",
			q: quantifier{
				From: 1,
				To:   nil,
				More: true,
			},
			input:  0,
			output: false,
		},
		{
			name: "{1} or '+' include 1",
			q: quantifier{
				From: 1,
				To:   nil,
				More: true,
			},
			input:  1,
			output: true,
		},
		{
			name: "{1} or '+' include max int",
			q: quantifier{
				From: 1,
				To:   nil,
				More: true,
			},
			input:  math.MaxInt,
			output: true,
		},
		{
			name: "{0,} or '*' include 0",
			q: quantifier{
				From: 0,
				To:   nil,
				More: true,
			},
			input:  0,
			output: true,
		},
		{
			name: "{0,} or '*' include 1",
			q: quantifier{
				From: 0,
				To:   nil,
				More: true,
			},
			input:  1,
			output: true,
		},
		{
			name: "{0,} or '*' include max int",
			q: quantifier{
				From: 0,
				To:   nil,
				More: true,
			},
			input:  math.MaxInt,
			output: true,
		},
		{
			name: "{2} exclude 0",
			q: quantifier{
				From: 2,
				To:   nil,
				More: false,
			},
			input:  0,
			output: false,
		},
		{
			name: "{2} exclude 1",
			q: quantifier{
				From: 2,
				To:   nil,
				More: false,
			},
			input:  1,
			output: false,
		},
		{
			name: "{2} include 2",
			q: quantifier{
				From: 2,
				To:   nil,
				More: false,
			},
			input:  2,
			output: true,
		},
		{
			name: "{2} exclude 3",
			q: quantifier{
				From: 2,
				To:   nil,
				More: false,
			},
			input:  3,
			output: false,
		},
		{
			name: "{0,2} include 0",
			q: quantifier{
				From: 0,
				To:   pointer(2),
				More: false,
			},
			input:  0,
			output: true,
		},
		{
			name: "{0,2} include 1",
			q: quantifier{
				From: 0,
				To:   pointer(2),
				More: false,
			},
			input:  1,
			output: true,
		},
		{
			name: "{0,2} include 2",
			q: quantifier{
				From: 0,
				To:   pointer(2),
				More: false,
			},
			input:  2,
			output: true,
		},
		{
			name: "{0,2} exclude 3",
			q: quantifier{
				From: 0,
				To:   pointer(2),
				More: false,
			},
			input:  3,
			output: false,
		},
		{
			name: "{2,2} exclude 0",
			q: quantifier{
				From: 2,
				To:   pointer(2),
				More: false,
			},
			input:  0,
			output: false,
		},
		{
			name: "{2,2} exclude 1",
			q: quantifier{
				From: 2,
				To:   pointer(2),
				More: false,
			},
			input:  1,
			output: false,
		},
		{
			name: "{2,2} include 2",
			q: quantifier{
				From: 2,
				To:   pointer(2),
				More: false,
			},
			input:  2,
			output: true,
		},
		{
			name: "{2,2} exclude 3",
			q: quantifier{
				From: 2,
				To:   pointer(2),
				More: false,
			},
			input:  3,
			output: false,
		},

		{
			name: "{2,3} exclude 0",
			q: quantifier{
				From: 2,
				To:   pointer(3),
				More: false,
			},
			input:  0,
			output: false,
		},
		{
			name: "{2,3} exclude 1",
			q: quantifier{
				From: 2,
				To:   pointer(3),
				More: false,
			},
			input:  1,
			output: false,
		},
		{
			name: "{2,3} include 2",
			q: quantifier{
				From: 2,
				To:   pointer(3),
				More: false,
			},
			input:  2,
			output: true,
		},
		{
			name: "{2,3} include 3",
			q: quantifier{
				From: 2,
				To:   pointer(3),
				More: false,
			},
			input:  3,
			output: true,
		},
		{
			name: "{2,3} exclude 4",
			q: quantifier{
				From: 2,
				To:   pointer(3),
				More: false,
			},
			input:  4,
			output: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.q.inBounds(test.input)

			t.Log("input", test.input)

			if result != test.output {
				t.Fatalf("expected %v, actual %v", test.output, result)
			}
		})
	}
}

func TestQuantifier_GetKey(t *testing.T) {
	t.Parallel()

	type example struct {
		expression    string
		key           string
		notQuantifier bool
		from          int
		to            *int
		more          bool
	}

	examples := []example{
		{
			expression: "x?",
			key:        "x?",
			from:       0,
			to:         pointer(1),
			more:       false,
		},
		{
			expression: "x{0,1}",
			key:        "x?",
			from:       0,
			to:         pointer(1),
			more:       false,
		},
		{
			expression: "x+",
			key:        "x+",
			from:       1,
			to:         nil,
			more:       true,
		},
		{
			expression: "x{1,}",
			key:        "x+",
			from:       1,
			to:         nil,
			more:       true,
		},
		{
			expression: "x*",
			key:        "x*",
			from:       0,
			to:         nil,
			more:       true,
		},
		{
			expression: "x{0,}",
			key:        "x*",
			from:       0,
			to:         nil,
			more:       true,
		},
		{
			expression: "x{1,1}",
			key:        "x{1}",
			from:       1,
			to:         nil,
			more:       false,
		},
		{
			expression: "x{1}",
			key:        "x{1}",
			from:       1,
			to:         nil,
			more:       false,
		},
		{
			expression: "x{2,}",
			key:        "x{2,}",
			from:       2,
			to:         nil,
			more:       true,
		},
		{
			expression: "x{,2}",
			key:        "x{0,2}",
			from:       0,
			to:         pointer(2),
			more:       false,
		},
		{
			expression: "x{0}",
			key:        "x{0}",
			from:       0,
			to:         nil,
			more:       false,
		},
		{
			expression:    "x{3,2}",
			notQuantifier: true,
		},
		{
			expression:    "x{2,2,}",
			notQuantifier: true,
		},
		{
			expression:    "x{,2,}",
			notQuantifier: true,
		},
	}

	parser := DefaultParser

	for _, test := range examples {
		t.Run(test.expression, func(t *testing.T) {
			input := newBuffer(test.expression)

			output, err := parser.Parse(input)
			if err != nil {
				t.Fatal(err)
			}

			q, ok := output.(*quantifier)
			if ok {
				if test.notQuantifier {
					t.Fatalf(
						"expected not *quantifier type, parser output %v",
						output,
					)
				}

				if q.GetKey() != test.key {
					t.Fatalf(
						"expected key %v, actual key %v",
						test.key,
						q.GetKey(),
					)
				}

				if q.From != test.from {
					t.Fatalf(
						"expected 'from' %v, actual %v",
						q.From,
						test.from,
					)
				}

				if test.to != nil && q.To == nil {
					t.Fatalf(
						"expected 'to' equal %v, actual is nil",
						*test.to,
					)
				}

				if test.to == nil && q.To != nil {
					t.Fatalf(
						"expected nil 'to', actual is %v",
						*q.To,
					)
				}

				if test.to != nil && q.To != nil && *test.to != *q.To {
					t.Fatalf(
						"expected 'to' %v, actual %v",
						*test.to,
						*q.To,
					)
				}

				if test.more != q.More {
					t.Fatalf(
						"expected 'more' %v, actual %v",
						test.more,
						&q.More,
					)
				}
			} else if !test.notQuantifier {
				t.Fatalf(
					"expected not *quantifier type, parser output %v",
					output,
				)
			}
		})
	}
}

func pointer[T any](x T) *T {
	return &x
}
