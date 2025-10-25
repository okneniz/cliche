package cliche

import (
	"bytes"
	"encoding/json"

	"golang.org/x/exp/maps"

	"github.com/okneniz/cliche/buf"
	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/scanner"
)

type Tree interface {
	// Add - add regular expressions to tree
	Add(...string) error
	// Size - return count of nodes in tree
	Size() int
	// String - dump tree to string
	String() string
	// Match - scan text and return matches
	Match(text string, options ...node.ScanOption) []*scanner.Match
}

type Parser interface {
	// Parse - parse regular expression and return node.Alternation as base type for it
	Parse(string) (node.Alternation, error)
}

var (
	_ Tree = new(tree)
)

type tree struct {
	nodes  map[string]node.Node
	parser Parser
}

// New - return Tree
func New(parser Parser) Tree {
	tr := new(tree)
	tr.nodes = make(map[string]node.Node)
	tr.parser = parser
	return tr
}

// Add - add regular expressions to tree
func (t *tree) Add(expressions ...string) error {
	for _, expression := range expressions {
		raw, err := t.parser.Parse(expression)
		if err != nil {
			return err
		}

		if err := node.Validate(expression, raw); err != nil {
			return err
		}

		for _, newNode := range node.Unify(raw) {
			key := newNode.GetKey()

			if oldNode, exists := t.nodes[key]; exists {
				t.merge(oldNode, newNode)
			} else {
				t.nodes[key] = newNode
			}
		}
	}

	return nil
}

func (t *tree) merge(oldNode, newNode node.Node) {
	for key, newNestedNode := range newNode.GetNestedNodes() {
		if oldNestedNode, exists := oldNode.GetNestedNodes()[key]; exists {
			t.merge(oldNestedNode, newNestedNode)
		} else {
			oldNode.GetNestedNodes()[key] = newNestedNode
		}
	}

	// only if oldNode.isLeaf()
	newNode.GetExpressions().AddTo(oldNode.GetExpressions())
}

// Size - return count of nodes in tree
func (t *tree) Size() int {
	size := 0

	for _, x := range t.nodes {
		node.Traverse(x, func(n node.Node) bool {
			size++
			return false
		})
	}

	return size
}

// Match - scan text and return matches
func (t *tree) Match(text string, options ...node.ScanOption) []*scanner.Match {
	input := buf.NewRunesBuffer(text)
	output := scanner.NewOutput()
	scanner := scanner.NewFullScanner(input, output, t.nodes, options...)

	scanner.Scan(0, input.Size())

	return output.Slice()
}

func (t *tree) marshalJSON() ([]byte, error) {
	data := bytes.NewBuffer(nil)

	encoder := json.NewEncoder(data)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", " ")

	x := newViewsList(maps.Values(t.nodes))

	if err := encoder.Encode(x); err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

// String - dump tree to string
func (t *tree) String() string {
	data, err := t.marshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(data)
}
