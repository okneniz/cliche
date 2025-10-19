package cliche

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	tableTests "github.com/okneniz/cliche/testing"
)

func TestTree_Match(t *testing.T) {
	t.Parallel()

	files, err := tableTests.LoadAllTestFiles(t, "./testdata/base/")
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		testFile := file

		t.Run(testFile.Name, func(t *testing.T) {
			t.Parallel()

			for _, ts := range testFile.Tests {
				test := ts

				t.Run(test.Name, func(t *testing.T) {
					t.Parallel()
					t.Logf("input: '%s'", test.Input)

					tr := New(DefaultParser)
					err := tr.Add(test.Expressions...)
					require.NoError(t, err)

					t.Logf("tree: %s", tr)

					options, err := tableTests.ToScanOptions(test.Options...)
					require.NoError(t, err)

					t.Logf("options: %v", test.Options)

					matches := tr.Match(test.Input, options...)
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

					expectedStr, err := json.MarshalIndent(test.Want, "", "  ")
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
}

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
