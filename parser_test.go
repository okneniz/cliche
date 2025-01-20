package cliche

// func TestParseQuantifierKeys(t *testing.T) {
// 	t.Parallel()

// 	type example struct {
// 		expression    string
// 		key           string
// 		notQuantifier bool
// 		from          int
// 		to            *int
// 		more          bool
// 	}

// 	examples := []example{
// 		{
// 			expression: "x?",
// 			key:        "x?",
// 			from:       0,
// 			to:         pointer(1),
// 			more:       false,
// 		},
// 		{
// 			expression: "x{0,1}",
// 			key:        "x?",
// 			from:       0,
// 			to:         pointer(1),
// 			more:       false,
// 		},
// 		{
// 			expression: "x+",
// 			key:        "x+",
// 			from:       1,
// 			to:         nil,
// 			more:       true,
// 		},
// 		{
// 			expression: "x{1,}",
// 			key:        "x+",
// 			from:       1,
// 			to:         nil,
// 			more:       true,
// 		},
// 		{
// 			expression: "x*",
// 			key:        "x*",
// 			from:       0,
// 			to:         nil,
// 			more:       true,
// 		},
// 		{
// 			expression: "x{0,}",
// 			key:        "x*",
// 			from:       0,
// 			to:         nil,
// 			more:       true,
// 		},
// 		{
// 			expression: "x{1,1}",
// 			key:        "x{1}",
// 			from:       1,
// 			to:         nil,
// 			more:       false,
// 		},
// 		{
// 			expression: "x{1}",
// 			key:        "x{1}",
// 			from:       1,
// 			to:         nil,
// 			more:       false,
// 		},
// 		{
// 			expression: "x{2,}",
// 			key:        "x{2,}",
// 			from:       2,
// 			to:         nil,
// 			more:       true,
// 		},
// 		{
// 			expression: "x{,2}",
// 			key:        "x{0,2}",
// 			from:       0,
// 			to:         pointer(2),
// 			more:       false,
// 		},
// 		{
// 			expression: "x{0}",
// 			key:        "x{0}",
// 			from:       0,
// 			to:         nil,
// 			more:       false,
// 		},
// 		{
// 			expression:    "x{3,2}",
// 			notQuantifier: true,
// 		},
// 		{
// 			expression:    "x{2,2,}",
// 			notQuantifier: true,
// 		},
// 		{
// 			expression:    "x{,2,}",
// 			notQuantifier: true,
// 		},
// 	}

// 	parser := DefaultParser

// 	for _, test := range examples {
// 		t.Run(test.expression, func(t *testing.T) {
// 			input := buf.NewRunesBuffer(test.expression)

// 			output, err := parser.Parse(input)
// 			if err != nil {
// 				t.Fatal(err)
// 			}

// 			q, ok := output.(*node.Quantifier)
// 			if ok {
// 				if test.notQuantifier {
// 					t.Fatalf(
// 						"expected not *quantifier type, parser output %v",
// 						output,
// 					)
// 				}

// 				if q.GetKey() != test.key {
// 					t.Fatalf(
// 						"expected key %v, actual key %v",
// 						test.key,
// 						q.GetKey(),
// 					)
// 				}

// 				if q.From != test.from {
// 					t.Fatalf(
// 						"expected 'from' %v, actual %v",
// 						q.From,
// 						test.from,
// 					)
// 				}

// 				if test.to != nil && q.To == nil {
// 					t.Fatalf(
// 						"expected 'to' equal %v, actual is nil",
// 						*test.to,
// 					)
// 				}

// 				if test.to == nil && q.To != nil {
// 					t.Fatalf(
// 						"expected nil 'to', actual is %v",
// 						*q.To,
// 					)
// 				}

// 				if test.to != nil && q.To != nil && *test.to != *q.To {
// 					t.Fatalf(
// 						"expected 'to' %v, actual %v",
// 						*test.to,
// 						*q.To,
// 					)
// 				}

// 				if test.more != q.More {
// 					t.Fatalf(
// 						"expected 'more' %v, actual %v",
// 						test.more,
// 						&q.More,
// 					)
// 				}
// 			} else if !test.notQuantifier {
// 				t.Fatalf(
// 					"expected not *quantifier type, parser output %v",
// 					output,
// 				)
// 			}
// 		})
// 	}
// }

// func pointer[T any](x T) *T {
// 	return &x
// }
