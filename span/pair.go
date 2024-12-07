package span

import (
	"fmt"
)

type pair struct {
	from int
	to   int
}

var (
	_ Interface = pair{0, 0}
)

func New(from int, to int) Interface {
	if from > to || from < 0 || to < 0 {
		panic(fmt.Sprintf("invalid bounds %d %d", from, to))
	}
	return pair{
		from: from,
		to:   to,
	}
}

func (p pair) From() int {
	return p.from
}

func (p pair) To() int {
	return p.to
}

func (p pair) Empty() bool {
	return false
}

func (p pair) Size() int {
	return p.to - p.from
}

func (p pair) IsInclude(x int) bool {
	if x < p.from || x > p.to {
		return false
	}

	return true
}

func (p pair) String() string {
	return fmt.Sprintf("[%d-%d]", p.from, p.to)
}
