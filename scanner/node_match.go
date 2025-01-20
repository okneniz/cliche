package scanner

import (
	"fmt"

	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/span"
)

type nodeMatch struct {
	node node.Node
	span span.Interface
}

func (m nodeMatch) String() string {
	return fmt.Sprintf("nodeMatch{%s: %s}", m.span, m.node.GetKey())
}
