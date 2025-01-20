package parser

import (
	"fmt"

	c "github.com/okneniz/parsec/common"
)

func SkipString(data string) c.Combinator[rune, int, struct{}] {
	none := struct{}{}

	return func(buf c.Buffer[rune, int]) (struct{}, error) {
		l := len(data)
		for _, x := range data {
			r, err := buf.Read(true)
			if err != nil {
				return none, err
			}
			if x != r {
				return none, c.NothingMatched
			}
			l -= 1
		}

		if l != 0 {
			return none, c.NothingMatched
		}

		return none, nil
	}
}

func TryAll[T any](parsers ...c.Combinator[rune, int, T]) c.Combinator[rune, int, T] {
	attempts := make([]c.Combinator[rune, int, T], len(parsers))

	for i, parse := range parsers {
		attempts[i] = c.Try(parse)
	}

	return c.Choice(attempts...)
}

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

func Number() c.Combinator[rune, int, int] {
	digit := c.Try[rune, int](c.Range[rune, int]('0', '9'))
	zero := rune('0')

	return func(buf c.Buffer[rune, int]) (int, error) {
		token, err := digit(buf)
		if err != nil {
			return 0, err
		}

		result := int(token - zero)
		for {
			token, err = digit(buf)
			if err != nil {
				break
			}

			result = result * 10
			result += int(token - zero)
		}

		return result, nil
	}
}
