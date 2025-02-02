package scanner

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTruncatedList_Append(t *testing.T) {
	type example struct {
		init   []int
		append []int
		want   []int
	}

	examples := []example{
		{
			init:   []int{},
			append: []int{},
			want:   []int{},
		},
		{
			init:   []int{},
			append: []int{1},
			want:   []int{1},
		},
		{
			init:   []int{1},
			append: []int{2, 3},
			want:   []int{1, 2, 3},
		},
		{
			init:   []int{1, 2, 3},
			append: []int{3, 2, 1},
			want:   []int{1, 2, 3, 3, 2, 1},
		},
	}

	for i, example := range examples {
		test := example

		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			list := newTruncatedList[int](0)
			list.Append(test.init...)

			list.Append(test.append...)
			require.Equal(t, list.Slice(), test.want)
		})
	}
}

func TestTruncatedList_At(t *testing.T) {
	type example struct {
		list   []int
		index  int
		exists bool
		value  int
	}

	examples := []example{
		{
			list:   []int{},
			index:  0,
			exists: false,
			value:  0,
		},
		{
			list:   []int{},
			index:  3,
			exists: false,
			value:  0,
		},
		{
			list:   []int{1},
			index:  0,
			exists: true,
			value:  1,
		},
		{
			list:   []int{1},
			index:  3,
			exists: false,
			value:  0,
		},
		{
			list:   []int{1, 2, 3, 4, 5},
			index:  3,
			exists: true,
			value:  4,
		},
	}

	for i, example := range examples {
		test := example

		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			list := newTruncatedList[int](0)
			list.Append(test.list...)

			value, exists := list.At(test.index)
			require.Equal(t, exists, test.exists)
			require.Equal(t, value, test.value)
		})
	}
}

func TestTruncatedList_Truncate(t *testing.T) {
	type example struct {
		init     []int
		truncate int
		want     []int
	}

	examples := []example{
		{
			init:     []int{},
			truncate: 0,
			want:     []int{},
		},
		{
			init:     []int{},
			truncate: 10,
			want:     []int{},
		},
		{
			init:     []int{},
			truncate: -10,
			want:     []int{},
		},
		{
			init:     []int{1},
			truncate: 0,
			want:     []int{},
		},
		{
			init:     []int{1},
			truncate: 1,
			want:     []int{1},
		},
		{
			init:     []int{1},
			truncate: 3,
			want:     []int{1},
		},
		{
			init:     []int{1},
			truncate: -1,
			want:     []int{1},
		},
		{
			init:     []int{1, 2, 3},
			truncate: 0,
			want:     []int{},
		},
		{
			init:     []int{1, 2, 3},
			truncate: -1,
			want:     []int{1, 2, 3},
		},
		{
			init:     []int{1, 2, 3},
			truncate: 1,
			want:     []int{1},
		},
		{
			init:     []int{1, 2, 3},
			truncate: 2,
			want:     []int{1, 2},
		},
		{
			init:     []int{1, 2, 3},
			truncate: 3,
			want:     []int{1, 2, 3},
		},
		{
			init:     []int{1, 2, 3},
			truncate: 4,
			want:     []int{1, 2, 3},
		},
	}

	for i, example := range examples {
		test := example

		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			list := newTruncatedList[int](0)
			list.Append(test.init...)

			require.Equal(t, list.Size(), len(test.init))
			list.Truncate(test.truncate)
			require.Equal(t, list.Slice(), test.want)
			require.Equal(t, list.Size(), len(test.want))
		})
	}
}

func TestTruncatedList_Size(t *testing.T) {
	type example struct {
		list []int
		size int
	}

	examples := []example{
		{
			list: []int{},
			size: 0,
		},
		{
			list: []int{1},
			size: 1,
		},
		{
			list: []int{1, 2, 3},
			size: 3,
		},
	}

	for i, example := range examples {
		test := example

		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			list := newTruncatedList[int](0)
			list.Append(test.list...)
			require.Equal(t, list.Size(), test.size)
		})
	}
}

func TestTruncatedList_First(t *testing.T) {
	type example struct {
		list   []int
		exists bool
		value  int
	}

	examples := []example{
		{
			list:   []int{},
			exists: false,
			value:  0,
		},
		{
			list:   []int{1},
			exists: true,
			value:  1,
		},
		{
			list:   []int{1, 2, 3},
			exists: true,
			value:  1,
		},
	}

	for i, example := range examples {
		test := example

		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			list := newTruncatedList[int](0)
			list.Append(test.list...)

			value, exists := list.First()
			require.Equal(t, exists, test.exists)
			require.Equal(t, value, test.value)
		})
	}
}

func TestTruncatedList_Last(t *testing.T) {
	type example struct {
		list   []int
		exists bool
		value  int
	}

	examples := []example{
		{
			list:   []int{},
			exists: false,
			value:  0,
		},
		{
			list:   []int{1},
			exists: true,
			value:  1,
		},
		{
			list:   []int{1, 2, 3},
			exists: true,
			value:  3,
		},
	}

	for i, example := range examples {
		test := example

		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			list := newTruncatedList[int](0)
			list.Append(test.list...)

			value, exists := list.Last()
			require.Equal(t, exists, test.exists)
			require.Equal(t, value, test.value)
		})
	}
}

func TestTruncatedList_Slice(t *testing.T) {
	type example struct {
		list []int
		want []int
	}

	examples := []example{
		{
			list: []int{},
			want: []int{},
		},
		{
			list: []int{1},
			want: []int{1},
		},
		{
			list: []int{1, 2, 3},
			want: []int{1, 2, 3},
		},
	}

	for i, example := range examples {
		test := example

		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			list := newTruncatedList[int](0)
			require.Equal(t, list.Slice(), []int{})
			list.Append(test.list...)
			require.Equal(t, list.Slice(), test.want)
		})
	}
}
