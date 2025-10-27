package re2_test

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/okneniz/cliche"
	"github.com/okneniz/cliche/re2"
	"github.com/stretchr/testify/require"
	// "github.com/stretchr/testify/require"
)

// TestCase represents a single test case
type TestCase struct {
	Type     string  // B, E, BE, etc.
	Pattern  string  // Regular expression pattern
	String   string  // String to test against
	Expected []Match // Expected match positions
	LineNum  int     // Line number in source file
	Flags    int     // Computed flags for regex compilation
}

// Match represents a single match position
type Match struct {
	Start int
	End   int
}

func (m Match) String() string {
	return fmt.Sprintf("(%v,%v)", m.Start, m.End)
}

// Constants for regex flags
const (
	FLAG_BASIC    = 1 << iota // Basic Regular Expressions
	FLAG_EXTENDED             // Extended Regular Expressions
	FLAG_NEWLINE              // REG_NEWLINE behavior
)

// ParseType parses the type string and returns appropriate flags
func ParseType(typeStr string) int {
	flags := 0
	for _, char := range typeStr {
		switch char {
		case 'B':
			flags |= FLAG_BASIC
		case 'E':
			flags |= FLAG_EXTENDED
		case 'N':
			flags |= FLAG_NEWLINE
		}
	}
	return flags
}

// ParseTestFile parses the testregex.c format
func ParseTestFile(t *testing.T, filename string) ([]TestCase, error) {
	t.Helper()

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var testCases []TestCase
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if shouldSkipLine(line) {
			continue
		}

		testCase, err := parseLine(line, lineNum)
		if err != nil {
			t.Logf("Warning: line %d: %v", lineNum, err)
			continue
		}

		if testCase != nil {
			testCases = append(testCases, *testCase)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return testCases, nil
}

func shouldSkipLine(line string) bool {
	return line == "" ||
		strings.HasPrefix(line, "NOTE") ||
		strings.HasPrefix(line, "#") ||
		strings.HasPrefix(line, "//") ||
		strings.HasPrefix(line, "--")
}

func parseLine(line string, lineNum int) (*TestCase, error) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return nil, nil
	}

	testCase := &TestCase{
		Type:    fields[0],
		Pattern: fields[1],
		LineNum: lineNum,
		Flags:   ParseType(fields[0]),
	}

	switch fields[2] {
	case "NULL", "null", "NIL":
		testCase.String = ""
	default:
		testCase.String = fields[2]
	}

	if len(fields) > 3 {
		matches, err := parseMatches(fields[3:])
		if err != nil {
			return nil, err
		}
		testCase.Expected = matches
	}

	if testCase.Expected == nil {
		testCase.Expected = make([]Match, 0)
	}

	return testCase, nil
}

func parseMatches(matchStrs []string) ([]Match, error) {
	var matches []Match

	for _, matchStr := range matchStrs {
		clean := strings.Trim(matchStr, "()")
		if clean == "" {
			continue
		}

		parts := strings.Split(clean, ",")
		if len(parts) != 2 {
			return nil, nil
		}

		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}

		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}

		matches = append(matches, Match{Start: start, End: end})
	}

	return matches, nil
}

func TestATTPosixRegex(t *testing.T) {
	t.Parallel()

	filename := "../testdata/at&t_posix/basic.dat"
	tests, err := ParseTestFile(t, filename)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		test := tt

		if (test.Type != "B") && (test.Type != "BE") {
			continue
		}

		if shouldSkipType(test.Type) {
			t.Skipf("Type %s not fully supported", test.Type)
		}

		t.Run(fmt.Sprintf("%d-line", tt.LineNum), func(t *testing.T) {
			t.Parallel()

			t.Log("type", test.Type)
			t.Log("pattern", test.Pattern)
			t.Log("string", test.String)

			tr := cliche.New(re2.Parser)

			err := tr.Add(test.Pattern)
			if err != nil {
				t.Fatal(err)
			}

			t.Log("tree:", tr)

			result := tr.Match(test.String)

			t.Log("expected", test.Expected)
			t.Log("actual", result)

			mathes := make([]Match, 0, len(result))
			for _, x := range result {
				t.Log("match", test.Type, x, x.Span().Empty())

				if x.Span().Empty() {
					if len(result) == 1 {
						mathes = append(mathes, Match{
							Start: x.Span().From(),
							End:   x.Span().To(),
						})
					}

					continue
				}

				mathes = append(mathes, Match{
					Start: x.Span().From(),
					End:   x.Span().To() + 1,
				})

				if test.Type == "BE" {
					break // one match
				}

				for _, g := range x.Groups() {
					t.Log("group match", g, g.Empty())
					if g.Empty() {
						continue
					}

					mathes = append(mathes, Match{
						Start: g.From(),
						End:   g.To() + 1,
					})
				}
			}

			t.Log("matches", mathes, test.Expected == nil)
			require.EqualValues(t, test.Expected, mathes)

			// re := regexp.MustCompile(test.Pattern)
			// res := re.FindAllStringSubmatchIndex(test.String, -1)
			// t.Log("wtf", res)

		})
	}
}

func shouldSkipType(typeStr string) bool {
	// Skip types that have significantly different behavior in Go
	unsupported := []string{"A", "S", "BN", "EN", "BEN"}
	for _, unsup := range unsupported {
		if typeStr == unsup {
			return true
		}
	}
	return false
}
