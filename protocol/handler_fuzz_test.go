package protocol

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/GopherConRu/pb-fuzz-workshop/kv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzHandler(f *testing.F) {
	var testdata = []struct {
		in  string
		out []string
	}{
		{in: "", out: []string{"-ERR TODO1"}},
		{in: "\n", out: []string{"-ERR TODO1"}},
		{in: "\n\n", out: []string{"-ERR TODO1"}},
		{in: "FOO", out: []string{"-ERR TODO2"}},
		{in: "GET myKey", out: []string{"$-1"}},
		{in: "SET myKey 123", out: []string{"+OK"}},
		{in: "GET myKey", out: []string{"$3", "123"}},
		{in: "PING", out: []string{"+PONG"}},
	}

	kv, err := kv.NewInMemoryBadgerKV()
	require.NoError(f, err)
	f.Cleanup(func() {
		require.NoError(f, kv.Close())
	})

	h := NewHandler(kv)

	var all string
	for _, td := range testdata {
		input, output := h.NewConn()

		_, err := fmt.Fprint(input, td.in+"\n")
		require.NoError(f, err)

		for _, expected := range td.out {
			actual, err := output.ReadString('\n')
			require.NoError(f, err)
			assert.Equal(f, expected+"\n", actual)
		}

		require.NoError(f, input.Close())

		res, err := output.ReadString('\n')
		require.Equal(f, io.EOF, err)
		assert.Empty(f, res)

		all += td.in + "\n"
		f.Add(td.in)
		f.Add(all)
	}

	f.Fuzz(func(t *testing.T, data string) {
		t.Parallel()

		input, output := h.NewConn()

		cmds := strings.Split(data, "\n")
		cmds = append(cmds, "PING")

		readDone := make(chan struct{})
		go func() {
			defer close(readDone)

			b, err := io.ReadAll(output)
			assert.NoError(t, err)

			if !bytes.HasSuffix(b, []byte("+PONG\n")) {
				msg := fmt.Sprintf("Input: %q\n", cmds)
				msg += fmt.Sprintf("Output: %q", strings.Split(string(b), "\n"))
				msg += hex.Dump(b)
				assert.Fail(t, msg)
			}
		}()

		for _, cmd := range cmds {
			_, err := input.Write([]byte(cmd + "\n"))
			require.NoError(t, err)
		}

		require.NoError(t, input.Close())

		<-readDone
	})
}
