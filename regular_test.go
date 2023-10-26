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
	t.Run("simple", func(t *testing.T) {
		tr, err := NewTrie(
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
		)
		require.NoError(t, err)

		t.Log(tr.String())

		examples := map[string][]*FullMatch{
			"testing string test ssss word words": {

				// [te test s ing s string ing te test s s s s s word word s]
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
		}

		for in, ex := range examples {
			expected := ex
			input := in

			sort.SliceStable(expected, func(i, j int) bool {
				return comparator(expected[i], expected[j])
			})

			t.Run(input, func(t *testing.T) {
				t.Logf("input: '%s'", string(input))

				actual := tr.Match(input)
				require.NoError(t, err)

				sort.SliceStable(actual, func(i, j int) bool {
					return comparator(actual[i], actual[j])
				})

				require.EqualValues(t, expected, actual)
			})
		}
	})
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
