package cliche

import (
	"bytes"

	c "github.com/okneniz/parsec/common"
)

type branch[T any] struct {
	parser   c.Combinator[rune, int, T]
	Children map[rune]*branch[T]
}

func newParserBranches[T any](
	cases map[string]c.Combinator[rune, int, T],
) *branch[T] {
	t := new(branch[T])
	t.Children = make(map[rune]*branch[T])

	for cs, parser := range cases {
		current := t

		for _, r := range cs {
			// TODO : handle conflicts

			child, exists := current.Children[r]
			if !exists {
				child = &branch[T]{
					Children: make(map[rune]*branch[T]),
				}

				current.Children[r] = child
			}

			current = child
		}

		current.parser = parser
	}

	return t
}

// TODO : remove it
func (b *branch[T]) String() string {
	var str func(prefix string, br *branch[T], output *bytes.Buffer)

	str = func(prefix string, br *branch[T], output *bytes.Buffer) {
		pre := prefix + " "

		for k, v := range br.Children {
			output.WriteString(prefix + "- '" + string(k) + "'")
			str(pre, v, output)
		}
	}

	buf := bytes.NewBuffer(nil)
	str("\n", b, buf)

	return buf.String()
}
