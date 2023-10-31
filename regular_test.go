package regular

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// https://www.rfc-editor.org/rfc/rfc9485.html#name-implementing-i-regexp

func TestTrie(t *testing.T) {
	t.Parallel()

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
		// like [aa1-2] == [1-2a]
		// like [121-2] == [1-2]
	})

	// Some quantifiers have the same meaning, but have different symbols.
	// For example:
	// - x+ is equal x{1,}
	// - x* is equal x{0,}
	// - x? is equal x{0,1}
	// - x{1,1} is equal x{1}
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

		err = tr.Add("x{1}")
		require.NoError(t, err)
		require.Equal(t, tr.Size(), 4)

		err = tr.Add("x{1,1}")
		require.NoError(t, err)
		require.Equal(t, tr.Size(), 4)
	})
}

func TestMatch(t *testing.T) {
	// t.Parallel()

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
		"groups": { // TODO : unions
			{
				// cases which brouk something :
				// - (f|b)(o|a)(o|\\w|\\D
				// - (f|b)(o|a)(o|r|z|)

				name: "chars matching and capturing",
				regexps: []string{
					"fo(o|b)",
					"f(o|b)o",
					"(f|b)(o|a)(o|r|z)",
					"(f|b)(o|a)(o|\\w|\\D)",
					"(f)(o)(o)",
				},
				input: "foo bar baz",
				output: []*FullMatch{
					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"fo(o|b)",
						},
						namedGroups: map[string]bounds{},
						groups: []bounds{
							{from: 2, to: 2},
						},
					},
					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"f(o|b)o",
						},
						namedGroups: map[string]bounds{},
						groups: []bounds{
							{from: 1, to: 1},
						},
					},
					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"(f|b)(o|a)(o|r|z)",
						},
						namedGroups: map[string]bounds{},
						groups: []bounds{
							{from: 0, to: 0},
							{from: 1, to: 1},
							{from: 2, to: 2},
						},
					},
					{
						subString: "bar",
						from:      4,
						to:        6,
						expressions: []string{
							"(f|b)(o|a)(o|r|z)",
						},
						namedGroups: map[string]bounds{},
						groups: []bounds{
							{from: 4, to: 4},
							{from: 5, to: 5},
							{from: 6, to: 6},
						},
					},
					{
						subString: "baz",
						from:      8,
						to:        10,
						expressions: []string{
							"(f|b)(o|a)(o|r|z)",
						},
						namedGroups: map[string]bounds{},
						groups: []bounds{
							{from: 8, to: 8},
							{from: 9, to: 9},
							{from: 10, to: 10},
						},
					},
					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"(f|b)(o|a)(o|\\w|\\D)",
						},
						namedGroups: map[string]bounds{},
						groups: []bounds{
							{from: 0, to: 0},
							{from: 1, to: 1},
							{from: 2, to: 2},
						},
					},
					{
						subString: "bar",
						from:      4,
						to:        6,
						expressions: []string{
							"(f|b)(o|a)(o|\\w|\\D)",
						},
						namedGroups: map[string]bounds{},
						groups: []bounds{
							{from: 4, to: 4},
							{from: 5, to: 5},
							{from: 6, to: 6},
						},
					},
					{
						subString: "baz",
						from:      8,
						to:        10,
						expressions: []string{
							"(f|b)(o|a)(o|\\w|\\D)",
						},
						namedGroups: map[string]bounds{},
						groups: []bounds{
							{from: 8, to: 8},
							{from: 9, to: 9},
							{from: 10, to: 10},
						},
					},
					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"(f)(o)(o)",
						},
						namedGroups: map[string]bounds{},
						groups: []bounds{
							{from: 0, to: 0},
							{from: 1, to: 1},
							{from: 2, to: 2},
						},
					},
				},
			},
			{
				name: "chars matching and capturing with nested groups",
				regexps: []string{
					"f(o(o))",
					"(b(a(r)))",
					"((b)az)",
				},
				input: "foo bar baz",
				output: []*FullMatch{
					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"f(o(o))",
						},
						namedGroups: map[string]bounds{},
						groups: []bounds{
							{from: 1, to: 2},
							{from: 2, to: 2},
						},
					},
					{
						subString: "bar",
						from:      4,
						to:        6,
						expressions: []string{
							"(b(a(r)))",
						},
						namedGroups: map[string]bounds{},
						groups: []bounds{
							{from: 4, to: 6},
							{from: 5, to: 6},
							{from: 6, to: 6},
						},
					},
					{
						subString: "baz",
						from:      8,
						to:        10,
						expressions: []string{
							"((b)az)",
						},
						namedGroups: map[string]bounds{},
						groups: []bounds{
							{from: 8, to: 10},
							{from: 8, to: 8},
						},
					},
				},
			},
		},
		"named groups": { // TODO : unions
			{
				name: "strings matching and capturing",
				regexps: []string{
					"fo(?<name>o|b)",
					"f(?<name>o|b)o",
					"(?<first>f|b)(?<second>o|a)(?<third>o|r|z)",
					"(?<first>f|b)(?<second>o|a)(?<third>o|\\w|\\D)",
					"(?<first>f)(?<second>o)(?<third>o)",
				},
				input: "foo bar baz",
				output: []*FullMatch{
					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"fo(?<name>o|b)",
						},
						namedGroups: map[string]bounds{
							"name": {from: 2, to: 2},
						},
						groups: []bounds{},
					},
					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"f(?<name>o|b)o",
						},
						namedGroups: map[string]bounds{
							"name": {from: 1, to: 1},
						},
						groups: []bounds{},
					},
					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"(?<first>f|b)(?<second>o|a)(?<third>o|r|z)",
						},
						namedGroups: map[string]bounds{
							"first":  {from: 0, to: 0},
							"second": {from: 1, to: 1},
							"third":  {from: 2, to: 2},
						},
						groups: []bounds{},
					},
					{
						subString: "bar",
						from:      4,
						to:        6,
						expressions: []string{
							"(?<first>f|b)(?<second>o|a)(?<third>o|r|z)",
						},
						namedGroups: map[string]bounds{
							"first":  {from: 4, to: 4},
							"second": {from: 5, to: 5},
							"third":  {from: 6, to: 6},
						},
						groups: []bounds{},
					},
					{
						subString: "baz",
						from:      8,
						to:        10,
						expressions: []string{
							"(?<first>f|b)(?<second>o|a)(?<third>o|r|z)",
						},
						namedGroups: map[string]bounds{
							"first":  {from: 8, to: 8},
							"second": {from: 9, to: 9},
							"third":  {from: 10, to: 10},
						},
						groups: []bounds{},
					},

					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"(?<first>f|b)(?<second>o|a)(?<third>o|\\w|\\D)",
						},
						namedGroups: map[string]bounds{
							"first":  {from: 0, to: 0},
							"second": {from: 1, to: 1},
							"third":  {from: 2, to: 2},
						},
						groups: []bounds{},
					},
					{
						subString: "bar",
						from:      4,
						to:        6,
						expressions: []string{
							"(?<first>f|b)(?<second>o|a)(?<third>o|\\w|\\D)",
						},
						namedGroups: map[string]bounds{
							"first":  {from: 4, to: 4},
							"second": {from: 5, to: 5},
							"third":  {from: 6, to: 6},
						},
						groups: []bounds{},
					},
					{
						subString: "baz",
						from:      8,
						to:        10,
						expressions: []string{
							"(?<first>f|b)(?<second>o|a)(?<third>o|\\w|\\D)",
						},
						namedGroups: map[string]bounds{
							"first":  {from: 8, to: 8},
							"second": {from: 9, to: 9},
							"third":  {from: 10, to: 10},
						},
						groups: []bounds{},
					},
					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"(?<first>f)(?<second>o)(?<third>o)",
						},
						namedGroups: map[string]bounds{
							"first":  {from: 0, to: 0},
							"second": {from: 1, to: 1},
							"third":  {from: 2, to: 2},
						},
						groups: []bounds{},
					},
				},
			},
			{
				name: "chars matching and capturing with nested groups",
				regexps: []string{
					"f(?<first>o(?<second>o))",
					"(?<first>b(?<second>a(?<third>r)))",
					"(?<first>(?<second>b)az)",
				},
				input: "foo bar baz",
				output: []*FullMatch{
					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"f(?<first>o(?<second>o))",
						},
						namedGroups: map[string]bounds{
							"first":  {from: 1, to: 2},
							"second": {from: 2, to: 2},
						},
						groups: []bounds{},
					},
					{
						subString: "bar",
						from:      4,
						to:        6,
						expressions: []string{
							"(?<first>b(?<second>a(?<third>r)))",
						},
						namedGroups: map[string]bounds{
							"first":  {from: 4, to: 6},
							"second": {from: 5, to: 6},
							"third":  {from: 6, to: 6},
						},
						groups: []bounds{},
					},
					{
						subString: "baz",
						from:      8,
						to:        10,
						expressions: []string{
							"(?<first>(?<second>b)az)",
						},
						namedGroups: map[string]bounds{
							"second": {from: 8, to: 8},
							"first":  {from: 8, to: 10},
						},
						groups: []bounds{},
					},
				},
			},
		},
		"not captured groups": { // TODO : unions
			{
				name: "strings matching and capturing",
				regexps: []string{
					"fo(?:o|b)",
					"f(?:o|b)o",
					"(?:f|b)(?:o|a)(?:o|r|z)",
					"(?:f|b)(?:o|a)(?:o|\\w|\\D)",
					"(?:f)(?:o)(?:o)",
				},
				input: "foo bar baz",
				output: []*FullMatch{
					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"fo(?:o|b)",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"f(?:o|b)o",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"(?:f|b)(?:o|a)(?:o|r|z)",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "bar",
						from:      4,
						to:        6,
						expressions: []string{
							"(?:f|b)(?:o|a)(?:o|r|z)",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "baz",
						from:      8,
						to:        10,
						expressions: []string{
							"(?:f|b)(?:o|a)(?:o|r|z)",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"(?:f|b)(?:o|a)(?:o|\\w|\\D)",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "bar",
						from:      4,
						to:        6,
						expressions: []string{
							"(?:f|b)(?:o|a)(?:o|\\w|\\D)",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "baz",
						from:      8,
						to:        10,
						expressions: []string{
							"(?:f|b)(?:o|a)(?:o|\\w|\\D)",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"(?:f)(?:o)(?:o)",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
				},
			},
			{
				name: "chars matching and capturing with nested groups",
				regexps: []string{
					"f(?:o(?:o))",
					"(?:b(?:a(?:r)))",
					"(?:(?:b)az)",
				},
				input: "foo bar baz",
				output: []*FullMatch{
					{
						subString: "foo",
						from:      0,
						to:        2,
						expressions: []string{
							"f(?:o(?:o))",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "bar",
						from:      4,
						to:        6,
						expressions: []string{
							"(?:b(?:a(?:r)))",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "baz",
						from:      8,
						to:        10,
						expressions: []string{
							"(?:(?:b)az)",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
				},
			},
		},
		"quantifiers": {
			{
				name: "optional",
				regexps: []string{
					"c?",
					"pics?",
					"pi.?c",
					"....?",
					"...?.",
				},
				input: "pic",
				output: []*FullMatch{
					{
						subString: "",
						from:      0,
						to:        0,
						expressions: []string{
							"c?",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty:       true,
					},
					{
						subString: "",
						from:      1,
						to:        1,
						expressions: []string{
							"c?",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty:       true,
					},
					{
						subString: "c",
						from:      2,
						to:        2,
						expressions: []string{
							"c?",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "pic",
						from:      0,
						to:        2,
						expressions: []string{
							"pics?",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "pic",
						from:      0,
						to:        2,
						expressions: []string{
							"pi.?c",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "pic",
						from:      0,
						to:        2,
						expressions: []string{
							"....?",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "pic",
						from:      0,
						to:        2,
						expressions: []string{
							"...?.",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
				},
			},
			{
				name: "zero or more, or '*'",
				regexps: []string{
					"x*",
					"x*x",
					"x.*",
					"x{0,}",
					"x{0,}x",
					"x.{0,}",
				},
				input: "xx x\n x",
				output: []*FullMatch{
					{
						subString: "xx",
						from:      0,
						to:        1,
						expressions: []string{
							"x*",
							"x{0,}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "",
						from:      2,
						to:        2,
						expressions: []string{
							"x*",
							"x{0,}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: true,
					},
					{
						subString: "x",
						from:      3,
						to:        3,
						expressions: []string{
							"x*",
							"x{0,}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
					{
						subString: "",
						from:      4,
						to:        4,
						expressions: []string{
							"x*",
							"x{0,}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: true,
					},
					{
						subString: "",
						from:      5,
						to:        5,
						expressions: []string{
							"x*",
							"x{0,}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: true,
					},
					{
						subString: "x",
						from:      6,
						to:        6,
						expressions: []string{
							"x*",
							"x{0,}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
					{
						subString: "xx x",
						from:      0,
						to:        3,
						expressions: []string{
							"x.*",
							"x.{0,}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "x",
						from:      6,
						to:        6,
						expressions: []string{
							"x.*",
							"x.{0,}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "xx",
						from:      0,
						to:        1,
						expressions: []string{
							"x*x",
							"x{0,}x",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
					},
					{
						subString: "x",
						from:      3,
						to:        3,
						expressions: []string{
							"x*x",
							"x{0,}x",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
					{
						subString: "x",
						from:      6,
						to:        6,
						expressions: []string{
							"x*x",
							"x{0,}x",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
				},
			},
			{
				name: "one or more, or '+'",
				regexps: []string{
					"x+",
					"x+x",
					"x.+",
					"x{1,}",
					"x{1,}x",
					"x.{1,}",
				},
				input: "xx x\n x",
				output: []*FullMatch{
					{
						subString: "xx",
						from:      0,
						to:        1,
						expressions: []string{
							"x+",
							"x{1,}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
					{
						subString: "x",
						from:      3,
						to:        3,
						expressions: []string{
							"x+",
							"x{1,}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
					{
						subString: "x",
						from:      6,
						to:        6,
						expressions: []string{
							"x+",
							"x{1,}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
					{
						subString: "xx",
						from:      0,
						to:        1	,
						expressions: []string{
							"x+x",
							"x{1,}x",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
					{
						subString: "xx x",
						from:      0,
						to:        3	,
						expressions: []string{
							"x.+",
							"x.{1,}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
				},
			},
			{
				name: "endless quantifier",
				regexps: []string{
					"x{2,}",
				},
				input: "xx xxx x",
				output: []*FullMatch{
					{
						subString: "xx",
						from:      0,
						to:        1	,
						expressions: []string{
							"x{2,}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
					{
						subString: "xxx",
						from:      3,
						to:        5	,
						expressions: []string{
							"x{2,}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
				},
			},
			{
				name: "limited quantifier",
				regexps: []string{
					"x{2,4}",
				},
				input: "xx xxx x xxxxxx",
				output: []*FullMatch{
					{
						subString: "xx",
						from:      0,
						to:        1	,
						expressions: []string{
							"x{2,4}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
					{
						subString: "xxx",
						from:      3,
						to:        5	,
						expressions: []string{
							"x{2,4}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
					{
						subString: "xxxx",
						from:      9,
						to:        12	,
						expressions: []string{
							"x{2,4}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
					{
						subString: "xx",
						from:      13,
						to:        14	,
						expressions: []string{
							"x{2,4}",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
				},
			},
		},
		"start of": {
			{
				name: "line",
				regexps: []string{
					"^...",
					"^.",
					".^",
				},
				input: "foo bar\nbaz",
				output: []*FullMatch{
					{
						subString: "foo",
						from:      0,
						to:        2	,
						expressions: []string{
							"^...",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
					{
						subString: "baz",
						from:      8,
						to:        10	,
						expressions: []string{
							"^...",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
					{
						subString: "f",
						from:      0,
						to:        0	,
						expressions: []string{
							"^.",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
					{
						subString: "b",
						from:      8,
						to:        8	,
						expressions: []string{
							"^.",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
				},
			},
		},
		"end of": {
			{
				name: "line",
				regexps: []string{
					"...$",
					// ".$", TODO : fix conflict with upper regexp
					"$.",
					"$",
				},
				input: "foo bar\nbaz",
				output: []*FullMatch{
					{
						subString: "bar",
						from:      4,
						to:        7,
						expressions: []string{
							"...$",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: false,
					},
					// {
					// 	subString: "r",
					// 	from:      6,
					// 	to:        7,
					// 	expressions: []string{
					// 		".$",
					// 	},
					// 	namedGroups: map[string]bounds{},
					// 	groups:      []bounds{},
					// 	empty: false,
					// },
					{
						subString: "",
						from:      7,
						to:        7	,
						expressions: []string{
							"$",
						},
						namedGroups: map[string]bounds{},
						groups:      []bounds{},
						empty: true,
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

					if len(test.output) != len(actual) {
						require.Equal(t, test.output, actual)
					}

					for i := range test.output {
						sort.SliceStable(test.output[i].expressions, func(x, y int) bool {
							return test.output[i].expressions[x] < test.output[i].expressions[y]
						})

						sort.SliceStable(actual[i].expressions, func(x, y int) bool {
							return actual[i].expressions[x] < actual[i].expressions[y]
						})

						require.Equalf(t, *test.output[i], *actual[i], "compare %d match", i)
					}
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

	if x.String() != y.String() {
		return x.String() < y.String()
	}

	sort.Slice(x.expressions, func(i, j int) bool {
		return x.expressions[i] < x.expressions[j]
	})

	sort.Slice(y.expressions, func(i, j int) bool {
		return y.expressions[i] < y.expressions[j]
	})

	return strings.Join(x.expressions, ", ") < strings.Join(y.expressions, ", ")
}

func TestQuantifierBounds(t *testing.T) {
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

func TestQuantifier_getKey(t *testing.T) {
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
			expression: "x{2,}",
			key:        "x{2,}",
			from:       2,
			to:         nil,
			more:       true,
		},
		{
			expression: "x{2}",
			key:        "x{2}",
			from:       2,
			to:         nil,
			more:       false,
		},
		{
			expression:    "x{3,2}",
			notQuantifier: true,
		},
		{
			expression:    "x{,2}",
			notQuantifier: true,
		},
		{
			expression:    "x{2,2,}",
			notQuantifier: true,
		},
		{
			expression:    "x{0}",
			notQuantifier: true,
		},
		{
			expression:    "x{0,0}",
			notQuantifier: true,
		},
	}

	parse := parseRegexp()

	for _, test := range examples {
		t.Run(test.expression, func(t *testing.T) {
			input := newBuffer(test.expression)

			output, err := parse(input)
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

				if q.getKey() != test.key {
					t.Fatalf(
						"expected key %v, actual key %v",
						q.getKey(),
						test.key,
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

type num int

func(t num) String() string {
	return fmt.Sprintf("%d", t)
}

func Test_list(t *testing.T) {
	l := newList[num](2)

	require.Equal(t, 0, l.len())
	require.Nil(t, l.first())
	require.Nil(t, l.last())
	require.Nil(t, l.toSlice())
	require.Equal(t, l.String(), "[]")

	l.append(1)

	require.Equal(t, 1, l.len())
	require.Equal(t, *l.first(), num(1))
	require.Equal(t, *l.last(), num(1))
	require.Equal(t, l.toSlice(), []num{1})
	require.Equal(t, l.String(), "[1]")

	l.append(2)

	require.Equal(t, 2, l.len())
	require.Equal(t, *l.first(), num(1))
	require.Equal(t, *l.last(), num(2))
	require.Equal(t, l.toSlice(), []num{1, 2})
	require.Equal(t, l.String(), "[1, 2]")

	l.append(3)

	require.Equal(t, 3, l.len())
	require.Equal(t, *l.first(), num(1))
	require.Equal(t, *l.last(), num(3))
	require.Equal(t, l.toSlice(), []num{1, 2, 3})
	require.Equal(t, l.String(), "[1, 2, 3]")

	l.append(4)

	require.Equal(t, 4, l.len())
	require.Equal(t, *l.first(), num(1))
	require.Equal(t, *l.last(), num(4))
	require.Equal(t, l.toSlice(), []num{1, 2, 3, 4})
	require.Equal(t, l.String(), "[1, 2, 3, 4]")

	l.truncate(2)

	require.Equal(t, 2, l.len())
	require.Equal(t, *l.first(), num(1))
	require.Equal(t, *l.last(), num(2))
	require.Equal(t, l.toSlice(), []num{1, 2})
	require.Equal(t, l.String(), "[1, 2]")

	l.append(10)

	require.Equal(t, 3, l.len())
	require.Equal(t, *l.first(), num(1))
	require.Equal(t, *l.last(), num(10))
	require.Equal(t, l.toSlice(), []num{1, 2, 10})
	require.Equal(t, l.String(), "[1, 2, 10]")

	l.truncate(0)

	require.Equal(t, 0, l.len())
	require.Nil(t, l.first())
	require.Nil(t, l.last())
	require.Nil(t, l.toSlice())
	require.Equal(t, l.String(), "[]")

	l.append(10)
	l.append(20)
	l.append(30)

	require.Equal(t, 3, l.len())
	require.Equal(t, *l.first(), num(10))
	require.Equal(t, *l.last(), num(30))
	require.Equal(t, l.toSlice(), []num{10, 20, 30})
	require.Equal(t, l.String(), "[10, 20, 30]")

	require.PanicsWithValue(
		t,
		OutOfBounds{Min: 0, Max: 3,Value: -1},
		func() { l.truncate(-1) },
	)

	require.PanicsWithValue(
		t,
		OutOfBounds{Min: 0, Max: 3,Value: 10},
		func() { l.truncate(10) },
	)

	l.truncate(0)

	require.PanicsWithValue(
		t,
		OutOfBounds{Min: 0, Max: 0,Value: -1},
		func() { l.truncate(-1) },
	)

	require.PanicsWithValue(
		t,
		OutOfBounds{Min: 0, Max: 0,Value: 10},
		func() { l.truncate(10) },
	)

	l.truncate(0)
}

func Test_OutOfBounds(t *testing.T) {
	err := OutOfBounds{
		Min: 10,
		Max: 100,
		Value: -50,
	}

	require.Equal(t, err.Error(), "-50 is ouf of bounds 10..100")
}
