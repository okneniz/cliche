package regular

import (
	"sort"
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
		"^",
		"$",
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

	t.Log(tr)
}

func TestMatch(t *testing.T) {
	t.Parallel()

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
				output: []*stringMatch{
					{
						subString: "te",
						span: span{
							from: 0,
							to:   1,
						},
						expressions: newDict(
							"te",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "test",
						span: span{
							from: 0,
							to:   3,
						},
						expressions: newDict(
							"test",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "s",
						span: span{
							from: 2,
							to:   2,
						},
						expressions: newDict(
							"s",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "ing",
						span: span{
							from: 4,
							to:   6,
						},
						expressions: newDict(
							"ing",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "s",
						span: span{
							from: 8,
							to:   8,
						},
						expressions: newDict(
							"s",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "string",
						span: span{
							from: 8,
							to:   13,
						},
						expressions: newDict(
							"string",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "ing",
						span: span{
							from: 11,
							to:   13,
						},
						expressions: newDict(
							"ing",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "te",
						span: span{
							from: 15,
							to:   16,
						},
						expressions: newDict(
							"te",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "test",
						span: span{
							from: 15,
							to:   18,
						},
						expressions: newDict(
							"test",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "s",
						span: span{
							from: 17,
							to:   17,
						},
						expressions: newDict(
							"s",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "s",
						span: span{
							from: 20,
							to:   20,
						},
						expressions: newDict(
							"s",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "s",
						span: span{
							from: 21,
							to:   21,
						},
						expressions: newDict(
							"s",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "s",
						span: span{
							from: 22,
							to:   22,
						},
						expressions: newDict(
							"s",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "s",
						span: span{
							from: 23,
							to:   23,
						},
						expressions: newDict(
							"s",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "word",
						span: span{
							from: 25,
							to:   28,
						},
						expressions: newDict(
							"word",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "word",
						span: span{
							from: 30,
							to:   33,
						},
						expressions: newDict(
							"word",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "s",
						span: span{
							from: 34,
							to:   34,
						},
						expressions: newDict(
							"s",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
			{
				name:    "match '.'",
				regexps: []string{"t."},
				input:   "testing string test ssss word words",
				output: []*stringMatch{
					{
						subString: "te",
						span: span{
							from: 0,
							to:   1,
						},
						expressions: newDict(
							"t.",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "ti",
						span: span{
							from: 3,
							to:   4,
						},
						expressions: newDict(
							"t.",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "tr",
						span: span{
							from: 9,
							to:   10,
						},
						expressions: newDict(
							"t.",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "te",
						span: span{
							from: 15,
							to:   16,
						},
						expressions: newDict(
							"t.",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "t ",
						span: span{
							from: 18,
							to:   19,
						},
						expressions: newDict(
							"t.",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
			{
				name:    "match multiple '.'",
				regexps: []string{"t.."},
				input:   "testing string test ssss word words",
				output: []*stringMatch{
					{
						subString: "tes",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"t..",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "tin",
						span: span{
							from: 3,
							to:   5,
						},
						expressions: newDict(
							"t..",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "tri",
						span: span{
							from: 9,
							to:   11,
						},
						expressions: newDict(
							"t..",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "tes",
						span: span{
							from: 15,
							to:   17,
						},
						expressions: newDict(
							"t..",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "t s",
						span: span{
							from: 18,
							to:   20,
						},
						expressions: newDict(
							"t..",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
			{
				name:    "match '\\d'",
				regexps: []string{"\\d"},
				input:   "asd 1 jsdfk 4234",
				output: []*stringMatch{
					{
						subString: "1",
						span: span{
							from: 4,
							to:   4,
						},
						expressions: newDict(
							"\\d",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "4",
						span: span{
							from: 12,
							to:   12,
						},
						expressions: newDict(
							"\\d",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "2",
						span: span{
							from: 13,
							to:   13,
						},
						expressions: newDict(
							"\\d",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "3",
						span: span{
							from: 14,
							to:   14,
						},
						expressions: newDict(
							"\\d",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "4",
						span: span{
							from: 15,
							to:   15,
						},
						expressions: newDict(
							"\\d",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
			{
				name:    "match '\\D'",
				regexps: []string{"\\D"},
				input:   "asd 1 jsdfk 4234",
				output: []*stringMatch{
					{
						subString: "a",
						span: span{
							from: 0,
							to:   0,
						},
						expressions: newDict(
							"\\D",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "s",
						span: span{
							from: 1,
							to:   1,
						},
						expressions: newDict(
							"\\D",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "d",
						span: span{
							from: 2,
							to:   2,
						},
						expressions: newDict(
							"\\D",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: " ",
						span: span{
							from: 3,
							to:   3,
						},
						expressions: newDict(
							"\\D",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: " ",
						span: span{
							from: 5,
							to:   5,
						},
						expressions: newDict(
							"\\D",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "j",
						span: span{
							from: 6,
							to:   6,
						},
						expressions: newDict(
							"\\D",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "s",
						span: span{
							from: 7,
							to:   7,
						},
						expressions: newDict(
							"\\D",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "d",
						span: span{
							from: 8,
							to:   8,
						},
						expressions: newDict(
							"\\D",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "f",
						span: span{
							from: 9,
							to:   9,
						},
						expressions: newDict(
							"\\D",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "k",
						span: span{
							from: 10,
							to:   10,
						},
						expressions: newDict(
							"\\D",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: " ",
						span: span{
							from: 11,
							to:   11,
						},
						expressions: newDict(
							"\\D",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
			{
				name:    "match '\\w'",
				regexps: []string{"\\w"},
				input:   "asd 1 jsdfk 4234",
				output: []*stringMatch{
					{
						subString: "a",
						span: span{
							from: 0,
							to:   0,
						},
						expressions: newDict(
							"\\w",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "s",
						span: span{
							from: 1,
							to:   1,
						},
						expressions: newDict(
							"\\w",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "d",
						span: span{
							from: 2,
							to:   2,
						},
						expressions: newDict(
							"\\w",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "1",
						span: span{
							from: 4,
							to:   4,
						},
						expressions: newDict(
							"\\w",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "j",
						span: span{
							from: 6,
							to:   6,
						},
						expressions: newDict(
							"\\w",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "s",
						span: span{
							from: 7,
							to:   7,
						},
						expressions: newDict(
							"\\w",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "d",
						span: span{
							from: 8,
							to:   8,
						},
						expressions: newDict(
							"\\w",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "f",
						span: span{
							from: 9,
							to:   9,
						},
						expressions: newDict(
							"\\w",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "k",
						span: span{
							from: 10,
							to:   10,
						},
						expressions: newDict(
							"\\w",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "4",
						span: span{
							from: 12,
							to:   12,
						},
						expressions: newDict(
							"\\w",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "2",
						span: span{
							from: 13,
							to:   13,
						},
						expressions: newDict(
							"\\w",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "3",
						span: span{
							from: 14,
							to:   14,
						},
						expressions: newDict(
							"\\w",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "4",
						span: span{
							from: 15,
							to:   15,
						},
						expressions: newDict(
							"\\w",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
			{
				name:    "match '\\W'",
				regexps: []string{"\\W"},
				input:   "asd 1 jsdfk 4234!\n\r",
				output: []*stringMatch{
					{
						subString: " ",
						span: span{
							from: 3,
							to:   3,
						},
						expressions: newDict(
							"\\W",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: " ",
						span: span{
							from: 5,
							to:   5,
						},
						expressions: newDict(
							"\\W",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: " ",
						span: span{
							from: 11,
							to:   11,
						},
						expressions: newDict(
							"\\W",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "!",
						span: span{
							from: 16,
							to:   16,
						},
						expressions: newDict(
							"\\W",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "\n",
						span: span{
							from: 17,
							to:   17,
						},
						expressions: newDict(
							"\\W",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "\r",
						span: span{
							from: 18,
							to:   18,
						},
						expressions: newDict(
							"\\W",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
			{
				name:    "match '\\s'",
				regexps: []string{"\\s"},
				input:   "asd 1 jsdfk 4234",
				output: []*stringMatch{
					{
						subString: " ",
						span: span{
							from: 3,
							to:   3,
						},
						expressions: newDict(
							"\\s",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: " ",
						span: span{
							from: 5,
							to:   5,
						},
						expressions: newDict(
							"\\s",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: " ",
						span: span{
							from: 11,
							to:   11,
						},
						expressions: newDict(
							"\\s",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
			{
				name:    "match '\\S'",
				regexps: []string{"\\S"},
				input:   "asd 1 jsdfk 4234!",
				output: []*stringMatch{
					{
						subString: "a",
						span: span{
							from: 0,
							to:   0,
						},
						expressions: newDict(
							"\\S",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "s",
						span: span{
							from: 1,
							to:   1,
						},
						expressions: newDict(
							"\\S",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "d",
						span: span{
							from: 2,
							to:   2,
						},
						expressions: newDict(
							"\\S",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "1",
						span: span{
							from: 4,
							to:   4,
						},
						expressions: newDict(
							"\\S",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "j",
						span: span{
							from: 6,
							to:   6,
						},
						expressions: newDict(
							"\\S",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "s",
						span: span{
							from: 7,
							to:   7,
						},
						expressions: newDict(
							"\\S",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "d",
						span: span{
							from: 8,
							to:   8,
						},
						expressions: newDict(
							"\\S",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "f",
						span: span{
							from: 9,
							to:   9,
						},
						expressions: newDict(
							"\\S",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "k",
						span: span{
							from: 10,
							to:   10,
						},
						expressions: newDict(
							"\\S",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "4",
						span: span{
							from: 12,
							to:   12,
						},
						expressions: newDict(
							"\\S",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "2",
						span: span{
							from: 13,
							to:   13,
						},
						expressions: newDict(
							"\\S",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "3",
						span: span{
							from: 14,
							to:   14,
						},
						expressions: newDict(
							"\\S",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "4",
						span: span{
							from: 15,
							to:   15,
						},
						expressions: newDict(
							"\\S",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "!",
						span: span{
							from: 16,
							to:   16,
						},
						expressions: newDict(
							"\\S",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
			{
				name: "escaped characters",
				regexps: []string{
					"\\.",
					"\\?",
					"\\+",
					"\\*",
					"\\^",
					"\\$",
					"\\[",
					"\\]",
					"\\{",
					"\\}",
				},
				input: ". ? + * ^ $ [ ] { }",
				output: []*stringMatch{
					{
						subString: ".",
						span: span{
							from: 0,
							to:   0,
						},
						expressions: newDict(
							"\\.",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "?",
						span: span{
							from: 2,
							to:   2,
						},
						expressions: newDict(
							"\\?",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "+",
						span: span{
							from: 4,
							to:   4,
						},
						expressions: newDict(
							"\\+",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "*",
						span: span{
							from: 6,
							to:   6,
						},
						expressions: newDict(
							"\\*",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "^",
						span: span{
							from: 8,
							to:   8,
						},
						expressions: newDict(
							"\\^",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "$",
						span: span{
							from: 10,
							to:   10,
						},
						expressions: newDict(
							"\\$",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "[",
						span: span{
							from: 12,
							to:   12,
						},
						expressions: newDict(
							"\\[",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "]",
						span: span{
							from: 14,
							to:   14,
						},
						expressions: newDict(
							"\\]",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "{",
						span: span{
							from: 16,
							to:   16,
						},
						expressions: newDict(
							"\\{",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "}",
						span: span{
							from: 18,
							to:   18,
						},
						expressions: newDict(
							"\\}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
		},
		"unions": {
			{
				name: "chars matching and capturing",
				regexps: []string{
					"fo(o|b)",
					"f(o|b)o",
					"(f|b)(o|a)(o|r|z)",
					"(f|b)(o|a)(o|\\w|\\D)",
					"(f)(o)(o)",
				},
				input: "foo bar baz",
				output: []*stringMatch{
					{
						subString: "foo",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"fo(o|b)",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 2, to: 2},
						},
					},
					{
						subString: "foo",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"f(o|b)o",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 1, to: 1},
						},
					},
					{
						subString: "foo",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"(f|b)(o|a)(o|r|z)",
							"(f|b)(o|a)(o|\\w|\\D)",
							"(f)(o)(o)",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 0, to: 0},
							{from: 1, to: 1},
							{from: 2, to: 2},
						},
					},
					{
						subString: "bar",
						span: span{
							from: 4,
							to:   6,
						},
						expressions: newDict(
							"(f|b)(o|a)(o|r|z)",
							"(f|b)(o|a)(o|\\w|\\D)",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 4, to: 4},
							{from: 5, to: 5},
							{from: 6, to: 6},
						},
					},
					{
						subString: "baz",
						span: span{
							from: 8,
							to:   10,
						},
						expressions: newDict(
							"(f|b)(o|a)(o|r|z)",
							"(f|b)(o|a)(o|\\w|\\D)",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 8, to: 8},
							{from: 9, to: 9},
							{from: 10, to: 10},
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
				output: []*stringMatch{
					{
						subString: "foo",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"f(o(o))",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 1, to: 2},
							{from: 2, to: 2},
						},
					},
					{
						subString: "bar",
						span: span{
							from: 4,
							to:   6,
						},
						expressions: newDict(
							"(b(a(r)))",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 4, to: 6},
							{from: 5, to: 6},
							{from: 6, to: 6},
						},
					},
					{
						subString: "baz",
						span: span{
							from: 8,
							to:   10,
						},
						expressions: newDict(
							"((b)az)",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 8, to: 10},
							{from: 8, to: 8},
						},
					},
				},
			},
		},
		"named groups": {
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
				output: []*stringMatch{
					{
						subString: "foo",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"fo(?<name>o|b)",
						),
						namedGroups: map[string]span{
							"name": {from: 2, to: 2},
						},
						groups: []span{},
					},
					{
						subString: "foo",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"f(?<name>o|b)o",
						),
						namedGroups: map[string]span{
							"name": {from: 1, to: 1},
						},
						groups: []span{},
					},
					{
						subString: "foo",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"(?<first>f|b)(?<second>o|a)(?<third>o|r|z)",
							"(?<first>f|b)(?<second>o|a)(?<third>o|\\w|\\D)",
							"(?<first>f)(?<second>o)(?<third>o)",
						),
						namedGroups: map[string]span{
							"first":  {from: 0, to: 0},
							"second": {from: 1, to: 1},
							"third":  {from: 2, to: 2},
						},
						groups: []span{},
					},
					{
						subString: "bar",
						span: span{
							from: 4,
							to:   6,
						},
						expressions: newDict(
							"(?<first>f|b)(?<second>o|a)(?<third>o|r|z)",
							"(?<first>f|b)(?<second>o|a)(?<third>o|\\w|\\D)",
						),
						namedGroups: map[string]span{
							"first":  {from: 4, to: 4},
							"second": {from: 5, to: 5},
							"third":  {from: 6, to: 6},
						},
						groups: []span{},
					},
					{
						subString: "baz",
						span: span{
							from: 8,
							to:   10,
						},
						expressions: newDict(
							"(?<first>f|b)(?<second>o|a)(?<third>o|r|z)",
							"(?<first>f|b)(?<second>o|a)(?<third>o|\\w|\\D)",
						),
						namedGroups: map[string]span{
							"first":  {from: 8, to: 8},
							"second": {from: 9, to: 9},
							"third":  {from: 10, to: 10},
						},
						groups: []span{},
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
				output: []*stringMatch{
					{
						subString: "foo",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"f(?<first>o(?<second>o))",
						),
						namedGroups: map[string]span{
							"first":  {from: 1, to: 2},
							"second": {from: 2, to: 2},
						},
						groups: []span{},
					},
					{
						subString: "bar",
						span: span{
							from: 4,
							to:   6,
						},
						expressions: newDict(
							"(?<first>b(?<second>a(?<third>r)))",
						),
						namedGroups: map[string]span{
							"first":  {from: 4, to: 6},
							"second": {from: 5, to: 6},
							"third":  {from: 6, to: 6},
						},
						groups: []span{},
					},
					{
						subString: "baz",
						span: span{
							from: 8,
							to:   10,
						},
						expressions: newDict(
							"(?<first>(?<second>b)az)",
						),
						namedGroups: map[string]span{
							"second": {from: 8, to: 8},
							"first":  {from: 8, to: 10},
						},
						groups: []span{},
					},
				},
			},
		},
		"not captured groups": {
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
				output: []*stringMatch{
					{
						subString: "foo",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"fo(?:o|b)",
							"f(?:o|b)o",
							"(?:f|b)(?:o|a)(?:o|r|z)",
							"(?:f|b)(?:o|a)(?:o|\\w|\\D)",
							"(?:f)(?:o)(?:o)",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "bar",
						span: span{
							from: 4,
							to:   6,
						},
						expressions: newDict(
							"(?:f|b)(?:o|a)(?:o|r|z)",
							"(?:f|b)(?:o|a)(?:o|\\w|\\D)",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "baz",
						span: span{
							from: 8,
							to:   10,
						},
						expressions: newDict(
							"(?:f|b)(?:o|a)(?:o|r|z)",
							"(?:f|b)(?:o|a)(?:o|\\w|\\D)",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
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
				output: []*stringMatch{
					{
						subString: "foo",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"f(?:o(?:o))",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "bar",
						span: span{
							from: 4,
							to:   6,
						},
						expressions: newDict(
							"(?:b(?:a(?:r)))",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "baz",
						span: span{
							from: 8,
							to:   10,
						},
						expressions: newDict(
							"(?:(?:b)az)",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
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
				output: []*stringMatch{
					{
						subString: "",
						span: span{
							from:  0,
							to:    0,
							empty: true,
						},
						expressions: newDict(
							"c?",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "",
						span: span{
							from:  1,
							to:    1,
							empty: true,
						},
						expressions: newDict(
							"c?",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "c",
						span: span{
							from: 2,
							to:   2,
						},
						expressions: newDict(
							"c?",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					// TODO : check c? in rubular again (it's want end of string too)
					{
						subString: "pic",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"pics?",
							"pi.?c",
							"....?",
							"...?.",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
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
				output: []*stringMatch{
					{
						subString: "xx",
						span: span{
							from: 0,
							to:   1,
						},
						expressions: newDict(
							"x*",
							"x*x",
							"x{0,}",
							"x{0,}x",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "",
						span: span{
							from:  2,
							to:    2,
							empty: true,
						},
						expressions: newDict(
							"x*",
							"x{0,}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "x",
						span: span{
							from: 3,
							to:   3,
						},
						expressions: newDict(
							"x*",
							"x*x",
							"x{0,}",
							"x{0,}x",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "xx x",
						span: span{
							from: 0,
							to:   3,
						},
						expressions: newDict(
							"x.*",
							"x.{0,}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "",
						span: span{
							from:  4,
							to:    4,
							empty: true,
						},
						expressions: newDict(
							"x*",
							"x{0,}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "",
						span: span{
							from:  5,
							to:    5,
							empty: true,
						},
						expressions: newDict(
							"x*",
							"x{0,}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "x",
						span: span{
							from: 6,
							to:   6,
						},
						expressions: newDict(
							"x*",
							"x*x",
							"x.*",
							"x{0,}",
							"x{0,}x",
							"x.{0,}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
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
				output: []*stringMatch{
					{
						subString: "xx",
						span: span{
							from: 0,
							to:   1,
						},
						expressions: newDict(
							"x+",
							"x{1,}",
							"x+x",
							"x{1,}x",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "x",
						span: span{
							from: 3,
							to:   3,
						},
						expressions: newDict(
							"x+",
							"x{1,}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "x",
						span: span{
							from: 6,
							to:   6,
						},
						expressions: newDict(
							"x+",
							"x{1,}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "xx x",
						span: span{
							from: 0,
							to:   3,
						},
						expressions: newDict(
							"x.+",
							"x.{1,}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
			{
				name: "endless quantifier",
				regexps: []string{
					"x{2,}",
				},
				input: "xx xxx x",
				output: []*stringMatch{
					{
						subString: "xx",
						span: span{
							from: 0,
							to:   1,
						},
						expressions: newDict(
							"x{2,}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "xxx",
						span: span{
							from: 3,
							to:   5,
						},
						expressions: newDict(
							"x{2,}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
			{
				name: "limited quantifier",
				regexps: []string{
					"x{2,4}",
				},
				input: "xx xxx x xxxxxx",
				output: []*stringMatch{
					{
						subString: "xx",
						span: span{
							from: 0,
							to:   1,
						},
						expressions: newDict(
							"x{2,4}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "xxx",
						span: span{
							from: 3,
							to:   5,
						},
						expressions: newDict(
							"x{2,4}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "xxxx",
						span: span{
							from: 9,
							to:   12,
						},
						expressions: newDict(
							"x{2,4}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "xx",
						span: span{
							from: 13,
							to:   14,
						},
						expressions: newDict(
							"x{2,4}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
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
				output: []*stringMatch{
					{
						subString: "foo",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"^...",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "baz",
						span: span{
							from: 8,
							to:   10,
						},
						expressions: newDict(
							"^...",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "f",
						span: span{
							from: 0,
							to:   0,
						},
						expressions: newDict(
							"^.",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "b",
						span: span{
							from: 8,
							to:   8,
						},
						expressions: newDict(
							"^.",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
			{
				name: "string",
				regexps: []string{
					"\\A...",
					"\\A.",
					"\\A",
					".\\A",
				},
				input: "foo bar\nbaz",
				output: []*stringMatch{
					{
						subString: "foo",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"\\A...",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "f",
						span: span{
							from: 0,
							to:   0,
						},
						expressions: newDict(
							"\\A.",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "",
						span: span{
							from:  0,
							to:    0,
							empty: true,
						},
						expressions: newDict(
							"\\A",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
		},
		"end of": {
			{
				name: "line",
				regexps: []string{
					"...$",
					".$",
					"$.",
					"$",
				},
				input: "foo bar\nbaz",
				output: []*stringMatch{
					{
						subString: "bar",
						span: span{
							from: 4,
							to:   6,
						},
						expressions: newDict(
							"...$",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "baz",
						span: span{
							from: 8,
							to:   10,
						},
						expressions: newDict(
							"...$",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "r",
						span: span{
							from: 6,
							to:   6,
						},
						expressions: newDict(
							".$",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "",
						span: span{
							from:  7,
							to:    7,
							empty: true,
						},
						expressions: newDict(
							"$",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "z",
						span: span{
							from: 10,
							to:   10,
						},
						expressions: newDict(
							".$",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					// TODO: match or not?
					// {
					// 	subString: "",
					// 	span: span{
					// 		from:  11,
					// 		to:    11,
					// 		empty: true,
					// 	},
					// 	expressions: newDict(
					// 		"$",
					// 	),
					// 	namedGroups: map[string]span{},
					// 	groups:      []span{},
					// },
				},
			},
			{
				name: "string",
				regexps: []string{
					"...\\z",
					// ".\\z", // TODO : fix conflict with upper
					// "\\z", // TODO : should be matched?
					// ".\\z",
				},
				input: "foo bar\nbaz",
				output: []*stringMatch{
					// {
					// 	subString: "z",
					// 	from:      10,
					// 	to:        10,
					// 	expressions: []string{
					// 		".\\z",
					// 	},
					// 	namedGroups: map[string]span{},
					// 	groups:      []span{},
					// 	empty: false,
					// },
					{
						subString: "baz",
						span: span{
							from: 8,
							to:   10,
						},
						expressions: newDict(
							"...\\z",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
		},
		"character classes": {
			{
				name: "positive",
				regexps: []string{
					"[0-9]",
					"[0-9]+",
					"ba[rz]",
					"[faborz]+",
					"[bar][bar][baz]",
				},
				input: "foo 1 bar\nbaz 123",
				output: []*stringMatch{
					{
						subString: "1",
						span: span{
							from: 4,
							to:   4,
						},
						expressions: newDict(
							"[0-9]",
							"[0-9]+",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "1",
						span: span{
							from: 14,
							to:   14,
						},
						expressions: newDict(
							"[0-9]",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "2",
						span: span{
							from: 15,
							to:   15,
						},
						expressions: newDict(
							"[0-9]",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "3",
						span: span{
							from: 16,
							to:   16,
						},
						expressions: newDict(
							"[0-9]",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "123",
						span: span{
							from: 14,
							to:   16,
						},
						expressions: newDict(
							"[0-9]+",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "bar",
						span: span{
							from: 6,
							to:   8,
						},
						expressions: newDict(
							"ba[rz]",
							"[faborz]+",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "baz",
						span: span{
							from: 10,
							to:   12,
						},
						expressions: newDict(
							"ba[rz]",
							"[faborz]+",
							"[bar][bar][baz]",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "foo",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"[faborz]+",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
			{
				name: "negated",
				regexps: []string{
					"[^a-z]",
					"[^\\s]+",
					"ba[^for]",
					"[^\\s][^\\s][^\\s]",
				},
				input: "foo 1 bar baz 123",
				output: []*stringMatch{
					{
						subString: " ",
						span: span{
							from: 3,
							to:   3,
						},
						expressions: newDict(
							"[^a-z]",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "1",
						span: span{
							from: 4,
							to:   4,
						},
						expressions: newDict(
							"[^a-z]",
							"[^\\s]+",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: " ",
						span: span{
							from: 5,
							to:   5,
						},
						expressions: newDict(
							"[^a-z]",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: " ",
						span: span{
							from: 9,
							to:   9,
						},
						expressions: newDict(
							"[^a-z]",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: " ",
						span: span{
							from: 13,
							to:   13,
						},
						expressions: newDict(
							"[^a-z]",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "1",
						span: span{
							from: 14,
							to:   14,
						},
						expressions: newDict(
							"[^a-z]",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "2",
						span: span{
							from: 15,
							to:   15,
						},
						expressions: newDict(
							"[^a-z]",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "3",
						span: span{
							from: 16,
							to:   16,
						},
						expressions: newDict(
							"[^a-z]",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "foo",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"[^\\s]+",
							"[^\\s][^\\s][^\\s]",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "bar",
						span: span{
							from: 6,
							to:   8,
						},
						expressions: newDict(
							"[^\\s]+",
							"[^\\s][^\\s][^\\s]",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "baz",
						span: span{
							from: 10,
							to:   12,
						},
						expressions: newDict(
							"[^\\s]+",
							"ba[^for]",
							"[^\\s][^\\s][^\\s]",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "123",
						span: span{
							from: 14,
							to:   16,
						},
						expressions: newDict(
							"[^\\s]+",
							"[^\\s][^\\s][^\\s]",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
		},
		"real world examples": {
			{
				name: "numeric ranges 000..255",
				regexps: []string{
					"([01][0-9][0-9]|2[0-4][0-9]|25[0-5])",
				},
				input: "000 111 255 256 00 25x 1 2 5",
				output: []*stringMatch{
					{
						subString: "000",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"([01][0-9][0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 0, to: 2},
						},
					},
					{
						subString: "111",
						span: span{
							from: 4,
							to:   6,
						},
						expressions: newDict(
							"([01][0-9][0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 4, to: 6},
						},
					},
					{
						subString: "255",
						span: span{
							from: 8,
							to:   10,
						},
						expressions: newDict(
							"([01][0-9][0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 8, to: 10},
						},
					},
				},
			},
			{
				name: "numeric ranges 0 or 000..255",
				regexps: []string{
					"([01]?[0-9]?[0-9]|2[0-4][0-9]|25[0-5])",
				},
				input: "000 111 255 256 0 12 025",
				output: []*stringMatch{
					{
						subString: "000",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"([01]?[0-9]?[0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 0, to: 2},
						},
					},
					{
						subString: "111",
						span: span{
							from: 4,
							to:   6,
						},
						expressions: newDict(
							"([01]?[0-9]?[0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 4, to: 6},
						},
					},
					{
						subString: "255",
						span: span{
							from: 8,
							to:   10,
						},
						expressions: newDict(
							"([01]?[0-9]?[0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 8, to: 10},
						},
					},
					{
						subString: "25",
						span: span{
							from: 12,
							to:   13,
						},
						expressions: newDict(
							"([01]?[0-9]?[0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 12, to: 13},
						},
					},
					{
						subString: "6",
						span: span{
							from: 14,
							to:   14,
						},
						expressions: newDict(
							"([01]?[0-9]?[0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 14, to: 14},
						},
					},
					{
						subString: "0",
						span: span{
							from: 16,
							to:   16,
						},
						expressions: newDict(
							"([01]?[0-9]?[0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 16, to: 16},
						},
					},
					{
						subString: "12",
						span: span{
							from: 18,
							to:   19,
						},
						expressions: newDict(
							"([01]?[0-9]?[0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 18, to: 19},
						},
					},
					{
						subString: "025",
						span: span{
							from: 21,
							to:   23,
						},
						expressions: newDict(
							"([01]?[0-9]?[0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 21, to: 23},
						},
					},
				},
			},
			{
				name: "numeric ranges 000..127",
				regexps: []string{
					"(0?[0-9]?[0-9]|1[01][0-9]|12[0-7])",
				},
				input: "000 111 127 128 0 12 025",
				output: []*stringMatch{
					{
						subString: "000",
						span: span{
							from: 0,
							to:   2,
						},
						expressions: newDict(
							"(0?[0-9]?[0-9]|1[01][0-9]|12[0-7])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 0, to: 2},
						},
					},
					{
						subString: "111",
						span: span{
							from: 4,
							to:   6,
						},
						expressions: newDict(
							"(0?[0-9]?[0-9]|1[01][0-9]|12[0-7])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 4, to: 6},
						},
					},
					{
						subString: "127",
						span: span{
							from: 8,
							to:   10,
						},
						expressions: newDict(
							"(0?[0-9]?[0-9]|1[01][0-9]|12[0-7])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 8, to: 10},
						},
					},
					{
						subString: "12",
						span: span{
							from: 12,
							to:   13,
						},
						expressions: newDict(
							"(0?[0-9]?[0-9]|1[01][0-9]|12[0-7])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 12, to: 13},
						},
					},
					{
						subString: "8",
						span: span{
							from: 14,
							to:   14,
						},
						expressions: newDict(
							"(0?[0-9]?[0-9]|1[01][0-9]|12[0-7])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 14, to: 14},
						},
					},
					{
						subString: "0",
						span: span{
							from: 16,
							to:   16,
						},
						expressions: newDict(
							"(0?[0-9]?[0-9]|1[01][0-9]|12[0-7])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 16, to: 16},
						},
					},
					{
						subString: "12",
						span: span{
							from: 18,
							to:   19,
						},
						expressions: newDict(
							"(0?[0-9]?[0-9]|1[01][0-9]|12[0-7])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 18, to: 19},
						},
					},
					{
						subString: "025",
						span: span{
							from: 21,
							to:   23,
						},
						expressions: newDict(
							"(0?[0-9]?[0-9]|1[01][0-9]|12[0-7])",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 21, to: 23},
						},
					},
				},
			},
			{
				name: "floating point numbers",
				regexps: []string{
					`[-+]?[0-9]+\.?[0-9]+`,
					`[-+]?[0-9]+.?[0-9]+`,
				},
				input: "+3.14 9.8 2.718 -1.1 +100.500",
				output: []*stringMatch{
					{
						subString: "+3.14",
						span: span{
							from: 0,
							to:   4,
						},
						expressions: newDict(
							`[-+]?[0-9]+\.?[0-9]+`,
							`[-+]?[0-9]+.?[0-9]+`,
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "9.8",
						span: span{
							from: 6,
							to:   8,
						},
						expressions: newDict(
							`[-+]?[0-9]+\.?[0-9]+`,
							`[-+]?[0-9]+.?[0-9]+`,
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "2.718",
						span: span{
							from: 10,
							to:   14,
						},
						expressions: newDict(
							`[-+]?[0-9]+\.?[0-9]+`,
							`[-+]?[0-9]+.?[0-9]+`,
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "-1.1",
						span: span{
							from: 16,
							to:   19,
						},
						expressions: newDict(
							`[-+]?[0-9]+\.?[0-9]+`,
							`[-+]?[0-9]+.?[0-9]+`,
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "+100.500",
						span: span{
							from: 21,
							to:   28,
						},
						expressions: newDict(
							`[-+]?[0-9]+\.?[0-9]+`,
							`[-+]?[0-9]+.?[0-9]+`,
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
			{
				name: "email",
				regexps: []string{
					"[a-z0-9._%+-]+@[a-z0-9.-]+\\.[a-z]{2,}",
					"([a-z0-9._%+-]+)@([a-z0-9.-]+\\.[a-z]{2,})",
					"(?:[a-z0-9._%+-]+)@(?:[a-z0-9.-]+\\.[a-z]{2,})",
					"(?<name>[a-z0-9._%+-]+)@(?<domain>[a-z0-9.-]+\\.[a-z]{2,})",
				},
				input: "123 asd c test@mail.ru asd da a.b@x.y.ru",
				output: []*stringMatch{
					{
						subString: "test@mail.ru",
						span: span{
							from: 10,
							to:   21,
						},
						expressions: newDict(
							"[a-z0-9._%+-]+@[a-z0-9.-]+\\.[a-z]{2,}",
							"(?:[a-z0-9._%+-]+)@(?:[a-z0-9.-]+\\.[a-z]{2,})",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "a.b@x.y.ru",
						span: span{
							from: 30,
							to:   39,
						},
						expressions: newDict(
							"[a-z0-9._%+-]+@[a-z0-9.-]+\\.[a-z]{2,}",
							"(?:[a-z0-9._%+-]+)@(?:[a-z0-9.-]+\\.[a-z]{2,})",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "test@mail.ru",
						span: span{
							from: 10,
							to:   21,
						},
						expressions: newDict(
							"([a-z0-9._%+-]+)@([a-z0-9.-]+\\.[a-z]{2,})",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 10, to: 13},
							{from: 15, to: 21},
						},
					},
					{
						subString: "a.b@x.y.ru",
						span: span{
							from: 30,
							to:   39,
						},
						expressions: newDict(
							"([a-z0-9._%+-]+)@([a-z0-9.-]+\\.[a-z]{2,})",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 30, to: 32},
							{from: 34, to: 39},
						},
					},
					{
						subString: "test@mail.ru",
						span: span{
							from: 10,
							to:   21,
						},
						expressions: newDict(
							"(?<name>[a-z0-9._%+-]+)@(?<domain>[a-z0-9.-]+\\.[a-z]{2,})",
						),
						namedGroups: map[string]span{
							"name": {
								from: 10,
								to:   13,
							},
							"domain": {
								from: 15,
								to:   21,
							},
						},
						groups: []span{},
					},
					{
						subString: "a.b@x.y.ru",
						span: span{
							from: 30,
							to:   39,
						},
						expressions: newDict(
							"(?<name>[a-z0-9._%+-]+)@(?<domain>[a-z0-9.-]+\\.[a-z]{2,})",
						),
						namedGroups: map[string]span{
							"name": {
								from: 30,
								to:   32,
							},
							"domain": {
								from: 34,
								to:   39,
							},
						},
						groups: []span{},
					},
				},
			},
			{
				name: "card numbers",
				regexps: []string{
					"4[0-9]{12}(?:[0-9]{3})?",
					"(?:5[1-5][0-9]{2}|222[1-9]|22[3-9][0-9]|2[3-6][0-9]{2}|27[01][0-9]|2720)[0-9]{12}",
				},
				input: "4111111111111111 5105105105105100 4012888888881881 4222222222222 5555555555554444",
				output: []*stringMatch{
					{
						subString: "4111111111111111",
						span: span{
							from: 0,
							to:   15,
						},
						expressions: newDict(
							"4[0-9]{12}(?:[0-9]{3})?",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "4012888888881881",
						span: span{
							from: 34,
							to:   49,
						},
						expressions: newDict(
							"4[0-9]{12}(?:[0-9]{3})?",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "4222222222222",
						span: span{
							from: 51,
							to:   63,
						},
						expressions: newDict(
							"4[0-9]{12}(?:[0-9]{3})?",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "5105105105105100",
						span: span{
							from: 17,
							to:   32,
						},
						expressions: newDict(
							"(?:5[1-5][0-9]{2}|222[1-9]|22[3-9][0-9]|2[3-6][0-9]{2}|27[01][0-9]|2720)[0-9]{12}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "5555555555554444",
						span: span{
							from: 65,
							to:   80,
						},
						expressions: newDict(
							"(?:5[1-5][0-9]{2}|222[1-9]|22[3-9][0-9]|2[3-6][0-9]{2}|27[01][0-9]|2720)[0-9]{12}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
				},
			},
			{
				name: "complete line",
				regexps: []string{
					"^.*$",
					"^(.*)$",
					"^.{2}",
					".{2}$",
				},
				input: "Lorem Ipsum is simply dummy text of the printing and typesetting industry.",
				output: []*stringMatch{
					{
						subString: "Lorem Ipsum is simply dummy text of the printing and typesetting industry.",
						span: span{
							from: 0,
							to:   73,
						},
						expressions: newDict(
							"^.*$",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "Lo",
						span: span{
							from: 0,
							to:   1,
						},
						expressions: newDict(
							"^.{2}",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "y.",
						span: span{
							from: 72,
							to:   73,
						},
						expressions: newDict(
							".{2}$",
						),
						namedGroups: map[string]span{},
						groups:      []span{},
					},
					{
						subString: "Lorem Ipsum is simply dummy text of the printing and typesetting industry.",
						span: span{
							from: 0,
							to:   73,
						},
						expressions: newDict(
							"^(.*)$",
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 0, to: 73},
						},
					},
				},
			},

			{
				name: "HTML",
				regexps: []string{
					`<p>(.*)</p>`,
				},
				input: "Lorem Ipsum is <p>simply dummy text</p> of the printing and typesetting industry.",
				output: []*stringMatch{
					{
						subString: "<p>simply dummy text</p>",
						span: span{
							from: 15,
							to:   38,
						},
						expressions: newDict(
							`<p>(.*)</p>`,
						),
						namedGroups: map[string]span{},
						groups: []span{
							{from: 18, to: 34},
						},
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

					// t.Log(tr.String())
					// t.Logf("input: '%s'", string(test.input))

					actual := tr.Match(test.input)
					require.NoError(t, err)

					sort.SliceStable(actual, func(i, j int) bool {
						return actual[i].Key() < actual[j].Key()
					})

					sort.SliceStable(test.output, func(i, j int) bool {
						return test.output[i].Key() < test.output[j].Key()
					})

					if len(test.output) != len(actual) {
						expectedStrings := make([]string, len(test.output))
						actualStrings := make([]string, len(actual))

						for i, x := range test.output {
							expectedStrings[i] = x.String()
						}

						for i, x := range actual {
							actualStrings[i] = x.String()
						}

						// t.Log("expected strings: ", expectedStrings)
						// t.Log("actual strings: ", actualStrings)

						require.Equal(
							t,
							expectedStrings,
							actualStrings,
						)
					}

					for i := range test.output {
						// t.Log("output", *test.output[i])
						// t.Log("actual", *actual[i])

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
	output  []*stringMatch
}

func comparator(x, y *stringMatch) bool {
	return x.Key() < y.Key()
}

func pointer[T any](x T) *T {
	return &x
}

func Test_It(t *testing.T) {
	tr, err := NewTrie(
		"ba[^for]",
	)
	require.NoError(t, err)

	t.Log(tr.String())

	expected := []*stringMatch{
		{
			subString: "baz",
			span: span{
				from: 0,
				to:   2,
			},
			expressions: newDict(
				"ba[^for]",
			),
			namedGroups: map[string]span{},
			groups:      []span{},
		},
	}

	sort.SliceStable(expected, func(i, j int) bool {
		return comparator(expected[i], expected[j])
	})

	actual := tr.Match("baz")
	require.NoError(t, err)

	sort.SliceStable(actual, func(i, j int) bool {
		return comparator(actual[i], actual[j])
	})

	if len(expected) != len(actual) {
		require.Equal(t, expected, actual)
	}

	for i := range expected {
		es := expected[i].expressions.Slice()
		as := actual[i].expressions.Slice()

		sort.SliceStable(es, func(x, y int) bool { return es[x] < as[y] })
		sort.SliceStable(as, func(x, y int) bool { return as[x] < as[y] })

		require.Equalf(t, *expected[i], *actual[i], "compare %d match", i)
	}
}

// func Test_Chain(t *testing.T) {
// 	tr, err := NewTrie(
// 		"...$",
// 		".$",
// 	)
// 	require.NoError(t, err)

// 	t.Log(tr.String())

// 	expected := []*stringMatch{
// 		{
// 			subString: "bar",
// 			from:      4,
// 			to:        7,
// 			expressions: []string{
// 				"...$",
// 			},
// 			namedGroups: map[string]span{},
// 			groups:      []span{},
// 			empty:       false,
// 		},
// 		{
// 			subString: "baz",
// 			from:      8,
// 			to:        10,
// 			expressions: []string{
// 				"...$",
// 			},
// 			namedGroups: map[string]span{},
// 			groups:      []span{},
// 			empty:       false,
// 		},
// 		{
// 			subString: "r",
// 			from:      6,
// 			to:        7,
// 			expressions: []string{
// 				".$",
// 			},
// 			namedGroups: map[string]span{},
//  			groups:      []span{},
// 			empty: false,
// 		},
// 		{
// 			subString: "z",
// 			from:      10,
// 			to:        10,
// 			expressions: []string{
// 				".$",
// 			},
// 			namedGroups: map[string]span{},
// 			groups:      []span{},
// 			empty: false,
// 		},
// 	}

// 	sort.SliceStable(expected, func(i, j int) bool {
// 		return comparator(expected[i], expected[j])
// 	})

// 	actual := tr.Match("foo bar\nbaz")
// 	require.NoError(t, err)

// 	sort.SliceStable(actual, func(i, j int) bool {
// 		return comparator(actual[i], actual[j])
// 	})

// 	if len(expected) != len(actual) {
// 		require.Equal(t, expected, actual)
// 	}

// 	for i := range expected {
// 		sort.SliceStable(expected[i].expressions, func(x, y int) bool {
// 			return expected[i].expressions[x] < expected[i].expressions[y]
// 		})

// 		sort.SliceStable(actual[i].expressions, func(x, y int) bool {
// 			return actual[i].expressions[x] < actual[i].expressions[y]
// 		})

// 		require.Equalf(t, *expected[i], *actual[i], "compare %d match", i)
// 	}
// }
