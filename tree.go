package cliche

import (
	"bytes"
	"encoding/json"

	"github.com/okneniz/cliche/buf"
	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/scanner"
)

type Tree interface {
	Add(...string) error
	Size() int
	MarshalJSON() ([]byte, error)
	String() string
	Match(string) []*scanner.Match
}

type Parser interface {
	Parse(string) (node.Node, error)
}

var (
	_ Tree = new(tree)
)

type tree struct {
	nodes  map[string]node.Node
	parser Parser
}

func New(parser Parser) *tree {
	tr := new(tree)
	tr.nodes = make(map[string]node.Node)
	tr.parser = parser
	return tr
}

func (t *tree) Add(expressions ...string) error {
	for _, expression := range expressions {
		node, err := t.parser.Parse(expression)
		if err != nil {
			return err
		}

		key := node.GetKey()

		if prev, exists := t.nodes[key]; exists {
			prev.Merge(node)
		} else {
			t.nodes[key] = node
		}
	}

	return nil
}

func (t *tree) Size() int {
	size := 0

	for _, x := range t.nodes {
		x.Traverse(func(_ node.Node) {
			size++
		})
	}

	return size
}

func (t *tree) Match(text string) []*scanner.Match {
	if len(text) == 0 {
		return nil
	}

	input := buf.NewRunesBuffer(text)
	output := scanner.NewOutput()
	scanner := scanner.NewFullScanner(input, output, t.nodes)

	scanner.Scan(0, input.Size()-1)

	return output.Slice()
}

func (t *tree) MarshalJSON() ([]byte, error) {
	data := bytes.NewBuffer(nil)

	encoder := json.NewEncoder(data)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(t.nodes); err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

func (t *tree) String() string {
	data, err := t.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(data)
}
