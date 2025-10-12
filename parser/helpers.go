package parser

import (
	"fmt"

	c "github.com/okneniz/parsec/common"
)

// TODO : move to parsec

func Quantifier[T any, P any, S any](
	errMessage string,
	from, to int,
	f c.Combinator[T, P, S],
) (c.Combinator[T, P, []S], error) {
	if from > to {
		return nil, fmt.Errorf(
			"param 'from' must be less than param 'to', actual from=%d, to=%d",
			from,
			to,
		)
	}

	// TODO : simplify it
	return func(buffer c.Buffer[T, P]) ([]S, c.Error[P]) {
		start := buffer.Position()
		result := make([]S, 0, to-from)

		for i := 0; i <= to; i++ {
			pos := buffer.Position()

			n, err := f(buffer)
			if err != nil {
				if len(result) >= from {
					if seekErr := buffer.Seek(pos); seekErr != nil {
						return nil, c.NewParseError(
							buffer.Position(),
							seekErr.Error(),
							err,
						)
					}

					return result, nil
				}

				if seekErr := buffer.Seek(start); seekErr != nil {
					return nil, c.NewParseError(
						buffer.Position(),
						seekErr.Error(),
						err,
					)
				}

				return nil, c.NewParseError(
					start,
					fmt.Sprintf(
						"expected between %d and %d, actual %d",
						from,
						to,
						len(result),
					),
				)
			}

			result = append(result, n)
		}

		return result, nil
	}, nil
}
