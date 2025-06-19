package parser

import (
	"fmt"

	c "github.com/okneniz/parsec/common"
)

// TODO : move to parsec

func Quantifier[T any, P any, S any](from, to int, f c.Combinator[T, P, S]) c.Combinator[T, P, []S] {
	return func(buffer c.Buffer[T, P]) ([]S, error) {
		if from > to {
			panic(fmt.Sprintf("param 'from' must be less than 'to', actual from=%d and to=%d", from, to))
		}

		result := make([]S, 0, to-from)

		for i := 0; i <= to; i++ {
			pos := buffer.Position()

			n, err := f(buffer)
			if err != nil {
				if len(result) >= from {
					buffer.Seek(pos)
					return result, nil
				}

				return nil, err
			}

			result = append(result, n)
		}

		return result, nil
	}
}

func Between[T any, S any](
	before c.Combinator[rune, int, S],
	body c.Combinator[rune, int, T],
	after c.Combinator[rune, int, S],
) c.Combinator[rune, int, T] {
	return c.Between(before, body, after)
}

func Parens[T any](
	body c.Combinator[rune, int, T],
) c.Combinator[rune, int, T] {
	return Between(
		c.Eq[rune, int]('('),
		body,
		c.Eq[rune, int](')'),
	)
}

func Braces[T any](
	body c.Combinator[rune, int, T],
) c.Combinator[rune, int, T] {
	return Between(
		c.Eq[rune, int]('{'),
		body,
		c.Eq[rune, int]('}'),
	)
}

func Angles[T any](
	body c.Combinator[rune, int, T],
) c.Combinator[rune, int, T] {
	return Between(
		c.Eq[rune, int]('<'),
		body,
		c.Eq[rune, int]('>'),
	)
}

func Squares[T any](
	body c.Combinator[rune, int, T],
) c.Combinator[rune, int, T] {
	return Between(
		c.Eq[rune, int]('['),
		body,
		c.Eq[rune, int](']'),
	)
}
