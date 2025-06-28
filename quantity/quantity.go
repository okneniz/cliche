package quantity

import (
	"fmt"
	"strings"
)

// TODO : move to span

type Quantity struct {
	from int
	to   *int
	more bool
}

func NewQuantity(from int, to int) *Quantity {
	// TODO : validate?

	if from == to {
		return &Quantity{
			from: from,
			to:   nil,
			more: false,
		}
	}

	return &Quantity{
		from: from,
		to:   &to,
		more: false,
	}
}

func NewEndlessQuantity(from int) *Quantity {
	return &Quantity{
		from: from,
		to:   nil,
		more: true,
	}
}

func (n *Quantity) Optional() bool {
	return n.from == 0
}

func (n *Quantity) Gt(value int) bool {
	return n.to == nil || *n.to >= value
}

func (n *Quantity) Include(value int) bool {
	if n.from > value {
		return false
	}

	if n.more {
		return true
	}

	if n.to != nil {
		return value <= *n.to
	}

	return n.from == value
}

func (n *Quantity) String() string {
	if n.from == 0 && n.to == nil && n.more {
		return "*"
	}

	if n.from == 1 && n.to == nil && n.more {
		return "+"
	}

	if n.from == 0 && n.to != nil && *n.to == 1 {
		return "?"
	}

	var b strings.Builder

	b.WriteRune('{')
	b.WriteString(fmt.Sprintf("%d", n.from))

	if n.more {
		b.WriteRune(',')
	} else if n.to != nil && n.from != *n.to {
		b.WriteRune(',')
		b.WriteString(fmt.Sprintf("%d", *n.to))
	}

	b.WriteRune('}')

	return b.String()
}
