package main

import (
	"context"

	"github.com/libp2p/go-libp2p"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO: add some libp2p.Transport options to this chain!
	transports := libp2p.ChainOptions()

	host, err := libp2p.New(ctx, transports)
	if err != nil {
		panic(err)
	}

	host.Close()
}
