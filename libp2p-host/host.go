package main

import (
	"context"
	"crypto/rand"
	"fmt"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	ma "github.com/multiformats/go-multiaddr"
)

func main() {
	// The context governs the lifetime of the libp2p node
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// To construct a simple host with all the default settings, just use `New`
	h, err := libp2p.New(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello World, my hosts ID is %s\n", h.ID())

	// If you want more control over the configuration, you can fill out fields
	// in the libp2p config, and use `NewWithCfg`
	cfg := new(libp2p.Config)

	// Set your own keypair
	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		panic(err)
	}
	cfg.PeerKey = priv

	// Set your own listen address
	maddr, err := ma.NewMultiaddr("/ip4/0.0.0.0/tcp/9000")
	if err != nil {
		panic(err)
	}

	// The config takes an array of addresses, specify as many as you want.
	cfg.ListenAddrs = []ma.Multiaddr{maddr}

	h2, err := libp2p.NewWithCfg(ctx, cfg)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello World, my second hosts ID is %s\n", h2.ID())
}
