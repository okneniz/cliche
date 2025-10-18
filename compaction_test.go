package cliche

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTreeCompaction(t *testing.T) {
	t.Parallel()

	examples := [][]string{
		{
			"[a-z1-2]",
			"[1-2a-z]",
			"[12a-z]",
			"[1a-z2]",
			"[1-2[a-z]]",
			"[[1-2][a-z]]",
			"[12[a-z]]",
			"[12a[b-z]]",
		},
		{
			"[a-z]",
			"[a-cd-z]",
		},
		{
			"[abc]",
			"[cab]",
			"[bac]",
			"[bca]",
		},
		{
			"[a]",
			"[aaaa]",
			"[a-a]",
			"[a-aa]",
			"[a-a[a]]",
			"[[a][a-a]]",
			"[[a-a][a-a]]",
		},
		{
			"[0-9]",
			"[0123456789]",
			"[0123-9]",
			"[0-34-9]",
		},
		{
			"x+",
			"x{1,}",
		},
		{
			"x?",
			"x{0,1}",
			"x{,1}",
		},
		{
			"x*",
			"x{0,}",
		},
		{
			"x{1}",
			"x{1,1}",
		}, {
			"x",
			"(?#123)x",
		},
	}

	for _, expressions := range examples {
		example := expressions

		name := fmt.Sprintf(
			"compact expressions to one : %s",
			strings.Join(example, ", "),
		)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tr := New(DefaultParser)
			require.Equal(t, tr.Size(), 0)

			err := tr.Add(example...)
			require.NoError(t, err)

			t.Log(tr.String())

			require.Equal(t, tr.Size(), 1)
		})
	}
}
