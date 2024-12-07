package cliche

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_truncatedList(t *testing.T) {
	t.Parallel()

	l := newTruncatedList[num](2)

	require.Equal(t, 0, l.len())

	_, ok := l.first()
	require.False(t, ok)

	_, ok = l.last()
	require.False(t, ok)

	require.Nil(t, l.toSlice())
	require.Equal(t, l.String(), "[]")

	l.append(1)

	require.Equal(t, 1, l.len())

	first, ok := l.first()
	require.True(t, ok)
	require.Equal(t, first, num(1))

	last, ok := l.last()
	require.True(t, ok)
	require.Equal(t, last, num(1))

	require.Equal(t, l.toSlice(), []num{1})
	require.Equal(t, l.String(), "[1]")

	l.append(2)

	require.Equal(t, 2, l.len())

	first, ok = l.first()
	require.True(t, ok)
	require.Equal(t, first, num(1))

	last, ok = l.last()
	require.True(t, ok)
	require.Equal(t, last, num(2))

	require.Equal(t, l.toSlice(), []num{1, 2})
	require.Equal(t, l.String(), "[1, 2]")

	l.append(3)

	require.Equal(t, 3, l.len())

	first, ok = l.first()
	require.True(t, ok)
	require.Equal(t, first, num(1))

	last, ok = l.last()
	require.True(t, ok)
	require.Equal(t, last, num(3))

	require.Equal(t, l.toSlice(), []num{1, 2, 3})
	require.Equal(t, l.String(), "[1, 2, 3]")

	l.append(4)

	require.Equal(t, 4, l.len())

	first, ok = l.first()
	require.True(t, ok)
	require.Equal(t, first, num(1))

	last, ok = l.last()
	require.True(t, ok)
	require.Equal(t, last, num(4))

	require.Equal(t, l.toSlice(), []num{1, 2, 3, 4})
	require.Equal(t, l.String(), "[1, 2, 3, 4]")

	l.truncate(2)

	require.Equal(t, 2, l.len())

	first, ok = l.first()
	require.True(t, ok)
	require.Equal(t, first, num(1))

	last, ok = l.last()
	require.True(t, ok)
	require.Equal(t, last, num(2))

	require.Equal(t, l.toSlice(), []num{1, 2})
	require.Equal(t, l.String(), "[1, 2]")

	l.append(10)

	require.Equal(t, 3, l.len())

	first, ok = l.first()
	require.True(t, ok)
	require.Equal(t, first, num(1))

	last, ok = l.last()
	require.True(t, ok)
	require.Equal(t, last, num(10))

	require.Equal(t, l.toSlice(), []num{1, 2, 10})
	require.Equal(t, l.String(), "[1, 2, 10]")

	l.truncate(0)

	require.Equal(t, 0, l.len())

	_, ok = l.first()
	require.False(t, ok)

	_, ok = l.last()
	require.False(t, ok)

	require.Nil(t, l.toSlice())
	require.Equal(t, l.String(), "[]")

	l.append(10)
	l.append(20)
	l.append(30)

	require.Equal(t, 3, l.len())

	first, ok = l.first()
	require.True(t, ok)
	require.Equal(t, first, num(10))

	last, ok = l.last()
	require.True(t, ok)
	require.Equal(t, last, num(30))

	require.Equal(t, l.toSlice(), []num{10, 20, 30})
	require.Equal(t, l.String(), "[10, 20, 30]")

	require.PanicsWithValue(
		t,
		OutOfBounds{Min: 0, Max: 3, Value: -1},
		func() { l.truncate(-1) },
	)

	require.PanicsWithValue(
		t,
		OutOfBounds{Min: 0, Max: 3, Value: 10},
		func() { l.truncate(10) },
	)

	l.truncate(0)

	require.PanicsWithValue(
		t,
		OutOfBounds{Min: 0, Max: 0, Value: -1},
		func() { l.truncate(-1) },
	)

	require.PanicsWithValue(
		t,
		OutOfBounds{Min: 0, Max: 0, Value: 10},
		func() { l.truncate(10) },
	)

	l.truncate(0)
}

type num int

func (t num) String() string {
	return fmt.Sprintf("%d", t)
}
