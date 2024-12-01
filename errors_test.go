package cliche

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_OutOfBounds(t *testing.T) {
	err := OutOfBounds{
		Min:   10,
		Max:   100,
		Value: -50,
	}

	require.Equal(t, err.Error(), "-50 is ouf of bounds 10..100")
}
