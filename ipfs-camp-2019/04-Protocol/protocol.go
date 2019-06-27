package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/libp2p/go-libp2p-core/network"
)

const protocol = "/libp2p/chat/1.0.0"

func chatHandler(s network.Stream) {
	var (
		err   error
		chunk []byte
		line  string
	)
	r := bufio.NewReader(s)
	for {
		prefix := true
		for prefix {
			chunk, prefix, err = r.ReadLine()
			switch err {
			case io.EOF:
				return
			default:
				fmt.Fprintln(os.Stderr, err)
			}
			line += string(chunk)
		}
		fmt.Println(line)
		line = ""
	}
}

func send(msg string, s network.Stream) error {
	w := bufio.NewWriter(s)
	n, err := w.WriteString(msg)
	if n != len(msg) {
		return fmt.Errorf("expected to write %d bytes, wrote %d", len(msg), n)
	}
	return err
}
