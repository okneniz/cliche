package cliche

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/okneniz/cliche/onigmo"
	"github.com/okneniz/cliche/re2"
	tableTests "github.com/okneniz/cliche/testing"
)

func TestTree_Match(t *testing.T) {
	t.Parallel()

	type test struct {
		name   string
		path   string
		parser Parser
	}

	tests := []test{
		{
			name:   "onigmo",
			path:   "./testdata/onigmo/",
			parser: onigmo.Parser,
		},
		{
			name:   "re2",
			path:   "./testdata/re2/",
			parser: re2.Parser,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			files, err := tableTests.LoadAllTestFiles(t, test.path)
			if err != nil {
				t.Fatal(err)
			}

			for _, file := range files {
				testFile := file

				t.Run(testFile.Name, func(t *testing.T) {
					t.Parallel()

					for i, x := range testFile.Tests {
						example := x
						name := fmt.Sprintf("%d_%s", i, example.Name)

						t.Run(name, func(t *testing.T) {
							t.Parallel()
							t.Logf("input: '%s'", example.Input)

							tr := New(test.parser)
							err := tr.Add(example.Expressions...)
							require.NoError(t, err)

							t.Logf("tree: %s", tr)

							options, err := tableTests.ToScanOptions(example.Options...)
							require.NoError(t, err)

							t.Logf("options: %v", example.Options)

							matches := tr.Match(example.Input, options...)
							require.NoError(t, err)

							actual := tableTests.TestMatchesToExpectations(matches...)
							for _, w := range actual {
								w.Normalize()
							}

							sort.Slice(actual, func(i, j int) bool {
								return actual[i].String() < actual[j].String()
							})

							actualStr, err := json.MarshalIndent(actual, "", "  ")
							require.NoError(t, err)

							expectedStr, err := json.MarshalIndent(example.Want, "", "  ")
							require.NoError(t, err)

							require.Equal(t, string(expectedStr), string(actualStr))

							for _, expectation := range actual {
								for _, group := range expectation.Groups {
									if group.OutOfString {
										outOf := (group.From > expectation.Span.From && group.To > expectation.Span.From) ||
											(group.From < expectation.Span.From && group.To < expectation.Span.From)

										require.True(t, outOf)
									} else {
										require.LessOrEqual(t, expectation.Span.From, group.From)
										require.GreaterOrEqual(t, expectation.Span.To, group.To)
									}
								}

								for _, group := range expectation.NamedGroups {
									if group.OutOfString {
										outOf := (group.Span.From > expectation.Span.From && group.Span.To > expectation.Span.From) ||
											(group.Span.From < expectation.Span.From && group.Span.To < expectation.Span.From)

										require.True(t, outOf)
									} else {
										require.LessOrEqual(t, expectation.Span.From, group.Span.From)
										require.GreaterOrEqual(t, expectation.Span.To, group.Span.To)
									}
								}
							}
						})
					}
				})
			}
		})
	}
}

// TODO : у каждого движка свои
func TestTree_MatchErrors(t *testing.T) {
	t.Parallel()

	type test struct {
		expression string
		error      string
	}

	tests := []test{
		{
			expression: "[]",
			error:      "empty char-class: /[]/",
		},
		{
			expression: "[b-C]",
			error:      "empty char-class: /[b-C]/",
		},
		{
			expression: "(?i:[b-C])",
			error:      "empty char-class: /(?i:[b-C])/",
		},
		{
			expression: "a|[]",
			error:      "empty char-class: /a|[]/",
		},
		{
			expression: "(a|[])",
			error:      "empty char-class: /(a|[])/",
		},
		{
			expression: "a|[9-8]",
			error:      "empty char-class: /a|[9-8]/",
		},
		{
			expression: "(a|[9-8])",
			error:      "empty char-class: /(a|[9-8])/",
		},
		{
			expression: "[^]",
			error:      "empty char-class: /[^]/",
		},
		{
			expression: "[^b-C]",
			error:      "empty char-class: /[^b-C]/",
		},
		{
			expression: "(?i:[^b-C])",
			error:      "empty char-class: /(?i:[^b-C])/",
		},
		{
			expression: "a|[^]",
			error:      "empty char-class: /a|[^]/",
		},
		{
			expression: "(a|[^])",
			error:      "empty char-class: /(a|[^])/",
		},
		{
			expression: "a|[^9-8]",
			error:      "empty char-class: /a|[^9-8]/",
		},
		{
			expression: "(a|[^9-8])",
			error:      "empty char-class: /(a|[^9-8])/",
		},
	}

	for i, tt := range tests {
		test := tt
		name := fmt.Sprintf("%d_%s_%s", i, tt.expression, tt.error)

		t.Run(name, func(t *testing.T) {
			tr := New(DefaultParser)
			err := tr.Add(test.expression)
			t.Logf("tree: %s", tr)
			require.Error(t, err)
			require.Equal(t, test.error, err.Error())
		})
	}
}
