package regular

import (
	c "github.com/okneniz/parsec/common"
)

type key interface {
	comparable
	~rune | ~byte
}

type node[T key] interface {
	Push(node[T])
	Scan()
	ToSlice()
	ToMap()
	Leafs()
}

type root[T key] struct {
	children []node[T]
}

type group[T key] struct {
	variant []node[T]
	children []node[T]
	leaf bool
}

type namedGroup[T key] struct {
	name []T
	variant []node[T]
	children []node[T]
	leaf bool
}

type notCapturedGroup[T key] struct {
	variant []node[T]
	children []node[T]
	leaf bool
}

type any[T key] struct {
	children []node[T]
	leaf bool
}

type digit[T key] struct {
	children []node[T]
	leaf bool
}

type nonDigit[T key] struct {
	children []node[T]
	leaf bool
}

type word[T key] struct {
	children []node[T]
	leaf bool
}

type nonWord[T key] struct {
	children []node[T]
	leaf bool
}

type space[T key] struct {
	children []node[T]
	leaf bool
}

type nonSpace[T key] struct {
	children []node[T]
	leaf bool
}

type startOfLine[T key] struct {
	children []node[T]
	leaf bool
}

type endOfLine[T key] struct {
	children []node[T]
	leaf bool
}

type startOfString[T key] struct {
	children []node[T]
	leaf bool
}

type endOfString[T key] struct {
	children []node[T]
	leaf bool
}

type rangeNode[T key] struct {
	from T
	to T
	children []node[T]
	leaf bool
}

type quantifier[T key] struct {
	count uint
	expression node[T]
	children []node[T]
	leaf bool
}

type many[T key] struct {
	expression node[T]
	children []node[T]
	leaf bool
}

type some[T key] struct {
	expression node[T]
	children []node[T]
	leaf bool
}

type optional[T key] struct {
	expression node[T]
	children []node[T]
	leaf bool
}

type positiveSet[T key] struct {
	set []node[T]
	children []node[T]
	leaf bool
}

type negativeSet[T key] struct {
	set []node[T]
	children []node[T]
	leaf bool
}

func Parse[T key]() (*node[T], error) {

	return nil, nil
}

func parseAny[T key, P any]() c.Combinator[T, P, node[T]] {
	any := c.Eq[T, P]('.')

	return func(buf c.Buffer[T, P]) (node[T], error) {
		_, err := any(buf)
		if err != nil {
			return nil, err
		}

		x := any[T]{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseDigit[T key, P any]() c.Combinator[T, P, node[T]] {
	parse := c.SequenceOf[T, P]('\\', 'd')

	return func(buf c.Buffer[T, P]) (node[T], error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := digit[T]{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseNonDigit[T key, P any]() c.Combinator[T, P, node[T]] {
	parse := c.SequenceOf[T, P]('\\', 'D')

	return func(buf c.Buffer[T, P]) (node[T], error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := nonDigit[T]{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseWord[T key, P any]() c.Combinator[T, P, node[T]] {
	parse := c.SequenceOf[T, P]('\\', 'w')

	return func(buf c.Buffer[T, P]) (node[T], error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := word[T]{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseNonWord[T key, P any]() c.Combinator[T, P, node[T]] {
	parse := c.SequenceOf[T, P]('\\', 'w')

	return func(buf c.Buffer[T, P]) (node[T], error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := nonWord[T]{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseSpace[T key, P any]() c.Combinator[T, P, node[T]] {
	parse := c.SequenceOf[T, P]('\\', 's')

	return func(buf c.Buffer[T, P]) (node[T], error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := space[T]{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseNonSpace[T key, P any]() c.Combinator[T, P, node[T]] {
	parse := c.SequenceOf[T, P]('\\', 'S')

	return func(buf c.Buffer[T, P]) (node[T], error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := nonSpace[T]{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseStartOfLine[T key, P any]() c.Combinator[T, P, node[T]] {
	parse := c.SequenceOf[T, P]('\\', '^')

	return func(buf c.Buffer[T, P]) (node[T], error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := startOfLine[T]{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseEndOfLine[T key, P any]() c.Combinator[T, P, node[T]] {
	parse := c.SequenceOf[T, P]('\\', '$')

	return func(buf c.Buffer[T, P]) (node[T], error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := endOfLine[T]{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseStartOfString[T key, P any]() c.Combinator[T, P, node[T]] {
	parse := c.SequenceOf[T, P]('\\', 'A')

	return func(buf c.Buffer[T, P]) (node[T], error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := startOfString[T]{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseEndOfString[T key, P any]() c.Combinator[T, P, node[T]] {
	parse := c.SequenceOf[T, P]('\\', 'z')

	return func(buf c.Buffer[T, P]) (node[T], error) {
		_, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := endOfString[T]{
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseGroup[T key, P any](
	expression c.Combinator[T, P, node[T]],
) {
	sep := c.Eq[T, P]('|')
	union := c.SepBy1[T, P](0, expression, sep)
	before := c.Eq[T, P]('(')
	after := c.Eq[T, P](')')

	return func(buf c.Buffer[T, P]) (node[T], error) {
		_, err := before(buf)
		if err != nil {
			return nil, err
		}

		variants, err := union(buf)
		if err != nil {
			return nil, err
		}

		_, err = after(buf)
		if err != nil {
			return nil, err
		}

		x := group[T]{
			variants: variants,
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseNotCapturedGroup[T key, P any](
	expression c.Combinator[T, P, node[T]],
) {
	sep := c.Eq[T, P]('|')
	union := c.SepBy1[T, P](0, expression, sep)
	before := c.SequenceOf[T, P]('(', '?', ':')
	after := c.Eq[T, P](')')

	return func(buf c.Buffer[T, P]) (node[T], error) {
		_, err := before(buf)
		if err != nil {
			return nil, err
		}

		variants, err := union(buf)
		if err != nil {
			return nil, err
		}

		_, err = after(buf)
		if err != nil {
			return nil, err
		}

		x := notCapturedGroup[T]{
			variants: variants,
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseNamedGroup[T key, P any](
	expression c.Combinator[T, P, node[T]],
) {
	sep := c.Eq[T, P]('|')
	union := c.SepBy1[T, P](0, expression, sep)
	before := c.SequenceOf[T, P]('(', '?', '"')
	after := c.Eq[T, P](')')

	groupName := c.Between[T, P](
		c.SequenceOf[T, P]('<', '?'),
		c.Many[T, P](0, c.NoneOf[T, P]('>')),
		c.Eq[T, P]('>'),
	)

	return func(buf c.Buffer[T, P]) (node[T], error) {
		_, err := before(buf)
		if err != nil {
			return nil, err
		}

		name, err := groupName(buf)
		if err != nil {
			return nil, err
		}

		variants, err := union(buf)
		if err != nil {
			return nil, err
		}

		_, err = after(buf)
		if err != nil {
			return nil, err
		}

		x := namedGroup[T]{
			name: name,
			variants: variants,
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseNegativeSet[T key, P any](
	expression c.Combinator[T, P, node[T]],
) {
	parse := c.Between[T, P](
		c.SequenceOf[T, P]('[', '^'),
		c.Many[T, P](expression),
		c.Eq[T, P](']'),
	)

	return func(buf c.Buffer[T, P]) (node[T], error) {
		set, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := negativeSet[T]{
			set: set,
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parsePositiveSet[T key, P any](
	expression c.Combinator[T, P, node[T]],
) {
	parse := c.Between[T, P](
		c.Eq[T, P]('['),
		c.Many[T, P](expression),
		c.Eq[T, P](']'),
	)

	return func(buf c.Buffer[T, P]) (node[T], error) {
		set, err := parse(buf)
		if err != nil {
			return nil, err
		}

		x := positiveSet[T]{
			set: set,
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}

func parseRange[T key, P any](
	except []T,
	expression c.Combinator[T, P, node[T]],
) {
	from := c.NoneOf[T, P](except...)
	to := c.NoneOf[T, P](except...)
	sep := c.Eq[T, P]('-')

	return func(buf c.Buffer[T, P]) (node[T], error) {
		f, err := from(buf)
		if err != nil {
			return nil, err
		}

		_, err = sep(buf)
		if err != nil {
			return nil, err
		}

		t, err := to(buf)
		if err != nil {
			return nil, err
		}

		x := rangeNode[T]{
			from: f,
			to: t,
			leaf: buf.IsEOF(),
		}

		return x, nil
	}
}
