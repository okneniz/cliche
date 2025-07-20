package testing

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/scanner"
)

type (
	TestFile struct {
		Name  string
		Tests []Test
	}

	Test struct {
		Name        string
		Expressions []string
		Input       string
		Options     []string
		Want        []*Expectation
	}

	Group struct {
		Span
		OutOfString bool
	}

	NamedGroup struct {
		Name        string
		Span        Span
		OutOfString bool
	}

	Expectation struct {
		SubString   string
		Span        Span
		Expressions []string
		Groups      []Group      `json:",omitempty"`
		NamedGroups []NamedGroup `json:",omitempty"`
	}

	Span struct {
		From  int
		To    int
		Empty bool
	}
)

func (g NamedGroup) String() string {
	return fmt.Sprintf("[%s, %s]", g.Name, g.Span)
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

func (s Span) String() string {
	return fmt.Sprintf("[%d, %d, %v]", s.From, s.To, s.Empty)
}

func LoadTestFile(path string) (*TestFile, error) {
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

func LoadAllTestFiles(dir string) ([]*TestFile, error) {
	var files []*TestFile

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".json" {
			testFile, err := LoadTestFile(path)
			if err != nil {
				return err
			}

			files = append(files, testFile)

			for _, test := range testFile.Tests {
				for _, w := range test.Want {
					w.Normalize()
				}

				sort.Slice(test.Want, func(i, j int) bool {
					return test.Want[i].String() < test.Want[j].String()
				})
			}
		}

		return nil
	})

	return files, err
}

func TestMatchesToExpectations(xs ...*scanner.Match) []*Expectation {
	exs := make([]*Expectation, 0, len(xs))

	for _, x := range xs {
		ex := &Expectation{
			SubString: x.SubString(),
			Span: Span{
				From:  x.Span().From(),
				To:    x.Span().To(),
				Empty: x.Span().Empty(),
			},
			Expressions: x.Expressions(),
		}

		if len(x.Groups()) > 0 {
			groups := make([]Group, 0, len(x.Groups()))

			for _, g := range x.Groups() {
				outOfString := ((g.From() < x.Span().From() && g.To() < x.Span().To()) ||
					(g.From() > x.Span().From() && g.From() > x.Span().To()))

				groups = append(groups, Group{
					OutOfString: outOfString,
					Span: Span{
						From:  g.From(),
						To:    g.To(),
						Empty: g.Empty(),
					},
				})
			}

			ex.Groups = groups
		}

		if len(x.NamedGroups()) > 0 {
			named := make([]NamedGroup, 0, len(x.NamedGroups()))

			for k, g := range x.NamedGroups() {
				outOfString := ((g.From() < x.Span().From() && g.To() < x.Span().To()) ||
					(g.From() > x.Span().From() && g.From() > x.Span().To()))

				named = append(named, NamedGroup{
					Name: k,
					Span: Span{
						From:  g.From(),
						To:    g.To(),
						Empty: g.Empty(),
					},
					OutOfString: outOfString,
				})
			}

			ex.NamedGroups = named
		}

		exs = append(exs, ex)
	}

	return exs
}

func ToScanOptions(xs ...string) ([]node.ScanOption, error) {
	opts := make([]node.ScanOption, len(xs))

	for i, x := range xs {
		opt, err := ToScanOption(x)
		if err != nil {
			return nil, err
		}

		opts[i] = opt
	}

	return opts, nil
}

func ToScanOption(x string) (node.ScanOption, error) {
	switch x {
	case "case insensetive":
		return node.ScanOptionCaseInsensetive, nil
	default:
		return 0, fmt.Errorf("invalid scan options: %v", x)
	}
}
