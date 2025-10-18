package structs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSet_Add(t *testing.T) {
	t.Parallel()

	type example struct {
		init   []string
		add    []string
		output []string
	}

	examples := []example{
		{
			init:   []string{"1", "2", "3"},
			add:    []string{"4"},
			output: []string{"1", "2", "3", "4"},
		},
		{
			init:   []string{},
			add:    []string{"4"},
			output: []string{"4"},
		},
		{
			init:   []string{"1", "2", "3"},
			add:    []string{},
			output: []string{"1", "2", "3"},
		},
		{
			init:   []string{"1", "2", "3"},
			add:    []string{"2", "4"},
			output: []string{"1", "2", "3", "4"},
		},
	}

	for _, example := range examples {
		set := NewMapSet[string](example.init...)
		set.Add(example.add...)

		require.Subset(t, example.output, set.Slice())
		require.Equal(t, len(example.output), len(set.Slice()))
	}
}

func TestSet_Size(t *testing.T) {
	t.Parallel()

	type example struct {
		init []string
		add  []string
		size int
	}

	examples := []example{
		{
			init: []string{"1", "2", "3"},
			add:  []string{"4"},
			size: 4,
		},
		{
			init: []string{},
			add:  []string{"4"},
			size: 1,
		},
		{
			init: []string{"1", "2", "3"},
			add:  []string{},
			size: 3,
		},
		{
			init: []string{"1", "2", "3"},
			add:  []string{"2", "4"},
			size: 4,
		},
	}

	for _, example := range examples {
		set := NewMapSet[string](example.init...)
		set.Add(example.add...)

		require.Equal(t, set.Size(), example.size)
	}
}

func TestSet_AddTo(t *testing.T) {
	t.Parallel()

	type state struct {
		first  []string
		second []string
	}

	type example struct {
		before state
		after  state
	}

	examples := []example{
		{
			before: state{
				first:  []string{},
				second: []string{},
			},
			after: state{
				first:  []string{},
				second: []string{},
			},
		},
		{
			before: state{
				first:  []string{"1"},
				second: []string{"2"},
			},
			after: state{
				first:  []string{"1"},
				second: []string{"1", "2"},
			},
		},
		{
			before: state{
				first:  []string{"1", "2"},
				second: []string{"2"},
			},
			after: state{
				first:  []string{"1", "2"},
				second: []string{"1", "2"},
			},
		},
		{
			before: state{
				first:  []string{"1", "2", "3"},
				second: []string{"1", "2", "3"},
			},
			after: state{
				first:  []string{"1", "2", "3"},
				second: []string{"1", "2", "3"},
			},
		},
		{
			before: state{
				first:  []string{"1", "2", "3"},
				second: []string{},
			},
			after: state{
				first:  []string{"1", "2", "3"},
				second: []string{"1", "2", "3"},
			},
		},
		{
			before: state{
				first:  []string{"1"},
				second: []string{"1", "2", "3"},
			},
			after: state{
				first:  []string{"1"},
				second: []string{"1", "2", "3"},
			},
		},
	}

	for _, example := range examples {
		first := NewMapSet[string](example.before.first...)
		second := NewMapSet[string](example.after.second...)

		first.AddTo(second)

		require.Subset(t, example.after.first, first.Slice())
		require.Equal(t, len(example.after.first), len(first.Slice()))

		require.Subset(t, example.after.second, second.Slice())
		require.Equal(t, len(example.after.second), len(second.Slice()))
	}
}

func TestSet_Clone(t *testing.T) {
	t.Parallel()

	examples := [][]string{
		{},
		{"1"},
		{"1", "2", "3"},
	}

	for _, example := range examples {
		set := NewMapSet[string](example...)

		require.Subset(t, set.Slice(), set.Clone().Slice())
		require.Equal(t, len(set.Slice()), len(set.Clone().Slice()))
	}
}
