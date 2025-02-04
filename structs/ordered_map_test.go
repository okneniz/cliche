package structs

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderedMap_Get(t *testing.T) {
	type pair struct {
		key   string
		value int
	}

	type example struct {
		input    []pair
		truncate int
		key      string
		value    int
		exists   bool
	}

	examples := []example{
		{
			input:    []pair{},
			truncate: -1,
			key:      "foo",
			value:    0,
			exists:   false,
		},
		{
			input:    []pair{},
			truncate: -1,
			key:      "bar",
			value:    0,
			exists:   false,
		},
		{
			input:    []pair{},
			truncate: -1,
			key:      "baz",
			value:    0,
			exists:   false,
		},
		{
			input:    []pair{},
			truncate: 0,
			key:      "baz",
			value:    0,
			exists:   false,
		},
		{
			input:    []pair{},
			truncate: 10,
			key:      "baz",
			value:    0,
			exists:   false,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
			},
			truncate: -1,
			key:      "foo",
			value:    1,
			exists:   true,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
			},
			truncate: -1,
			key:      "bar",
			value:    2,
			exists:   true,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
			},
			truncate: 0,
			key:      "bar",
			value:    0,
			exists:   false,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
			},
			truncate: 1,
			key:      "foo",
			value:    1,
			exists:   true,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
			},
			truncate: 1,
			key:      "bar",
			value:    0,
			exists:   false,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
				{"bar", 3},
			},
			truncate: -1,
			key:      "bar",
			value:    2,
			exists:   true,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
				{"bar", 3},
			},
			truncate: -1,
			key:      "foo",
			value:    1,
			exists:   true,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
				{"bar", 3},
			},
			truncate: 1,
			key:      "bar",
			value:    0,
			exists:   false,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
				{"bar", 3},
			},
			truncate: 2,
			key:      "bar",
			value:    2,
			exists:   true,
		},
	}

	for i, example := range examples {
		name := fmt.Sprintf("case %d", i)
		test := example

		t.Run(name, func(t *testing.T) {
			m := NewOrderedMap[string, int](0)

			for _, pair := range test.input {
				m.Put(pair.key, pair.value)
			}

			if test.truncate >= 0 {
				m.Truncate(test.truncate)
			}

			value, exists := m.Get(test.key)
			require.Equal(t, test.value, value)
			require.Equal(t, test.exists, exists)
		})
	}
}

func TestOrderedMap_Put(t *testing.T) {
	type pair struct {
		key   string
		value int
	}

	type example struct {
		input    []pair
		truncate int
		data     map[string]int
	}

	examples := []example{
		{
			input:    []pair{},
			truncate: -1,
			data:     map[string]int{},
		},
		{
			input:    []pair{},
			truncate: 0,
			data:     map[string]int{},
		},
		{
			input:    []pair{},
			truncate: 1,
			data:     map[string]int{},
		},
		{
			input: []pair{
				{"foo", 1},
			},
			truncate: -1,
			data: map[string]int{
				"foo": 1,
			},
		},
		{
			input: []pair{
				{"foo", 1},
			},
			truncate: 0,
			data:     map[string]int{},
		},
		{
			input: []pair{
				{"foo", 1},
			},
			truncate: 10,
			data: map[string]int{
				"foo": 1,
			},
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
			},
			truncate: -1,
			data: map[string]int{
				"foo": 1,
				"bar": 2,
			},
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
				{"foo", 2},
			},
			truncate: -1,
			data: map[string]int{
				"foo": 1,
				"bar": 2,
			},
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
				{"foo", 2},
			},
			truncate: 0,
			data:     map[string]int{},
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
				{"foo", 2},
			},
			truncate: 1,
			data: map[string]int{
				"foo": 1,
			},
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
				{"foo", 2},
			},
			truncate: 2,
			data: map[string]int{
				"foo": 1,
				"bar": 2,
			},
		},
		{
			input: []pair{
				{"foo", 1},
				{"foo", 2},
			},
			truncate: -1,
			data: map[string]int{
				"foo": 1,
			},
		},
		{
			input: []pair{
				{"foo", 1},
				{"foo", 2},
			},
			truncate: 0,
			data:     map[string]int{},
		},
		{
			input: []pair{
				{"foo", 1},
				{"foo", 2},
			},
			truncate: 1,
			data: map[string]int{
				"foo": 1,
			},
		},
	}

	for i, example := range examples {
		name := fmt.Sprintf("case %d", i)
		test := example

		t.Run(name, func(t *testing.T) {
			m := NewOrderedMap[string, int](0)

			for _, pair := range test.input {
				m.Put(pair.key, pair.value)
			}

			if test.truncate >= 0 {
				m.Truncate(test.truncate)
			}

			want, _ := json.Marshal(test.data)
			require.JSONEq(t, string(want), m.String())
		})
	}
}

func TestOrderedMap_Size(t *testing.T) {
	type pair struct {
		key   string
		value int
	}

	type example struct {
		input    []pair
		truncate int
		empty    bool
	}

	examples := []example{
		{
			input:    []pair{},
			truncate: -1,
			empty:    true,
		},
		{
			input:    []pair{},
			truncate: 0,
			empty:    true,
		},
		{
			input:    []pair{},
			truncate: 1,
			empty:    true,
		},
		{
			input: []pair{
				{"foo", 1},
			},
			truncate: -1,
			empty:    false,
		},
		{
			input: []pair{
				{"foo", 1},
			},
			truncate: 0,
			empty:    true,
		},
		{
			input: []pair{
				{"foo", 1},
			},
			truncate: 10,
			empty:    false,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
			},
			truncate: -1,
			empty:    false,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
				{"foo", 2},
			},
			truncate: -1,
			empty:    false,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
				{"foo", 2},
			},
			truncate: 0,
			empty:    true,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
				{"foo", 2},
			},
			truncate: 1,
			empty:    false,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
				{"foo", 2},
			},
			truncate: 2,
			empty:    false,
		},
		{
			input: []pair{
				{"foo", 1},
				{"foo", 2},
			},
			truncate: -1,
			empty:    false,
		},
		{
			input: []pair{
				{"foo", 1},
				{"foo", 2},
			},
			truncate: 0,
			empty:    true,
		},
		{
			input: []pair{
				{"foo", 1},
				{"foo", 2},
			},
			truncate: 1,
			empty:    false,
		},
	}

	for i, example := range examples {
		name := fmt.Sprintf("case %d", i)
		test := example

		t.Run(name, func(t *testing.T) {
			m := NewOrderedMap[string, int](0)

			for _, pair := range test.input {
				m.Put(pair.key, pair.value)
			}

			if test.truncate >= 0 {
				m.Truncate(test.truncate)
			}

			require.Equal(t, test.empty, m.Empty())
		})
	}
}

func TestOrderedMap_Empty(t *testing.T) {
	type pair struct {
		key   string
		value int
	}

	type example struct {
		input    []pair
		truncate int
		size     int
	}

	examples := []example{
		{
			input:    []pair{},
			truncate: -1,
			size:     0,
		},
		{
			input:    []pair{},
			truncate: 0,
			size:     0,
		},
		{
			input:    []pair{},
			truncate: 1,
			size:     0,
		},
		{
			input: []pair{
				{"foo", 1},
			},
			truncate: -1,
			size:     1,
		},
		{
			input: []pair{
				{"foo", 1},
			},
			truncate: 0,
			size:     0,
		},
		{
			input: []pair{
				{"foo", 1},
			},
			truncate: 10,
			size:     1,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
			},
			truncate: -1,
			size:     2,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
				{"foo", 2},
			},
			truncate: -1,
			size:     2,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
				{"foo", 2},
			},
			truncate: 0,
			size:     0,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
				{"foo", 2},
			},
			truncate: 1,
			size:     1,
		},
		{
			input: []pair{
				{"foo", 1},
				{"bar", 2},
				{"foo", 2},
			},
			truncate: 2,
			size:     2,
		},
		{
			input: []pair{
				{"foo", 1},
				{"foo", 2},
			},
			truncate: -1,
			size:     1,
		},
		{
			input: []pair{
				{"foo", 1},
				{"foo", 2},
			},
			truncate: 0,
			size:     0,
		},
		{
			input: []pair{
				{"foo", 1},
				{"foo", 2},
			},
			truncate: 1,
			size:     1,
		},
	}

	for i, example := range examples {
		name := fmt.Sprintf("case %d", i)
		test := example

		t.Run(name, func(t *testing.T) {
			m := NewOrderedMap[string, int](0)

			for _, pair := range test.input {
				m.Put(pair.key, pair.value)
			}

			if test.truncate >= 0 {
				m.Truncate(test.truncate)
			}

			require.Equal(t, test.size, m.Size())
		})
	}
}
