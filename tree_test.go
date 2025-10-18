package cliche

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	tableTests "github.com/okneniz/cliche/testing"
)

func TestTree_New(t *testing.T) {
	t.Parallel()

	// TODO : add more examples (look behind, look ahead, conditions, etc)

	tr := New(DefaultParser)
	err := tr.Add(
		"x",
		"t",
		"te",
		"test",
		"tost",
		"tot",
		".",
		"^",
		"$",
		`\d`,
		"\\D",
		"\\w",
		"\\W",
		"\\S",
		"\\s",
		"\\A",
		"\\z",
		"\\?",
		"\\.",
		"[a-z]",
		"[^a-z]",
		"[0-9]",
		"[^0-9]",
		"[0-9a-z]",
		"[a-z0-9]",
		`[\d]`,
		`[0-9a-zxy\d]`,
		"[^0-9a-zxy\\d]",
		"(y)",
		"(y|x)",
		"(x|y)",
		"x|y",
		"y|x",
		"(?:y)",
		"(?<x>y)",
		"(?>y)",
		"(?=y)",
		"(?!y)",
		"(?<=y)",
		"(?<!y)",
		"foo",
		"(foo)",
		"(foo|bar|baz)",
		"(foo|bar|baz)+",
		"(?:foo|bar|baz)+",
		"(?<name>x|y|z)",
		"(?<name>y|x|z)",
		"(?<test>foo|bar|baz)+",
		"(?<test>foo|(ba|za|r)|baz)+",
		"[^abc1-3]?",
		"\\d*",
		"a{1}",
		"a{1,}",
		"a+",
		"a{0,}",
		"a*",
		"a{0,1}",
		"a?",
		"a{1,1}",
		"a{1,3}",
		"a{1,3,}",
		"[a{1,3}bc]+",
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(tr)
}

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
