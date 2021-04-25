package pb_fuzz_workshop

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestReverseNaive(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// generate data
		s := rapid.SliceOf(rapid.Int()).Draw(t, "slice").([]int)

		// run prop.
		require.Equal(t, s, Reverse(Reverse(s)))
	})
}

func TestReverse(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		s := rapid.SliceOf(rapid.Int()).Draw(t, "slice").([]int)

		orig := make([]int, len(s))
		require.Equal(t, len(s), copy(orig, s))

		reversed := Reverse(s)

		for i := range orig {
			require.Equal(t, orig[i], reversed[len(orig)-1-i])
		}
	})
}

func TestReverseGen(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		s := rapid.SliceOf(rapid.Int()).Filter(func(_s []int) bool {
			return len(_s) > 0
		}).Draw(t, "slice").([]int)
		i := rapid.IntRange(0, len(s)-1).Draw(t, "pos").(int)

		orig := make([]int, len(s))
		require.Equal(t, len(s), copy(orig, s))

		reversed := Reverse(s)

		require.Equal(t, orig[i], reversed[len(orig)-1-i])
	})
}
