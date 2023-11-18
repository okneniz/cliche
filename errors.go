package regular

import (
	"fmt"
)

type OutOfBounds struct {
	Min   int
	Max   int
	Value int
}

func (err OutOfBounds) Error() string {
	return fmt.Sprintf("%d is ouf of bounds %d..%d", err.Value, err.Min, err.Max)
}
