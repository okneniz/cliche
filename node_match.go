package cliche

import (
	"fmt"

	"github.com/okneniz/cliche/span"
)

type nodeMatch struct {
	node Node
	span span.Interface
}

func (m nodeMatch) String() string {
	return fmt.Sprintf("nodeMatch{%s: %s}", m.span, m.node.GetKey())
}
