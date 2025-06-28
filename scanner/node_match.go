package scanner

import (
	"fmt"

	"github.com/okneniz/cliche/node"
	"github.com/okneniz/cliche/quantity"
)

type nodeMatch struct {
	node   node.Node
	bounds quantity.Interface
}

func (m nodeMatch) String() string {
	return fmt.Sprintf("nodeMatch{%s: %s}", m.bounds, m.node.GetKey())
}
