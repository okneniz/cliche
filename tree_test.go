package cliche

import (
	"sort"
	"testing"

	"github.com/okneniz/cliche/span"

	"github.com/stretchr/testify/require"
)

// https://www.rfc-editor.org/rfc/rfc9485.html#name-implementing-i-regexp

func TestTree(t *testing.T) {
	t.Parallel()

	tr, err := New(
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

type TestFile struct {
	Name  string
	Tests []Test
}

type Test struct {
	Name        string
	Expressions []string
	Input       string
	Want        []*Match
}

type Expectation struct {
	SubString   string
	Span        span.Interface
	Expressions Set
	Groups      []span.Interface
	NamedGroups map[string]span.Interface
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
				output: []*Match{
					{
						subString: "te",
						span:      span.New(0, 1),
						expressions: newSet(
							"te",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "test",
						span: span.New(
							0,
							3,
						),
						expressions: newSet(
							"test",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "s",
						span: span.New(
							2,
							2,
						),
						expressions: newSet(
							"s",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "ing",
						span: span.New(
							4,
							6,
						),
						expressions: newSet(
							"ing",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "s",
						span: span.New(
							8,
							8,
						),
						expressions: newSet(
							"s",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "string",
						span: span.New(
							8,
							13,
						),
						expressions: newSet(
							"string",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "ing",
						span: span.New(
							11,
							13,
						),
						expressions: newSet(
							"ing",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "te",
						span: span.New(
							15,
							16,
						),
						expressions: newSet(
							"te",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "test",
						span: span.New(
							15,
							18,
						),
						expressions: newSet(
							"test",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "s",
						span: span.New(
							17,
							17,
						),
						expressions: newSet(
							"s",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "s",
						span: span.New(
							20,
							20,
						),
						expressions: newSet(
							"s",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "s",
						span: span.New(
							21,
							21,
						),
						expressions: newSet(
							"s",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "s",
						span: span.New(
							22,
							22,
						),
						expressions: newSet(
							"s",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "s",
						span: span.New(
							23,
							23,
						),
						expressions: newSet(
							"s",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "word",
						span: span.New(
							25,
							28,
						),
						expressions: newSet(
							"word",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "word",
						span: span.New(
							30,
							33,
						),
						expressions: newSet(
							"word",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "s",
						span: span.New(
							34,
							34,
						),
						expressions: newSet(
							"s",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
				},
			},
			{
				name:    "match '.'",
				regexps: []string{"t."},
				input:   "testing string test ssss word words",
				output: []*Match{
					{
						subString: "te",
						span: span.New(
							0,
							1,
						),
						expressions: newSet(
							"t.",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "ti",
						span: span.New(
							3,
							4,
						),
						expressions: newSet(
							"t.",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "tr",
						span: span.New(
							9,
							10,
						),
						expressions: newSet(
							"t.",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "te",
						span: span.New(
							15,
							16,
						),
						expressions: newSet(
							"t.",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "t ",
						span: span.New(
							18,
							19,
						),
						expressions: newSet(
							"t.",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
				},
			},
			{
				name:    "match multiple '.'",
				regexps: []string{"t.."},
				input:   "testing string test ssss word words",
				output: []*Match{
					{
						subString: "tes",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"t..",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "tin",
						span: span.New(
							3,
							5,
						),
						expressions: newSet(
							"t..",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "tri",
						span: span.New(
							9,
							11,
						),
						expressions: newSet(
							"t..",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "tes",
						span: span.New(
							15,
							17,
						),
						expressions: newSet(
							"t..",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "t s",
						span: span.New(
							18,
							20,
						),
						expressions: newSet(
							"t..",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
				},
			},
			{
				name:    "match '\\d'",
				regexps: []string{"\\d"},
				input:   "asd 1 jsdfk 4234",
				output: []*Match{
					{
						subString: "1",
						span: span.New(
							4,
							4,
						),
						expressions: newSet(
							"\\d",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "4",
						span: span.New(
							12,
							12,
						),
						expressions: newSet(
							"\\d",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "2",
						span: span.New(
							13,
							13,
						),
						expressions: newSet(
							"\\d",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "3",
						span: span.New(
							14,
							14,
						),
						expressions: newSet(
							"\\d",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "4",
						span: span.New(
							15,
							15,
						),
						expressions: newSet(
							"\\d",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
				},
			},
			{
				name:    "match '\\D'",
				regexps: []string{"\\D"},
				input:   "asd 1 jsdfk 4234",
				output: []*Match{
					{
						subString: "a",
						span: span.New(
							0,
							0,
						),
						expressions: newSet(
							"\\D",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "s",
						span: span.New(
							1,
							1,
						),
						expressions: newSet(
							"\\D",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "d",
						span: span.New(
							2,
							2,
						),
						expressions: newSet(
							"\\D",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: " ",
						span: span.New(
							3,
							3,
						),
						expressions: newSet(
							"\\D",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: " ",
						span: span.New(
							5,
							5,
						),
						expressions: newSet(
							"\\D",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "j",
						span: span.New(
							6,
							6,
						),
						expressions: newSet(
							"\\D",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "s",
						span: span.New(
							7,
							7,
						),
						expressions: newSet(
							"\\D",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "d",
						span: span.New(
							8,
							8,
						),
						expressions: newSet(
							"\\D",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "f",
						span: span.New(
							9,
							9,
						),
						expressions: newSet(
							"\\D",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "k",
						span: span.New(
							10,
							10,
						),
						expressions: newSet(
							"\\D",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: " ",
						span: span.New(
							11,
							11,
						),
						expressions: newSet(
							"\\D",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
				},
			},
			{
				name:    "match '\\w'",
				regexps: []string{"\\w"},
				input:   "asd 1 jsdfk 4234",
				output: []*Match{
					{
						subString: "a",
						span: span.New(
							0,
							0,
						),
						expressions: newSet(
							"\\w",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "s",
						span: span.New(
							1,
							1,
						),
						expressions: newSet(
							"\\w",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "d",
						span: span.New(
							2,
							2,
						),
						expressions: newSet(
							"\\w",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "1",
						span: span.New(
							4,
							4,
						),
						expressions: newSet(
							"\\w",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "j",
						span: span.New(
							6,
							6,
						),
						expressions: newSet(
							"\\w",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "s",
						span: span.New(
							7,
							7,
						),
						expressions: newSet(
							"\\w",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "d",
						span: span.New(
							8,
							8,
						),
						expressions: newSet(
							"\\w",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "f",
						span: span.New(
							9,
							9,
						),
						expressions: newSet(
							"\\w",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "k",
						span: span.New(
							10,
							10,
						),
						expressions: newSet(
							"\\w",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "4",
						span: span.New(
							12,
							12,
						),
						expressions: newSet(
							"\\w",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "2",
						span: span.New(
							13,
							13,
						),
						expressions: newSet(
							"\\w",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "3",
						span: span.New(
							14,
							14,
						),
						expressions: newSet(
							"\\w",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "4",
						span: span.New(
							15,
							15,
						),
						expressions: newSet(
							"\\w",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
				},
			},
			{
				name:    "match '\\W'",
				regexps: []string{"\\W"},
				input:   "asd 1 jsdfk 4234!\n\r",
				output: []*Match{
					{
						subString: " ",
						span: span.New(
							3,
							3,
						),
						expressions: newSet(
							"\\W",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: " ",
						span: span.New(
							5,
							5,
						),
						expressions: newSet(
							"\\W",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: " ",
						span: span.New(
							11,
							11,
						),
						expressions: newSet(
							"\\W",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "!",
						span: span.New(
							16,
							16,
						),
						expressions: newSet(
							"\\W",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "\n",
						span: span.New(
							17,
							17,
						),
						expressions: newSet(
							"\\W",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "\r",
						span: span.New(
							18,
							18,
						),
						expressions: newSet(
							"\\W",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
				},
			},
			{
				name:    "match '\\s'",
				regexps: []string{"\\s"},
				input:   "asd 1 jsdfk 4234",
				output: []*Match{
					{
						subString: " ",
						span: span.New(
							3,
							3,
						),
						expressions: newSet(
							"\\s",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: " ",
						span: span.New(
							5,
							5,
						),
						expressions: newSet(
							"\\s",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: " ",
						span: span.New(
							11,
							11,
						),
						expressions: newSet(
							"\\s",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
				},
			},
			{
				name:    "match '\\S'",
				regexps: []string{"\\S"},
				input:   "asd 1 jsdfk 4234!",
				output: []*Match{
					{
						subString: "a",
						span: span.New(
							0,
							0,
						),
						expressions: newSet(
							"\\S",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "s",
						span: span.New(
							1,
							1,
						),
						expressions: newSet(
							"\\S",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "d",
						span: span.New(
							2,
							2,
						),
						expressions: newSet(
							"\\S",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "1",
						span: span.New(
							4,
							4,
						),
						expressions: newSet(
							"\\S",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "j",
						span: span.New(
							6,
							6,
						),
						expressions: newSet(
							"\\S",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "s",
						span: span.New(
							7,
							7,
						),
						expressions: newSet(
							"\\S",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "d",
						span: span.New(
							8,
							8,
						),
						expressions: newSet(
							"\\S",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "f",
						span: span.New(
							9,
							9,
						),
						expressions: newSet(
							"\\S",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "k",
						span: span.New(
							10,
							10,
						),
						expressions: newSet(
							"\\S",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "4",
						span: span.New(
							12,
							12,
						),
						expressions: newSet(
							"\\S",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "2",
						span: span.New(
							13,
							13,
						),
						expressions: newSet(
							"\\S",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "3",
						span: span.New(
							14,
							14,
						),
						expressions: newSet(
							"\\S",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "4",
						span: span.New(
							15,
							15,
						),
						expressions: newSet(
							"\\S",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "!",
						span: span.New(
							16,
							16,
						),
						expressions: newSet(
							"\\S",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
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
				output: []*Match{
					{
						subString: ".",
						span: span.New(
							0,
							0,
						),
						expressions: newSet(
							"\\.",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "?",
						span: span.New(
							2,
							2,
						),
						expressions: newSet(
							"\\?",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "+",
						span: span.New(
							4,
							4,
						),
						expressions: newSet(
							"\\+",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "*",
						span: span.New(
							6,
							6,
						),
						expressions: newSet(
							"\\*",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "^",
						span: span.New(
							8,
							8,
						),
						expressions: newSet(
							"\\^",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "$",
						span: span.New(
							10,
							10,
						),
						expressions: newSet(
							"\\$",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "[",
						span: span.New(
							12,
							12,
						),
						expressions: newSet(
							"\\[",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "]",
						span: span.New(
							14,
							14,
						),
						expressions: newSet(
							"\\]",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "{",
						span: span.New(
							16,
							16,
						),
						expressions: newSet(
							"\\{",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "}",
						span: span.New(
							18,
							18,
						),
						expressions: newSet(
							"\\}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
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
				output: []*Match{
					{
						subString: "foo",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"fo(o|b)",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(2, 2),
						},
					},
					{
						subString: "foo",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"f(o|b)o",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(1, 1),
						},
					},
					{
						subString: "foo",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"(f|b)(o|a)(o|r|z)",
							"(f|b)(o|a)(o|\\w|\\D)",
							"(f)(o)(o)",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(0, 0),
							span.New(1, 1),
							span.New(2, 2),
						},
					},
					{
						subString: "bar",
						span: span.New(
							4,
							6,
						),
						expressions: newSet(
							"(f|b)(o|a)(o|r|z)",
							"(f|b)(o|a)(o|\\w|\\D)",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(4, 4),
							span.New(5, 5),
							span.New(6, 6),
						},
					},
					{
						subString: "baz",
						span: span.New(
							8,
							10,
						),
						expressions: newSet(
							"(f|b)(o|a)(o|r|z)",
							"(f|b)(o|a)(o|\\w|\\D)",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(8, 8),
							span.New(9, 9),
							span.New(10, 10),
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
				output: []*Match{
					{
						subString: "foo",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"f(o(o))",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(1, 2),
							span.New(2, 2),
						},
					},
					{
						subString: "bar",
						span: span.New(
							4,
							6,
						),
						expressions: newSet(
							"(b(a(r)))",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(4, 6),
							span.New(5, 6),
							span.New(6, 6),
						},
					},
					{
						subString: "baz",
						span: span.New(
							8,
							10,
						),
						expressions: newSet(
							"((b)az)",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(8, 10),
							span.New(8, 8),
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
				output: []*Match{
					{
						subString: "foo",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"fo(?<name>o|b)",
						),
						namedGroups: map[string]span.Interface{
							"name": span.New(2, 2),
						},
						groups: []span.Interface{},
					},
					{
						subString: "foo",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"f(?<name>o|b)o",
						),
						namedGroups: map[string]span.Interface{
							"name": span.New(1, 1),
						},
						groups: []span.Interface{},
					},
					{
						subString: "foo",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"(?<first>f|b)(?<second>o|a)(?<third>o|r|z)",
							"(?<first>f|b)(?<second>o|a)(?<third>o|\\w|\\D)",
							"(?<first>f)(?<second>o)(?<third>o)",
						),
						namedGroups: map[string]span.Interface{
							"first":  span.New(0, 0),
							"second": span.New(1, 1),
							"third":  span.New(2, 2),
						},
						groups: []span.Interface{},
					},
					{
						subString: "bar",
						span: span.New(
							4,
							6,
						),
						expressions: newSet(
							"(?<first>f|b)(?<second>o|a)(?<third>o|r|z)",
							"(?<first>f|b)(?<second>o|a)(?<third>o|\\w|\\D)",
						),
						namedGroups: map[string]span.Interface{
							"first":  span.New(4, 4),
							"second": span.New(5, 5),
							"third":  span.New(6, 6),
						},
						groups: []span.Interface{},
					},
					{
						subString: "baz",
						span: span.New(
							8,
							10,
						),
						expressions: newSet(
							"(?<first>f|b)(?<second>o|a)(?<third>o|r|z)",
							"(?<first>f|b)(?<second>o|a)(?<third>o|\\w|\\D)",
						),
						namedGroups: map[string]span.Interface{
							"first":  span.New(8, 8),
							"second": span.New(9, 9),
							"third":  span.New(10, 10),
						},
						groups: []span.Interface{},
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
				output: []*Match{
					{
						subString: "foo",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"f(?<first>o(?<second>o))",
						),
						namedGroups: map[string]span.Interface{
							"first":  span.New(1, 2),
							"second": span.New(2, 2),
						},
						groups: []span.Interface{},
					},
					{
						subString: "bar",
						span: span.New(
							4,
							6,
						),
						expressions: newSet(
							"(?<first>b(?<second>a(?<third>r)))",
						),
						namedGroups: map[string]span.Interface{
							"first":  span.New(4, 6),
							"second": span.New(5, 6),
							"third":  span.New(6, 6),
						},
						groups: []span.Interface{},
					},
					{
						subString: "baz",
						span: span.New(
							8,
							10,
						),
						expressions: newSet(
							"(?<first>(?<second>b)az)",
						),
						namedGroups: map[string]span.Interface{
							"second": span.New(8, 8),
							"first":  span.New(8, 10),
						},
						groups: []span.Interface{},
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
				output: []*Match{
					{
						subString: "foo",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"fo(?:o|b)",
							"f(?:o|b)o",
							"(?:f|b)(?:o|a)(?:o|r|z)",
							"(?:f|b)(?:o|a)(?:o|\\w|\\D)",
							"(?:f)(?:o)(?:o)",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "bar",
						span: span.New(
							4,
							6,
						),
						expressions: newSet(
							"(?:f|b)(?:o|a)(?:o|r|z)",
							"(?:f|b)(?:o|a)(?:o|\\w|\\D)",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "baz",
						span: span.New(
							8,
							10,
						),
						expressions: newSet(
							"(?:f|b)(?:o|a)(?:o|r|z)",
							"(?:f|b)(?:o|a)(?:o|\\w|\\D)",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
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
				output: []*Match{
					{
						subString: "foo",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"f(?:o(?:o))",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "bar",
						span: span.New(
							4,
							6,
						),
						expressions: newSet(
							"(?:b(?:a(?:r)))",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "baz",
						span: span.New(
							8,
							10,
						),
						expressions: newSet(
							"(?:(?:b)az)",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
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
				output: []*Match{
					{
						subString: "",
						span:      span.Empty(0),
						expressions: newSet(
							"c?",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "",
						span:      span.Empty(1),
						expressions: newSet(
							"c?",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "c",
						span: span.New(
							2,
							2,
						),
						expressions: newSet(
							"c?",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					// TODO : check c? in rubular again (it's want end of string too)
					{
						subString: "pic",
						span:      span.New(0, 2),
						expressions: newSet(
							"pics?",
							"pi.?c",
							"....?",
							"...?.",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
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
				output: []*Match{
					{
						subString: "xx",
						span: span.New(
							0,
							1,
						),
						expressions: newSet(
							"x*",
							"x*x",
							"x{0,}",
							"x{0,}x",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "",
						span:      span.Empty(2),
						expressions: newSet(
							"x*",
							"x{0,}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "x",
						span: span.New(
							3,
							3,
						),
						expressions: newSet(
							"x*",
							"x*x",
							"x{0,}",
							"x{0,}x",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "xx x",
						span: span.New(
							0,
							3,
						),
						expressions: newSet(
							"x.*",
							"x.{0,}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "",
						span:      span.Empty(4),
						expressions: newSet(
							"x*",
							"x{0,}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "",
						span:      span.Empty(5),
						expressions: newSet(
							"x*",
							"x{0,}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "x",
						span: span.New(
							6,
							6,
						),
						expressions: newSet(
							"x*",
							"x*x",
							"x.*",
							"x{0,}",
							"x{0,}x",
							"x.{0,}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
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
				output: []*Match{
					{
						subString: "xx",
						span: span.New(
							0,
							1,
						),
						expressions: newSet(
							"x+",
							"x{1,}",
							"x+x",
							"x{1,}x",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "x",
						span: span.New(
							3,
							3,
						),
						expressions: newSet(
							"x+",
							"x{1,}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "x",
						span: span.New(
							6,
							6,
						),
						expressions: newSet(
							"x+",
							"x{1,}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "xx x",
						span: span.New(
							0,
							3,
						),
						expressions: newSet(
							"x.+",
							"x.{1,}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
				},
			},
			{
				name: "endless quantifier",
				regexps: []string{
					"x{2,}",
				},
				input: "xx xxx x",
				output: []*Match{
					{
						subString: "xx",
						span: span.New(
							0,
							1,
						),
						expressions: newSet(
							"x{2,}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "xxx",
						span: span.New(
							3,
							5,
						),
						expressions: newSet(
							"x{2,}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
				},
			},
			{
				name: "limited quantifier",
				regexps: []string{
					"x{2,4}",
				},
				input: "xx xxx x xxxxxx",
				output: []*Match{
					{
						subString: "xx",
						span: span.New(
							0,
							1,
						),
						expressions: newSet(
							"x{2,4}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "xxx",
						span: span.New(
							3,
							5,
						),
						expressions: newSet(
							"x{2,4}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "xxxx",
						span: span.New(
							9,
							12,
						),
						expressions: newSet(
							"x{2,4}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "xx",
						span: span.New(
							13,
							14,
						),
						expressions: newSet(
							"x{2,4}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
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
				output: []*Match{
					{
						subString: "foo",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"^...",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "baz",
						span: span.New(
							8,
							10,
						),
						expressions: newSet(
							"^...",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "f",
						span: span.New(
							0,
							0,
						),
						expressions: newSet(
							"^.",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "b",
						span: span.New(
							8,
							8,
						),
						expressions: newSet(
							"^.",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
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
				output: []*Match{
					{
						subString: "foo",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"\\A...",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "f",
						span: span.New(
							0,
							0,
						),
						expressions: newSet(
							"\\A.",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "",
						span:      span.Empty(0),
						expressions: newSet(
							"\\A",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
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
				output: []*Match{
					{
						subString: "bar",
						span: span.New(
							4,
							6,
						),
						expressions: newSet(
							"...$",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "baz",
						span: span.New(
							8,
							10,
						),
						expressions: newSet(
							"...$",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "r",
						span: span.New(
							6,
							6,
						),
						expressions: newSet(
							".$",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "",
						span:      span.Empty(7),
						expressions: newSet(
							"$",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "z",
						span: span.New(
							10,
							10,
						),
						expressions: newSet(
							".$",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					// TODO: match or not?
					// {
					// 	subString: "",
					// 	span: span.New(
					// 		from:  11,
					// 		to:    11,
					// 		empty: true,
					// ),
					// 	expressions: newSet(
					// 		"$",
					// 	),
					// 	namedGroups: map[string]span.Interface{},
					// 	groups:      []span.Interface{},
					// },
				},
			},
			{
				name: "string",
				regexps: []string{
					"...\\z",
					".\\z",
					// "\\z", should be matched?
				},
				input: "foo bar\nbaz",
				output: []*Match{
					{
						subString: "z",
						span: span.New(
							10,
							10,
						),
						expressions: newSet(
							".\\z",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "baz",
						span: span.New(
							8,
							10,
						),
						expressions: newSet(
							"...\\z",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					// {
					// 	subString: "",
					// 	span:      span.Empty(11),
					// 	expressions: newSet(
					// 		"\\z",
					// 	),
					// 	namedGroups: map[string]span.Interface{},
					// 	groups:      []span.Interface{},
					// },
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
				output: []*Match{
					{
						subString: "1",
						span: span.New(
							4,
							4,
						),
						expressions: newSet(
							"[0-9]",
							"[0-9]+",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "1",
						span: span.New(
							14,
							14,
						),
						expressions: newSet(
							"[0-9]",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "2",
						span: span.New(
							15,
							15,
						),
						expressions: newSet(
							"[0-9]",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "3",
						span: span.New(
							16,
							16,
						),
						expressions: newSet(
							"[0-9]",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "123",
						span: span.New(
							14,
							16,
						),
						expressions: newSet(
							"[0-9]+",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "bar",
						span: span.New(
							6,
							8,
						),
						expressions: newSet(
							"ba[rz]",
							"[faborz]+",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "baz",
						span: span.New(
							10,
							12,
						),
						expressions: newSet(
							"ba[rz]",
							"[faborz]+",
							"[bar][bar][baz]",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "foo",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"[faborz]+",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
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
				output: []*Match{
					{
						subString: " ",
						span: span.New(
							3,
							3,
						),
						expressions: newSet(
							"[^a-z]",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "1",
						span: span.New(
							4,
							4,
						),
						expressions: newSet(
							"[^a-z]",
							"[^\\s]+",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: " ",
						span: span.New(
							5,
							5,
						),
						expressions: newSet(
							"[^a-z]",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: " ",
						span: span.New(
							9,
							9,
						),
						expressions: newSet(
							"[^a-z]",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: " ",
						span: span.New(
							13,
							13,
						),
						expressions: newSet(
							"[^a-z]",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "1",
						span: span.New(
							14,
							14,
						),
						expressions: newSet(
							"[^a-z]",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "2",
						span: span.New(
							15,
							15,
						),
						expressions: newSet(
							"[^a-z]",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "3",
						span: span.New(
							16,
							16,
						),
						expressions: newSet(
							"[^a-z]",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "foo",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"[^\\s]+",
							"[^\\s][^\\s][^\\s]",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "bar",
						span: span.New(
							6,
							8,
						),
						expressions: newSet(
							"[^\\s]+",
							"[^\\s][^\\s][^\\s]",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "baz",
						span: span.New(
							10,
							12,
						),
						expressions: newSet(
							"[^\\s]+",
							"ba[^for]",
							"[^\\s][^\\s][^\\s]",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "123",
						span: span.New(
							14,
							16,
						),
						expressions: newSet(
							"[^\\s]+",
							"[^\\s][^\\s][^\\s]",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
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
				output: []*Match{
					{
						subString: "000",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"([01][0-9][0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(0, 2),
						},
					},
					{
						subString: "111",
						span: span.New(
							4,
							6,
						),
						expressions: newSet(
							"([01][0-9][0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(4, 6),
						},
					},
					{
						subString: "255",
						span: span.New(
							8,
							10,
						),
						expressions: newSet(
							"([01][0-9][0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(8, 10),
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
				output: []*Match{
					{
						subString: "000",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"([01]?[0-9]?[0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(0, 2),
						},
					},
					{
						subString: "111",
						span: span.New(
							4,
							6,
						),
						expressions: newSet(
							"([01]?[0-9]?[0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(4, 6),
						},
					},
					{
						subString: "255",
						span: span.New(
							8,
							10,
						),
						expressions: newSet(
							"([01]?[0-9]?[0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(8, 10),
						},
					},
					{
						subString: "25",
						span: span.New(
							12,
							13,
						),
						expressions: newSet(
							"([01]?[0-9]?[0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(12, 13),
						},
					},
					{
						subString: "6",
						span: span.New(
							14,
							14,
						),
						expressions: newSet(
							"([01]?[0-9]?[0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(14, 14),
						},
					},
					{
						subString: "0",
						span: span.New(
							16,
							16,
						),
						expressions: newSet(
							"([01]?[0-9]?[0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(16, 16),
						},
					},
					{
						subString: "12",
						span: span.New(
							18,
							19,
						),
						expressions: newSet(
							"([01]?[0-9]?[0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(18, 19),
						},
					},
					{
						subString: "025",
						span: span.New(
							21,
							23,
						),
						expressions: newSet(
							"([01]?[0-9]?[0-9]|2[0-4][0-9]|25[0-5])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(21, 23),
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
				output: []*Match{
					{
						subString: "000",
						span: span.New(
							0,
							2,
						),
						expressions: newSet(
							"(0?[0-9]?[0-9]|1[01][0-9]|12[0-7])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(0, 2),
						},
					},
					{
						subString: "111",
						span: span.New(
							4,
							6,
						),
						expressions: newSet(
							"(0?[0-9]?[0-9]|1[01][0-9]|12[0-7])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(4, 6),
						},
					},
					{
						subString: "127",
						span: span.New(
							8,
							10,
						),
						expressions: newSet(
							"(0?[0-9]?[0-9]|1[01][0-9]|12[0-7])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(8, 10),
						},
					},
					{
						subString: "12",
						span: span.New(
							12,
							13,
						),
						expressions: newSet(
							"(0?[0-9]?[0-9]|1[01][0-9]|12[0-7])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(12, 13),
						},
					},
					{
						subString: "8",
						span: span.New(
							14,
							14,
						),
						expressions: newSet(
							"(0?[0-9]?[0-9]|1[01][0-9]|12[0-7])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(14, 14),
						},
					},
					{
						subString: "0",
						span: span.New(
							16,
							16,
						),
						expressions: newSet(
							"(0?[0-9]?[0-9]|1[01][0-9]|12[0-7])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(16, 16),
						},
					},
					{
						subString: "12",
						span: span.New(
							18,
							19,
						),
						expressions: newSet(
							"(0?[0-9]?[0-9]|1[01][0-9]|12[0-7])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(18, 19),
						},
					},
					{
						subString: "025",
						span: span.New(
							21,
							23,
						),
						expressions: newSet(
							"(0?[0-9]?[0-9]|1[01][0-9]|12[0-7])",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(21, 23),
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
				output: []*Match{
					{
						subString: "+3.14",
						span: span.New(
							0,
							4,
						),
						expressions: newSet(
							`[-+]?[0-9]+\.?[0-9]+`,
							`[-+]?[0-9]+.?[0-9]+`,
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "9.8",
						span: span.New(
							6,
							8,
						),
						expressions: newSet(
							`[-+]?[0-9]+\.?[0-9]+`,
							`[-+]?[0-9]+.?[0-9]+`,
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "2.718",
						span: span.New(
							10,
							14,
						),
						expressions: newSet(
							`[-+]?[0-9]+\.?[0-9]+`,
							`[-+]?[0-9]+.?[0-9]+`,
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "-1.1",
						span: span.New(
							16,
							19,
						),
						expressions: newSet(
							`[-+]?[0-9]+\.?[0-9]+`,
							`[-+]?[0-9]+.?[0-9]+`,
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "+100.500",
						span: span.New(
							21,
							28,
						),
						expressions: newSet(
							`[-+]?[0-9]+\.?[0-9]+`,
							`[-+]?[0-9]+.?[0-9]+`,
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
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
				output: []*Match{
					{
						subString: "test@mail.ru",
						span: span.New(
							10,
							21,
						),
						expressions: newSet(
							"[a-z0-9._%+-]+@[a-z0-9.-]+\\.[a-z]{2,}",
							"(?:[a-z0-9._%+-]+)@(?:[a-z0-9.-]+\\.[a-z]{2,})",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "a.b@x.y.ru",
						span: span.New(
							30,
							39,
						),
						expressions: newSet(
							"[a-z0-9._%+-]+@[a-z0-9.-]+\\.[a-z]{2,}",
							"(?:[a-z0-9._%+-]+)@(?:[a-z0-9.-]+\\.[a-z]{2,})",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "test@mail.ru",
						span: span.New(
							10,
							21,
						),
						expressions: newSet(
							"([a-z0-9._%+-]+)@([a-z0-9.-]+\\.[a-z]{2,})",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(10, 13),
							span.New(15, 21),
						},
					},
					{
						subString: "a.b@x.y.ru",
						span: span.New(
							30,
							39,
						),
						expressions: newSet(
							"([a-z0-9._%+-]+)@([a-z0-9.-]+\\.[a-z]{2,})",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(30, 32),
							span.New(34, 39),
						},
					},
					{
						subString: "test@mail.ru",
						span: span.New(
							10,
							21,
						),
						expressions: newSet(
							"(?<name>[a-z0-9._%+-]+)@(?<domain>[a-z0-9.-]+\\.[a-z]{2,})",
						),
						namedGroups: map[string]span.Interface{
							"name": span.New(
								10,
								13,
							),
							"domain": span.New(
								15,
								21,
							),
						},
						groups: []span.Interface{},
					},
					{
						subString: "a.b@x.y.ru",
						span: span.New(
							30,
							39,
						),
						expressions: newSet(
							"(?<name>[a-z0-9._%+-]+)@(?<domain>[a-z0-9.-]+\\.[a-z]{2,})",
						),
						namedGroups: map[string]span.Interface{
							"name": span.New(
								30,
								32,
							),
							"domain": span.New(
								34,
								39,
							),
						},
						groups: []span.Interface{},
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
				output: []*Match{
					{
						subString: "4111111111111111",
						span: span.New(
							0,
							15,
						),
						expressions: newSet(
							"4[0-9]{12}(?:[0-9]{3})?",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "4012888888881881",
						span: span.New(
							34,
							49,
						),
						expressions: newSet(
							"4[0-9]{12}(?:[0-9]{3})?",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "4222222222222",
						span: span.New(
							51,
							63,
						),
						expressions: newSet(
							"4[0-9]{12}(?:[0-9]{3})?",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "5105105105105100",
						span: span.New(
							17,
							32,
						),
						expressions: newSet(
							"(?:5[1-5][0-9]{2}|222[1-9]|22[3-9][0-9]|2[3-6][0-9]{2}|27[01][0-9]|2720)[0-9]{12}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "5555555555554444",
						span: span.New(
							65,
							80,
						),
						expressions: newSet(
							"(?:5[1-5][0-9]{2}|222[1-9]|22[3-9][0-9]|2[3-6][0-9]{2}|27[01][0-9]|2720)[0-9]{12}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
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
				output: []*Match{
					{
						subString: "Lorem Ipsum is simply dummy text of the printing and typesetting industry.",
						span: span.New(
							0,
							73,
						),
						expressions: newSet(
							"^.*$",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "Lo",
						span: span.New(
							0,
							1,
						),
						expressions: newSet(
							"^.{2}",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "y.",
						span: span.New(
							72,
							73,
						),
						expressions: newSet(
							".{2}$",
						),
						namedGroups: map[string]span.Interface{},
						groups:      []span.Interface{},
					},
					{
						subString: "Lorem Ipsum is simply dummy text of the printing and typesetting industry.",
						span: span.New(
							0,
							73,
						),
						expressions: newSet(
							"^(.*)$",
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(0, 73),
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
				output: []*Match{
					{
						subString: "<p>simply dummy text</p>",
						span: span.New(
							15,
							38,
						),
						expressions: newSet(
							`<p>(.*)</p>`,
						),
						namedGroups: map[string]span.Interface{},
						groups: []span.Interface{
							span.New(18, 34),
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
					tr, err := New(test.regexps...)
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
	output  []*Match
}

func pointer[T any](x T) *T {
	return &x
}

func Test_It(t *testing.T) { // TODO : move it to examples
	tr, err := New(
		"ba[^for]",
	)
	require.NoError(t, err)

	t.Log(tr.String())

	expected := []*Match{
		{
			subString: "baz",
			span: span.New(
				0,
				2,
			),
			expressions: newSet(
				"ba[^for]",
			),
			namedGroups: map[string]span.Interface{},
			groups:      []span.Interface{},
		},
	}

	sort.SliceStable(expected, func(i, j int) bool {
		return expected[i].Key() < expected[j].Key()
	})

	actual := tr.Match("baz")
	require.NoError(t, err)

	sort.SliceStable(actual, func(i, j int) bool {
		return actual[i].Key() < actual[j].Key()
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
