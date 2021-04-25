package pb_fuzz_workshop

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func FuzzReverse(f *testing.F) {
	for _, s := range [][]int{
		nil,
		{},
		{1},
		{1, 2},
		{1, 2, 3},
	} {
		b, err := json.Marshal(s)
		require.NoError(f, err)
		f.Add(b)
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		t.Parallel()

		var s []int
		if err := json.Unmarshal(b, &s); err != nil {
			t.Skip(err)
		}

		require.Equal(t, s, Reverse(Reverse(s)))
		require.Equal(t, s, NotReverse(s))
	})
}
