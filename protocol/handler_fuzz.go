// +build gofuzz

package protocol

import (
	"bytes"
	"encoding/hex"
	"io"
	"log"
	"strings"

	"github.com/GopherConRu/pb-fuzz-workshop/kv"
)

func Fuzz(data []byte) int {
	kv, err := kv.NewInMemoryBadgerKV()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := kv.Close(); err != nil {
			panic(err)
		}
	}()

	h := NewHandler(kv)
	input, output := h.NewConn()

	cmds := strings.Split(string(data), "\n")
	cmds = append(cmds, "PING")

	readDone := make(chan int)
	go func() {
		b, err := io.ReadAll(output)
		if err != nil {
			panic(err)
		}

		if !bytes.HasSuffix(b, []byte("+PONG\n")) {
			log.Printf("Input: %q", cmds)
			log.Printf("Output: %q", strings.Split(string(b), "\n"))
			panic("Raw output:\n" + hex.Dump(b))
		}

		if bytes.Contains(b, []byte("-ERR")) {
			readDone <- 0
			return
		}

		readDone <- 1
	}()

	for _, cmd := range cmds {
		if _, err := input.Write([]byte(cmd + "\n")); err != nil {
			panic(err)
		}
	}

	if err := input.Close(); err != nil {
		panic(err)
	}

	return <-readDone
}
