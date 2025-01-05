package cliche

import (
	"bytes"
	"encoding/json"
)

type Tree interface {
	Add(...string) error
	Size() int
	MarshalJSON() ([]byte, error)
	String() string
	Match(string) []*Match
}

var _ Tree = new(tree)

type tree struct {
	nodes  map[string]Node
	parser Parser
}

func New(parser Parser) *tree {
	tr := new(tree)
	tr.nodes = make(map[string]Node)
	tr.parser = parser
	return tr
}

func (t *tree) Add(strs ...string) error {
	for _, str := range strs {
		buf := newBuffer(str)

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
		x.Traverse(func(_ Node) {
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

func (t *tree) addExpression(str string, newNode Node) {
	newNode.Traverse(func(x Node) {
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

func (t *tree) Match(text string) []*Match {
	if len(text) == 0 {
		return nil
	}

	input := newBuffer(text)
	output := newOutput()
	t.Scan(0, input.Size()-1, input, output)

	return output.Slice()
}

// TODO : what about different scanners?
func (t *tree) Scan(from, to int, input Input, output Output) {
	// TODO : capacity = max count of groups in expression
	captures := newCaptures(10)
	namedCaptures := newNamedCaptures(10)

	scanner := newFullScanner(input, output, captures, namedCaptures)
	skip := func(_ Node, _, _ int, _ bool) {}

	for _, n := range t.nodes {
		nextFrom := from

		for nextFrom <= to {
			lastFrom := nextFrom
			n.Visit(scanner, input, nextFrom, to, skip)

			if pos, ok := output.LastPosOf(n); ok && pos >= nextFrom {
				nextFrom = pos
			}

			if lastFrom == nextFrom {
				nextFrom++
			}

			scanner.Rewind(0)
		}
	}
}
