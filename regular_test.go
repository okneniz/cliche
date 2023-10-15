package regular

import (
	"testing"
)

// https://www.rfc-editor.org/rfc/rfc9485.html#name-implementing-i-regexp

func TestTrie(t *testing.T) {
	tr, err := NewTrie(
		"t",
		"test",
		"tost",
		"tot",
		".",
		"\\d",
		"\\?",
		"\\.",
		"\\D",
		"\\w",
		"\\W",
		"[a-z]",
		"[0-9]",
		"[0-9a-z]",
		"[a-z0-9]",
		"[0-9a-zxy\\d]",
		"[^0-9a-zxy\\d]",
		"(foo)",
		"(f|b)",
		"(foo|bar|baz)",
		"(foo|bar|baz)+",
		"(?:foo|bar|baz)+",
		"(?<x>y)",
		"(?<name>x|y|z)",
		"(?<name>y|x|z)",
		"(?<test>foo|bar|baz)+",
		"(?<test>foo|(ba|za|r)|baz)+",
		"[^abc1-3]?",
		"\\d*",
		"\\S",
		"\\s",
		"^",
		"$",
		"\\A",
		"\\z",
		"a{1}",
		"a{1,}",
		"a{1,3}",
		"a{1,3,}",
		"[a{1,3}bc]+",
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("wtf %#v", tr)
	t.Log(tr)
}
