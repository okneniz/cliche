package cliche

import (
	"bytes"
	"encoding/json"
)

// https://www.regular-expressions.info/repeat.html

// bnf / ebnf
//
// https://www2.cs.sfu.ca/~cameron/Teaching/384/99-3/regexp-plg.html

// do a lot of methods for different scanning
// - for match without allocations
// - for replacements
// - for data extractions
//
// and scanner for all of them?
//
// try to copy official API
//
// https://pkg.go.dev/regexp#Regexp.FindString
//
// https://swtch.com/~rsc/regexp/regexp2.html#posix
//
// https://www.rfc-editor.org/rfc/rfc9485.html#name-multi-character-escapes

type Trie interface {
	Add(...string) error
	Size() int
	MarshalJSON() ([]byte, error)
	String() string
	Match(string) []*stringMatch
}

var _ Trie = new(trie)

type trie struct {
	nodes map[string]Node
}

func NewTrie(regexps ...string) (*trie, error) {
	tr := new(trie)
	tr.nodes = make(map[string]Node)

	for _, regexp := range regexps {
		err := tr.Add(regexp)
		if err != nil {
			return nil, err
		}
	}

	return tr, nil
}

func (t *trie) Add(strs ...string) error {
	for _, str := range strs {
		buf := newBuffer(str)

		node, err := defaultParser(buf)
		if err != nil {
			return err
		}

		t.addExpression(str, node)
	}

	return nil
}

func (t *trie) Size() int {
	size := 0

	for _, x := range t.nodes {
		x.Walk(func(n Node) {
			size++
		})
	}

	return size
}

func (t *trie) MarshalJSON() ([]byte, error) {
	scanner := bytes.NewBuffer(nil)

	encoder := json.NewEncoder(scanner)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", " ")
	err := encoder.Encode(t.nodes)
	if err != nil {
		return nil, err
	}

	return scanner.Bytes(), nil
}

func (t *trie) String() string {
	data, err := t.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(data)
}

func (t *trie) addExpression(str string, newNode Node) {
	newNode.Walk(func(x Node) {
		if len(x.GetNestedNodes()) == 0 {
			x.AddExpression(str)
		}
	})

	key := newNode.GetKey()

	if prev, exists := t.nodes[key]; exists {
		prev.Merge(newNode)
	} else {
		t.nodes[key] = newNode
	}
}

func (t *trie) Match(text string) []*stringMatch {
	if len(text) == 0 {
		return nil
	}

	input := newBuffer(text)
	from := 0
	to := input.Size() - 1
	scanner := newFullScanner(input, from, to)

	return t.Scan(from, to, input, scanner)
}

func (t *trie) Scan(from, to int, input TextBuffer, output Output) []*stringMatch {
	skip := func(_ Node, _, _ int, _ bool) {}

	for _, n := range t.nodes {
		nextFrom := from

		for nextFrom <= to {
			lastFrom := nextFrom
			n.Scan(output, input, nextFrom, to, skip)

			if pos, ok := output.LastPosOf(n); ok && pos >= nextFrom {
				nextFrom = pos
			}

			if lastFrom == nextFrom {
				nextFrom++
			}

			output.Rewind(0)
		}
	}

	return output.Matches()
}
