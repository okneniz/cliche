package regular

import (
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
	tr, err := NewTrie()
	require.NoError(t, err)
	require.Equal(t, tr.Size(), 0)

	err = tr.Add("x")
	require.NoError(t, err)

	t.Log("trie", tr)

	result, err := tr.Match("x")
	require.NoError(t, err)

	t.Logf("result 1 %v", result)

	result, err = tr.Match("xxx")
	require.NoError(t, err)

	t.Logf("result 2 %v", result)
}
