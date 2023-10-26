package regular

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

// https://www.rfc-editor.org/rfc/rfc9485.html#name-implementing-i-regexp

func TestTrie(t *testing.T) {
	t.Skip()

	tr, err := NewTrie(
		"x",
		"t",
		"te",
		"test",
		"tost",
		"tot",
		".",
		"^", // only for start of regexp?
		"$", // only for end of regexp?
		"\\d",
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
		"[0-9]",
		"[0-9a-z]",
		"[a-z0-9]",
		"[0-9a-zxy\\d]",
		"[^0-9a-zxy\\d]",
		"(y)",
		"(y|x)",
		"(x|y)",
		"x|y",
		"y|x",
		"(?:y)",
		"(?<x>y)",
		"foo",
		"(foo)",
		"(f|b)",
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

	// t.Logf("size: %d", tr.Size())
	// t.Logf("wtf %#v", tr)
	t.Log(tr)
}

func TestTrieCompression(t *testing.T) {
	t.Parallel()

	// Positive and negative set store elements in ordered collection.
	// This allows you to avoid duplicating a certain number of expressions.
	// For example [a-z1-2] and [1-2a-z] are equal expressions for trie.
	t.Run("sets", func(t *testing.T) {
		tr, err := NewTrie()
		require.NoError(t, err)
		require.Equal(t, tr.Size(), 0)

		err = tr.Add("[a-z1-2]")
		require.NoError(t, err)
		require.Equal(t, tr.Size(), 1)

		err = tr.Add("[1-2a-z]")
		require.NoError(t, err)
		require.Equal(t, tr.Size(), 1)

		err = tr.Add("[12a-z]")
		require.NoError(t, err)
		require.Equal(t, tr.Size(), 2)

		err = tr.Add("[1a-z2]")
		require.NoError(t, err)
		require.Equal(t, tr.Size(), 2)

		err = tr.Add("[abc]")
		require.NoError(t, err)
		require.Equal(t, tr.Size(), 3)

		err = tr.Add("[cab]")
		require.NoError(t, err)
		require.Equal(t, tr.Size(), 3)

		// TODO : may be remove duplications?
	})

	// Some quantifiers have the same meaning, but have different symbols.
	// For example:
	// - x+ is equal x{1,}
	// - x* is equal x{0,}
	// - x? is equal x{0,1}
	t.Run("quantifiers", func(t *testing.T) {
		tr, err := NewTrie()
		require.NoError(t, err)
		require.Equal(t, tr.Size(), 0)

		err = tr.Add("x+")
		require.NoError(t, err)
		require.Equal(t, tr.Size(), 1)

		err = tr.Add("x{1,}")
		require.NoError(t, err)
		require.Equal(t, tr.Size(), 1)

		err = tr.Add("x?")
		require.NoError(t, err)
		require.Equal(t, tr.Size(), 2)

		err = tr.Add("x{0,1}")
		require.NoError(t, err)
		require.Equal(t, tr.Size(), 2)

		err = tr.Add("x*")
		require.NoError(t, err)
		require.Equal(t, tr.Size(), 3)

		err = tr.Add("x{0,}")
		require.NoError(t, err)
		require.Equal(t, tr.Size(), 3)
	})
}

func TestMatch(t *testing.T) {
	examples := map[string][]example{
		"simple": {
			{
				name: "match chars as is",
				regexps: []string{
					"te",
					"toast",
					"toaster",
					"word",
					"strong",
					"wizard",
					"test",
					"string",
					"s",
					"ing",
				},
				input: "testing string test ssss word words",
				output: []*FullMatch{
					{
						subString: "te",
						from:      0,
						to:        1,
						expressions: []string{
							"te",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "test",
						from:      0,
						to:        3,
						expressions: []string{
							"test",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "s",
						from:      2,
						to:        2,
						expressions: []string{
							"s",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "ing",
						from:      4,
						to:        6,
						expressions: []string{
							"ing",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "s",
						from:      8,
						to:        8,
						expressions: []string{
							"s",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "string",
						from:      8,
						to:        13,
						expressions: []string{
							"string",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "ing",
						from:      11,
						to:        13,
						expressions: []string{
							"ing",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "te",
						from:      15,
						to:        16,
						expressions: []string{
							"te",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "test",
						from:      15,
						to:        18,
						expressions: []string{
							"test",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "s",
						from:      17,
						to:        17,
						expressions: []string{
							"s",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "s",
						from:      20,
						to:        20,
						expressions: []string{
							"s",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "s",
						from:      21,
						to:        21,
						expressions: []string{
							"s",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "s",
						from:      22,
						to:        22,
						expressions: []string{
							"s",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "s",
						from:      23,
						to:        23,
						expressions: []string{
							"s",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "word",
						from:      25,
						to:        28,
						expressions: []string{
							"word",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "word",
						from:      30,
						to:        33,
						expressions: []string{
							"word",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "s",
						from:      34,
						to:        34,
						expressions: []string{
							"s",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
				},
			},
			{
				name:    "match '.'",
				regexps: []string{"t."},
				input:   "testing string test ssss word words",
				output: []*FullMatch{
					{
						subString: "te",
						from:      0,
						to:        1,
						expressions: []string{
							"t.",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "ti",
						from:      3,
						to:        4,
						expressions: []string{
							"t.",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "tr",
						from:      9,
						to:        10,
						expressions: []string{
							"t.",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "te",
						from:      15,
						to:        16,
						expressions: []string{
							"t.",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "t ",
						from:      18,
						to:        19,
						expressions: []string{
							"t.",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
				},
			},
			{
				name:    "match multiple '.'", // TODO : move to another group?
				regexps: []string{"t.."},
				input:   "testing string test ssss word words",
				output: []*FullMatch{
					{
						subString: "tes",
						from:      0,
						to:        2,
						expressions: []string{
							"t..",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "tin",
						from:      3,
						to:        5,
						expressions: []string{
							"t..",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "tri",
						from:      9,
						to:        11,
						expressions: []string{
							"t..",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "tes",
						from:      15,
						to:        17,
						expressions: []string{
							"t..",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "t s",
						from:      18,
						to:        20,
						expressions: []string{
							"t..",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
				},
			},
			{
				name:    "match '\\d'",
				regexps: []string{"\\d"},
				input:   "asd 1 jsdfk 4234",
				output: []*FullMatch{
					{
						subString: "1",
						from:      4,
						to:        4,
						expressions: []string{
							"\\d",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "4",
						from:      12,
						to:        12,
						expressions: []string{
							"\\d",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "2",
						from:      13,
						to:        13,
						expressions: []string{
							"\\d",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "3",
						from:      14,
						to:        14,
						expressions: []string{
							"\\d",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "4",
						from:      15,
						to:        15,
						expressions: []string{
							"\\d",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
				},
			},
			{
				name:    "match '\\D'",
				regexps: []string{"\\D"},
				input:   "asd 1 jsdfk 4234",
				output: []*FullMatch{
					{
						subString: "a",
						from:      0,
						to:        0,
						expressions: []string{
							"\\D",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "s",
						from:      1,
						to:        1,
						expressions: []string{
							"\\D",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "d",
						from:      2,
						to:        2,
						expressions: []string{
							"\\D",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: " ",
						from:      3,
						to:        3,
						expressions: []string{
							"\\D",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: " ",
						from:      5,
						to:        5,
						expressions: []string{
							"\\D",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "j",
						from:      6,
						to:        6,
						expressions: []string{
							"\\D",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "s",
						from:      7,
						to:        7,
						expressions: []string{
							"\\D",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "d",
						from:      8,
						to:        8,
						expressions: []string{
							"\\D",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "f",
						from:      9,
						to:        9,
						expressions: []string{
							"\\D",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "k",
						from:      10,
						to:        10,
						expressions: []string{
							"\\D",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: " ",
						from:      11,
						to:        11,
						expressions: []string{
							"\\D",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
				},
			},
			{
				name:    "match '\\w'",
				regexps: []string{"\\w"},
				input:   "asd 1 jsdfk 4234",
				output: []*FullMatch{
					{
						subString: "a",
						from:      0,
						to:        0,
						expressions: []string{
							"\\w",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "s",
						from:      1,
						to:        1,
						expressions: []string{
							"\\w",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "d",
						from:      2,
						to:        2,
						expressions: []string{
							"\\w",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "1",
						from:      4,
						to:        4,
						expressions: []string{
							"\\w",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "j",
						from:      6,
						to:        6,
						expressions: []string{
							"\\w",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "s",
						from:      7,
						to:        7,
						expressions: []string{
							"\\w",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "d",
						from:      8,
						to:        8,
						expressions: []string{
							"\\w",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "f",
						from:      9,
						to:        9,
						expressions: []string{
							"\\w",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "k",
						from:      10,
						to:        10,
						expressions: []string{
							"\\w",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "4",
						from:      12,
						to:        12,
						expressions: []string{
							"\\w",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "2",
						from:      13,
						to:        13,
						expressions: []string{
							"\\w",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "3",
						from:      14,
						to:        14,
						expressions: []string{
							"\\w",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "4",
						from:      15,
						to:        15,
						expressions: []string{
							"\\w",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
				},
			},
			{
				name:    "match '\\W'",
				regexps: []string{"\\W"},
				input:   "asd 1 jsdfk 4234!\n\r",
				output: []*FullMatch{
					{
						subString: " ",
						from:      3,
						to:        3,
						expressions: []string{
							"\\W",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: " ",
						from:      5,
						to:        5,
						expressions: []string{
							"\\W",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: " ",
						from:      11,
						to:        11,
						expressions: []string{
							"\\W",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "!",
						from:      16,
						to:        16,
						expressions: []string{
							"\\W",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "\n",
						from:      17,
						to:        17,
						expressions: []string{
							"\\W",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "\r",
						from:      18,
						to:        18,
						expressions: []string{
							"\\W",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
				},
			},
			{
				name:    "match '\\s'",
				regexps: []string{"\\s"},
				input:   "asd 1 jsdfk 4234",
				output: []*FullMatch{
					{
						subString: " ",
						from:      3,
						to:        3,
						expressions: []string{
							"\\s",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: " ",
						from:      5,
						to:        5,
						expressions: []string{
							"\\s",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: " ",
						from:      11,
						to:        11,
						expressions: []string{
							"\\s",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
				},
			},
			{
				name:    "match '\\S'",
				regexps: []string{"\\S"},
				input:   "asd 1 jsdfk 4234!",
				output: []*FullMatch{
					{
						subString: "a",
						from:      0,
						to:        0,
						expressions: []string{
							"\\S",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "s",
						from:      1,
						to:        1,
						expressions: []string{
							"\\S",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "d",
						from:      2,
						to:        2,
						expressions: []string{
							"\\S",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "1",
						from:      4,
						to:        4,
						expressions: []string{
							"\\S",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "j",
						from:      6,
						to:        6,
						expressions: []string{
							"\\S",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "s",
						from:      7,
						to:        7,
						expressions: []string{
							"\\S",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "d",
						from:      8,
						to:        8,
						expressions: []string{
							"\\S",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "f",
						from:      9,
						to:        9,
						expressions: []string{
							"\\S",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "k",
						from:      10,
						to:        10,
						expressions: []string{
							"\\S",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "4",
						from:      12,
						to:        12,
						expressions: []string{
							"\\S",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "2",
						from:      13,
						to:        13,
						expressions: []string{
							"\\S",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "3",
						from:      14,
						to:        14,
						expressions: []string{
							"\\S",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "4",
						from:      15,
						to:        15,
						expressions: []string{
							"\\S",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "!",
						from:      16,
						to:        16,
						expressions: []string{
							"\\S",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
				},
			},
		},
	}

	for groupName, subGroups := range examples {
		t.Run(groupName, func(t *testing.T) {
			for _, ex := range subGroups {
				test := ex
				t.Run(ex.name, func(t *testing.T) {
					tr, err := NewTrie(test.regexps...)
					require.NoError(t, err)

					t.Log(tr.String())
					t.Logf("input: '%s'", string(test.input))

					sort.SliceStable(test.output, func(i, j int) bool {
						return comparator(test.output[i], test.output[j])
					})

					actual := tr.Match(test.input)
					require.NoError(t, err)

					sort.SliceStable(actual, func(i, j int) bool {
						return comparator(actual[i], actual[j])
					})

					require.EqualValues(t, test.output, actual)
				})
			}
		})
	}
}

type example struct {
	name    string
	regexps []string
	input   string
	output  []*FullMatch
}

func comparator(x, y *FullMatch) bool {
	if x.From() != y.From() {
		return x.From() < y.From()
	}

	if x.To() != y.To() {
		return x.To() < y.To()
	}

	return x.String() < y.String()
}
