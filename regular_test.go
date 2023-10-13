package regular

import (
	"testing"
)

func TestTrie(t *testing.T) {
	tr, err := NewTrie(
		"test",
		".",
		"\\d",
		"\\D",
		"\\w",
		"\\W",
		"[a-z]",
		"[0-9]",
		"[0-9a-zxy\\d]",
		"[^0-9a-zxy\\d]",
		"(foo|bar|baz)",
		"(foo|bar|baz)+",
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
}
