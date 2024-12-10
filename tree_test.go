package cliche

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// https://www.rfc-editor.org/rfc/rfc9485.html#name-implementing-i-regexp

func TestTree_New(t *testing.T) {
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
	Want        []*Expectation
}

type NamedGroup struct {
	Name string
	Span Span
}

func (g NamedGroup) String() string {
	return fmt.Sprintf("[%s, %s]", g.Name, g.Span)
}

type Expectation struct {
	SubString   string
	Span        Span
	Expressions []string
	Groups      []Span       `json:",omitempty"`
	NamedGroups []NamedGroup `json:",omitempty"`
}

func (ex *Expectation) String() string {
	s := ex.Span.String()
	s += "-"
	s += ex.groupsToString()
	s += "-"
	s += ex.namedGroupsToString()
	return s
}

func (m *Expectation) groupsToString() string {
	s := make([]string, len(m.Groups))
	for i, x := range m.Groups {
		s[i] = x.String()
	}

	sort.SliceStable(s, func(i, j int) bool { return s[i] < s[j] })
	return strings.Join(s, ", ")
}

func (m *Expectation) namedGroupsToString() string {
	pairs := make([]string, 0, len(m.NamedGroups))
	for _, v := range m.NamedGroups {
		pairs = append(pairs, v.Name+": "+v.String())
	}
	sort.SliceStable(pairs, func(i, j int) bool { return pairs[i] < pairs[j] })
	return strings.Join(pairs, ", ")
}

func (ex *Expectation) Normalize() {
	sort.SliceStable(ex.Expressions, func(i, j int) bool {
		return ex.Expressions[i] < ex.Expressions[j]
	})

	sort.SliceStable(ex.Groups, func(i, j int) bool {
		return ex.Groups[i].String() < ex.Groups[j].String()
	})

	sort.SliceStable(ex.NamedGroups, func(i, j int) bool {
		return ex.NamedGroups[i].String() < ex.NamedGroups[j].String()
	})
}

type Span struct {
	From  int
	To    int
	Empty bool
}

func (s Span) String() string {
	return fmt.Sprintf("[%d, %d, %v]", s.From, s.To, s.Empty)
}

func loadTestFile(path string) (*TestFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	testFile := new(TestFile)
	err = json.Unmarshal(data, testFile)
	if err != nil {
		return nil, err
	}

	return testFile, nil
}

func loadAllTestFiles(dir string) ([]*TestFile, error) {
	var files []*TestFile

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".json" {
			fs, err := loadTestFile(path)
			if err != nil {
				return err
			}

			files = append(files, fs)
		}

		return nil
	})

	return files, err
}

func toTestMatches(xs ...*Match) []*Expectation {
	exs := make([]*Expectation, 0, len(xs))

	for _, x := range xs {
		ex := &Expectation{
			SubString: x.subString,
			Span: Span{
				From:  x.span.From(),
				To:    x.span.To(),
				Empty: x.span.Empty(),
			},
			Expressions: x.expressions.Slice(),
		}

		if len(x.groups) > 0 {
			groups := make([]Span, 0, len(x.groups))

			for _, g := range x.groups {
				groups = append(groups, Span{
					From:  g.From(),
					To:    g.To(),
					Empty: g.Empty(),
				})
			}

			ex.Groups = groups
		}

		if len(x.namedGroups) > 0 {
			named := make([]NamedGroup, len(x.namedGroups))

			for k, g := range x.namedGroups {
				named = append(named, NamedGroup{
					Name: k,
					Span: Span{
						From:  g.From(),
						To:    g.To(),
						Empty: g.Empty(),
					},
				})
			}

			ex.NamedGroups = named
		}

		exs = append(exs, ex)
	}

	return exs
}

func TestTree_Match(t *testing.T) {
	t.Parallel()

	files, err := loadAllTestFiles("/Users/andi/dev/golang/regular/testdata/base/")
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

					tr, err := New(test.Expressions...)
					require.NoError(t, err)

					matches := tr.Match(test.Input)
					require.NoError(t, err)

					t.Logf("tree: %s", tr)
					t.Logf("input: '%s'", test.Input)

					actual := toTestMatches(matches...)
					for _, w := range actual {
						w.Normalize()
					}
					sort.Slice(actual, func(i, j int) bool {
						return actual[i].String() < actual[j].String()
					})
					actualStr, err := json.Marshal(actual)
					require.NoError(t, err)

					for _, w := range test.Want {
						w.Normalize()
					}
					sort.Slice(test.Want, func(i, j int) bool {
						return test.Want[i].String() < test.Want[j].String()
					})
					expectedStr, err := json.Marshal(test.Want)
					require.NoError(t, err)

					require.Equal(t, string(expectedStr), string(actualStr))
				})
			}
		})
	}
}

func pointer[T any](x T) *T {
	return &x
}
