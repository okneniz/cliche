package span

import (
	"fmt"
)

type empty int

var (
	_ Interface = empty(0)
)

func Empty(pos int) Interface {
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

func (x empty) IsInclude(_ int) bool {
	return false
}

func (x empty) String() string {
	return fmt.Sprintf("(%d)", x)
}
