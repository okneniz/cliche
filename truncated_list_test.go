package cliche

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_truncatedList(t *testing.T) {
	t.Parallel()

	l := newTruncatedList[num](2)

	require.Equal(t, 0, l.Size())

	_, ok := l.First()
	require.False(t, ok)

	_, ok = l.Last()
	require.False(t, ok)

	require.Nil(t, l.Slice())
	require.Equal(t, l.String(), "[]")

	l.Append(1)

	require.Equal(t, 1, l.Size())

	first, ok := l.First()
	require.True(t, ok)
	require.Equal(t, first, num(1))

	last, ok := l.Last()
	require.True(t, ok)
	require.Equal(t, last, num(1))

	require.Equal(t, l.Slice(), []num{1})
	require.Equal(t, l.String(), "[1]")

	l.Append(2)

	require.Equal(t, 2, l.Size())

	first, ok = l.First()
	require.True(t, ok)
	require.Equal(t, first, num(1))

	last, ok = l.Last()
	require.True(t, ok)
	require.Equal(t, last, num(2))

	require.Equal(t, l.Slice(), []num{1, 2})
	require.Equal(t, l.String(), "[1, 2]")

	l.Append(3)

	require.Equal(t, 3, l.Size())

	first, ok = l.First()
	require.True(t, ok)
	require.Equal(t, first, num(1))

	last, ok = l.Last()
	require.True(t, ok)
	require.Equal(t, last, num(3))

	require.Equal(t, l.Slice(), []num{1, 2, 3})
	require.Equal(t, l.String(), "[1, 2, 3]")

	l.Append(4)

	require.Equal(t, 4, l.Size())

	first, ok = l.First()
	require.True(t, ok)
	require.Equal(t, first, num(1))

	last, ok = l.Last()
	require.True(t, ok)
	require.Equal(t, last, num(4))

	require.Equal(t, l.Slice(), []num{1, 2, 3, 4})
	require.Equal(t, l.String(), "[1, 2, 3, 4]")

	l.Truncate(2)

	require.Equal(t, 2, l.Size())

	first, ok = l.First()
	require.True(t, ok)
	require.Equal(t, first, num(1))

	last, ok = l.Last()
	require.True(t, ok)
	require.Equal(t, last, num(2))

	require.Equal(t, l.Slice(), []num{1, 2})
	require.Equal(t, l.String(), "[1, 2]")

	l.Append(10)

	require.Equal(t, 3, l.Size())

	first, ok = l.First()
	require.True(t, ok)
	require.Equal(t, first, num(1))

	last, ok = l.Last()
	require.True(t, ok)
	require.Equal(t, last, num(10))

	require.Equal(t, l.Slice(), []num{1, 2, 10})
	require.Equal(t, l.String(), "[1, 2, 10]")

	l.Truncate(0)

	require.Equal(t, 0, l.Size())

	_, ok = l.First()
	require.False(t, ok)

	_, ok = l.Last()
	require.False(t, ok)

	require.Nil(t, l.Slice())
	require.Equal(t, l.String(), "[]")

	l.Append(10)
	l.Append(20)
	l.Append(30)

	require.Equal(t, 3, l.Size())

	first, ok = l.First()
	require.True(t, ok)
	require.Equal(t, first, num(10))

	last, ok = l.Last()
	require.True(t, ok)
	require.Equal(t, last, num(30))

	require.Equal(t, l.Slice(), []num{10, 20, 30})
	require.Equal(t, l.String(), "[10, 20, 30]")

	require.NotPanics(t, func() { l.Truncate(-1) })
	require.NotPanics(t, func() { l.Truncate(10) })

	l.Truncate(0)

	require.NotPanics(t, func() { l.Truncate(-1) })
	require.NotPanics(t, func() { l.Truncate(10) })

	l.Truncate(0)
}

type num int

func (t num) String() string {
	return fmt.Sprintf("%d", t)
}
