package scanner

import (
	"fmt"

	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/quantity"
)

type nodeMatch struct {
	node node.Node
	span quantity.Interface // TODO : rename to bounds
}

func (m nodeMatch) String() string {
	return fmt.Sprintf("nodeMatch{%s: %s}", m.span, m.node.GetKey())
}
