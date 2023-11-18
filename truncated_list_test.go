package regular

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_list(t *testing.T) {
	t.Parallel()

	l := newList[num](2)

	require.Equal(t, 0, l.len())
	require.Nil(t, l.first())
	require.Nil(t, l.last())
	require.Nil(t, l.toSlice())
	require.Equal(t, l.String(), "[]")

	l.append(1)

	require.Equal(t, 1, l.len())
	require.Equal(t, *l.first(), num(1))
	require.Equal(t, *l.last(), num(1))
	require.Equal(t, l.toSlice(), []num{1})
	require.Equal(t, l.String(), "[1]")

	l.append(2)

	require.Equal(t, 2, l.len())
	require.Equal(t, *l.first(), num(1))
	require.Equal(t, *l.last(), num(2))
	require.Equal(t, l.toSlice(), []num{1, 2})
	require.Equal(t, l.String(), "[1, 2]")

	l.append(3)

	require.Equal(t, 3, l.len())
	require.Equal(t, *l.first(), num(1))
	require.Equal(t, *l.last(), num(3))
	require.Equal(t, l.toSlice(), []num{1, 2, 3})
	require.Equal(t, l.String(), "[1, 2, 3]")

	l.append(4)

	require.Equal(t, 4, l.len())
	require.Equal(t, *l.first(), num(1))
	require.Equal(t, *l.last(), num(4))
	require.Equal(t, l.toSlice(), []num{1, 2, 3, 4})
	require.Equal(t, l.String(), "[1, 2, 3, 4]")

	l.truncate(2)

	require.Equal(t, 2, l.len())
	require.Equal(t, *l.first(), num(1))
	require.Equal(t, *l.last(), num(2))
	require.Equal(t, l.toSlice(), []num{1, 2})
	require.Equal(t, l.String(), "[1, 2]")

	l.append(10)

	require.Equal(t, 3, l.len())
	require.Equal(t, *l.first(), num(1))
	require.Equal(t, *l.last(), num(10))
	require.Equal(t, l.toSlice(), []num{1, 2, 10})
	require.Equal(t, l.String(), "[1, 2, 10]")

	l.truncate(0)

	require.Equal(t, 0, l.len())
	require.Nil(t, l.first())
	require.Nil(t, l.last())
	require.Nil(t, l.toSlice())
	require.Equal(t, l.String(), "[]")

	l.append(10)
	l.append(20)
	l.append(30)

	require.Equal(t, 3, l.len())
	require.Equal(t, *l.first(), num(10))
	require.Equal(t, *l.last(), num(30))
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
