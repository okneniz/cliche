package parser

import (
	"errors"
	"fmt"

	c "github.com/okneniz/parsec/common"
)

// TODO : move to parsec

func Quantifier[T any](from, to int, f Parser[T]) Parser[[]T] {
	return func(buffer c.Buffer[rune, int]) ([]T, Error) {
		pos := buffer.Position()

		if from > to {
			return nil, Expected(
				"quantifier param 'from' must be less than param 'to'",
				pos,
				fmt.Errorf("quantity from=%d > to=%d", from, to),
			)
		}

		result := make([]T, 0, to-from)

		for i := 0; i <= to; i++ {
			pos := buffer.Position()

			n, err := f(buffer)
			if err != nil {
				if len(result) >= from {
					buffer.Seek(pos)
					return result, nil
				}

				return nil, Expected(
					"invalid count of elements in quantifier",
					pos,
					fmt.Errorf("only %d elements", len(result)),
				)
			}

			result = append(result, n)
		}

		return result, nil
	}
}

func Many[T any](expect string, parseItem Parser[T]) Parser[[]T] {
	return func(buf c.Buffer[rune, int]) ([]T, Error) {
		list := make([]T, 0)

		for !buf.IsEOF() {
			pos := buf.Position()

			x, err := parseItem(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			list = append(list, x)
		}

		return list, nil
	}
}

func Some[T any](expect string, parse Parser[T]) Parser[[]T] {
	return func(buf c.Buffer[rune, int]) ([]T, Error) {
		list := make([]T, 0)

		start := buf.Position()

		for !buf.IsEOF() {
			pos := buf.Position()

			x, err := parse(buf)
			if err != nil {
				buf.Seek(pos)
				break
			}

			list = append(list, x)
		}

		if len(list) == 0 {
			return nil, Expected(expect, start, c.NotEnoughElements)
		}

		return list, nil
	}
}

func Try[T any](parse Parser[T]) Parser[T] {
	return func(buf c.Buffer[rune, int]) (T, Error) {
		pos := buf.Position()

		value, err := parse(buf)
		if err != nil {
			buf.Seek(pos)

			var t T
			return t, err
		}

		return value, nil
	}
}

func Skip[T any, S any](before Parser[S], parse Parser[T]) Parser[T] {
	return func(buf c.Buffer[rune, int]) (T, Error) {
		_, err := before(buf)
		if err != nil {
			var t T
			return t, err
		}

		return parse(buf)
	}
}

func OneOf(xs ...rune) Parser[rune] {
	list := make(map[rune]struct{})
	for _, x := range xs {
		list[x] = struct{}{}
	}

	return func(buf c.Buffer[rune, int]) (rune, Error) {
		pos := buf.Position()

		x, err := buf.Read(true)
		if err != nil {
			return -1, Expected(
				fmt.Sprintf("one of '%s'", string(xs)),
				pos,
				err,
			)
		}

		if _, exists := list[x]; exists {
			return x, nil
		}

		return -1, Expected(
			fmt.Sprintf("one of '%s'", string(xs)),
			pos,
			fmt.Errorf("'%s'", string(x)),
		)
	}
}

func NoneOf(xs ...rune) Parser[rune] {
	list := make(map[rune]struct{})
	for _, x := range xs {
		list[x] = struct{}{}
	}

	return func(buf c.Buffer[rune, int]) (rune, Error) {
		pos := buf.Position()

		x, err := buf.Read(true)
		if err != nil {
			return -1, Expected(
				fmt.Sprintf("none of '%s'", string(xs)),
				pos,
				err,
			)
		}

		if _, exists := list[x]; !exists {
			return x, nil
		}

		return -1, Expected(
			fmt.Sprintf("none of '%s'", string(xs)),
			pos,
			fmt.Errorf("'%s'", string(x)),
		)
	}
}

func Eq(x rune) Parser[rune] {
	return func(buf c.Buffer[rune, int]) (rune, Error) {
		pos := buf.Position()

		y, err := buf.Read(true)
		if err != nil {
			// TODO : remove it?
			if errors.Is(err, c.EndOfFile) {
				return -1, Expected(
					fmt.Sprintf("'%s' rune(%d)", string(x), x),
					pos,
					err,
				)
			}

			return -1, Expected(
				fmt.Sprintf("'%s' rune(%d)", string(x), x),
				pos,
				fmt.Errorf("'%s' rune(%d)", string(y), y),
			)
		}

		if x == y {
			return y, nil
		}

		return -1, Expected(
			fmt.Sprintf("'%s'", string(x)),
			pos,
			fmt.Errorf("'%s'", string(y)),
		)
	}
}

func Between[T any](
	before Parser[rune],
	body Parser[T],
	after Parser[rune],
) Parser[T] {
	var t T

	return func(buf c.Buffer[rune, int]) (T, Error) {
		_, beforeErr := before(buf)
		if beforeErr != nil {
			return t, beforeErr
		}

		value, valErr := body(buf)
		if valErr != nil {
			return t, valErr
		}

		_, afterErr := after(buf)
		if afterErr != nil {
			return t, afterErr
		}

		return value, nil
	}
}

func Parens[T any](body Parser[T]) Parser[T] {
	return Between(Eq('('), body, Eq(')'))
}

func Braces[T any](body Parser[T]) Parser[T] {
	return Between(Eq('{'), body, Eq('}'))
}

func Angles[T any](body Parser[T]) Parser[T] {
	return Between(Eq('<'), body, Eq('>'))
}

func Squares[T any](body Parser[T]) Parser[T] {
	return Between(Eq('['), body, Eq(']'))
}
