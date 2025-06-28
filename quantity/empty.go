package quantity

import (
	"fmt"
)

type empty int

var (
	_ Interface = empty(0)
)

func Empty(pos int) Interface {
	if pos < 0 {
		panic(fmt.Sprintf("invalid bounds %d", pos))
	}

	return empty(pos)
}

func (x empty) From() int {
	return int(x)
}

func (x empty) To() int {
	return int(x)
}

func (x empty) Empty() bool {
	return true
}

func (x empty) Size() int {
	return 0
}

func (x empty) Include(_ int) bool {
	return false
}

func (x empty) String() string {
	return fmt.Sprintf("(%d)", x)
}
