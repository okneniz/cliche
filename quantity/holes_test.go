package quantity

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Get(t *testing.T) {
	t.Parallel()

	type example struct {
		span Interface
		skip []Interface
		want Interface
	}

	examples := []example{
		{
			span: Pair(0, 5),
			skip: []Interface{
				Pair(1, 4),
			},
			want: Pair(0, 5),
		},
		{
			span: Pair(0, 5),
			skip: []Interface{
				Pair(2, 5),
			},
			want: Pair(0, 1),
		},
		{
			span: Pair(0, 5),
			skip: []Interface{
				Pair(0, 1),
			},
			want: Pair(2, 5),
		},
	}

	for i := range examples {
		test := examples[i]
		name := fmt.Sprintf("case %d", i)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			t.Log("span", test.span)
			t.Log("skip", test.skip)

			actual := Get(
				test.span,
				newTestList(test.skip),
			)

			require.Equal(t, test.want.String(), actual.String())
		})
	}

}

type testList struct {
	data []Interface
}

func newTestList(data []Interface) *testList {
	return &testList{
		data: data,
	}
}

func (l *testList) Size() int {
	return len(l.data)
}

func (l *testList) At(idx int) (Interface, bool) {
	return l.data[idx], true
}
