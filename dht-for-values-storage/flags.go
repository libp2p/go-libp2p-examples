package main

import (
	"flag"
)

type config struct {
	topic      string
	ProtocolID string
	listenHost string
	listenPort int
}

func parseFlags() *config {
	c := &config{}

	flag.StringVar(&c.topic, "topic", "/libp2p/example/chat/1.0.0", "Unique string to identify group of nodes")
	flag.StringVar(&c.listenHost, "host", "0.0.0.0", "The bootstrap node host listen address\n")
	flag.StringVar(&c.ProtocolID, "pid", "/chat/1.1.0", "Sets a protocol id for stream headers")
	flag.IntVar(&c.listenPort, "port", 4001, "node listen port")

	flag.Parse()
	return c
}
