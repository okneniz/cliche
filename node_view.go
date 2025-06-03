package cliche

import (
	"fmt"

	"github.com/okneniz/cliche/node"
	// pp "github.com/okneniz/cliche/parser"
	"golang.org/x/exp/maps"
)

type nodeView struct {
	Key         string      `json:"key"`
	NodeType    string      `json:"type"`
	Address     string      `json:"address"`
	Expressions []string    `json:"expressions,omitempty"`
	Value       *nodeView   `json:"value,omitempty"`
	Variants    []*nodeView `json:"variants,omitempty"`
	Nested      []*nodeView `json:"nested,omitempty"`
}

func newView(n node.Node) *nodeView {
	v := &nodeView{
		Key:         n.GetKey(),
		NodeType:    fmt.Sprintf("%T", n),
		Address:     fmt.Sprintf("node=%p nested=%p expressions=%p", n, n.GetNestedNodes(), n.GetExpressions()),
		Expressions: n.GetExpressions().Slice(),
	}

	if c, ok := n.(node.Container); ok {
		v.Value = newView(c.GetValue())
	}

	if alt, ok := n.(node.Alternation); ok {
		v.Variants = newViewsList(alt.GetVariants())
	}

	if len(n.GetNestedNodes()) > 0 {
		v.Nested = newViewsList(maps.Values(n.GetNestedNodes()))
	}

	return v
}

func newViewsList(ns []node.Node) []*nodeView {
	views := make([]*nodeView, 0, len(ns))

	for _, x := range ns {
		views = append(views, newView(x))
	}

	return views
}
