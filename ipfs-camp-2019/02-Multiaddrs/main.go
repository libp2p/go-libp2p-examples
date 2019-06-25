package main

import (
	"context"

	"github.com/libp2p/go-libp2p"
	tcp "github.com/libp2p/go-tcp-transport"
	ws "github.com/libp2p/go-ws-transport"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	transports := libp2p.ChainOptions(
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(ws.New),
	)

	host, err := libp2p.New(ctx, transports)
	if err != nil {
		panic(err)
	}

	host.Close()
}
