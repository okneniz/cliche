package cliche

import (
	"bytes"
	"encoding/json"

	"github.com/okneniz/cliche/buf"
	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/parser"
	"github.com/okneniz/cliche/scanner"
	c "github.com/okneniz/parsec/common"
)

type Tree interface {
	Add(...string) error
	Size() int
	MarshalJSON() ([]byte, error)
	String() string
	Match(string) []*scanner.Match
}

var (
	_ Tree                = new(tree)
	_ c.Buffer[rune, int] = buf.NewRunesBuffer("")
)

type tree struct {
	nodes  map[string]node.Node
	parser parser.Parser
}

func New(parser parser.Parser) *tree {
	tr := new(tree)
	tr.nodes = make(map[string]node.Node)
	tr.parser = parser
	return tr
}

func (t *tree) Add(strs ...string) error {
	for _, str := range strs {
		buf := buf.NewRunesBuffer(str)

		node, err := t.parser.Parse(buf)
		if err != nil {
			return err
		}

		t.addExpression(str, node)
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

func (t *tree) addExpression(str string, newNode node.Node) {
	newNode.Traverse(func(x node.Node) {
		if len(x.GetNestedNodes()) == 0 { // TODO : strange hack)
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
