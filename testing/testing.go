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
	"testing"

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
		Index int
		Name  string `json:",omitempty"`
		Span
		OutOfString bool
	}

	Expectation struct {
		SubString   string
		Span        Span
		Expressions []string
		Groups      []Group `json:",omitempty"`
	}

	Span struct {
		From  int
		To    int
		Empty bool
	}
)

func (ex *Expectation) String() string {
	s := ex.Span.String()
	s += "-"
	s += ex.groupsToString()
	return s
}

func (m *Expectation) groupsToString() string {
	s := make([]string, 0, len(m.Groups))

	for _, v := range m.Groups {
		if v.Name != "" {
			s = append(s, fmt.Sprintf("%d (%s): %s", v.Index, v.Name, v.Span.String()))
		} else {
			s = append(s, fmt.Sprintf("%d: %s", v.Index, v.Span.String()))
		}
	}

	sort.SliceStable(s, func(i, j int) bool { return s[i] < s[j] })
	return strings.Join(s, ", ")
}

func (ex *Expectation) Normalize() {
	sort.SliceStable(ex.Expressions, func(i, j int) bool {
		return ex.Expressions[i] < ex.Expressions[j]
	})

	sort.SliceStable(ex.Groups, func(i, j int) bool {
		return ex.Groups[i].Index < ex.Groups[j].Index
	})
}

func (s Span) String() string {
	return fmt.Sprintf("[%d, %d, %v]", s.From, s.To, s.Empty)
}

func LoadTestFile(t testing.TB, path string) (*TestFile, error) {
	t.Helper()

	t.Log("open file", path)

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

func LoadAllTestFiles(t testing.TB, dir string) ([]*TestFile, error) {
	t.Helper()

	var files []*TestFile

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".json" {
			testFile, err := LoadTestFile(t, path)
			if err != nil {
				return err
			}

			files = append(files, testFile)
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

			for _, gr := range x.Groups() {
				g := gr.Span

				outOfString := ((g.From() < x.Span().From() && g.To() < x.Span().To()) ||
					(g.From() > x.Span().From() && g.From() > x.Span().To()))

				groups = append(groups, Group{
					Index:       gr.Index,
					Name:        gr.Name,
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

		exs = append(exs, ex)
	}

	return exs
}

func NormalizeExpectations(exps []*Expectation) {
	for _, exp := range exps {
		sort.Slice(exp.Expressions, func(i, j int) bool {
			return strings.Compare(exp.Expressions[i], exp.Expressions[j]) < 1
		})

		sort.Slice(exp.Groups, func(i, j int) bool {
			return exp.Groups[i].Index < exp.Groups[j].Index
		})
	}

	sort.Slice(exps, func(i, j int) bool {
		return exps[i].String() < exps[j].String()
	})
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
	case "case insensitive":
		return node.ScanOptionCaseInsensitive, nil
	case "multiline":
		return node.ScanOptionMultiline, nil
	default:
		return 0, fmt.Errorf("invalid scan options: %v", x)
	}
}
