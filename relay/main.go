package main

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"

	circuit "github.com/libp2p/go-libp2p-circuit"
	swarm "github.com/libp2p/go-libp2p-swarm"
	ma "github.com/multiformats/go-multiaddr"
)

func main() {
	// Create three libp2p hosts, enable relay client capabilities on all
	// of them.

	// Tell the host to monitor for relays.
	h1, err := libp2p.New(context.Background(), libp2p.EnableRelay(circuit.OptDiscovery))
	if err != nil {
		panic(err)
	}

	// Tell the host to relay connections for other peers (The ability to *use*
	// a relay vs the ability to *be* a relay)
	h2, err := libp2p.New(context.Background(), libp2p.EnableRelay(circuit.OptHop))
	if err != nil {
		panic(err)
	}

	// Zero out the listen addresses for the host, so it can only communicate
	// via p2p-circuit for our example
	h3, err := libp2p.New(context.Background(), libp2p.ListenAddrs(), libp2p.EnableRelay())
	if err != nil {
		panic(err)
	}

	h2info := peer.AddrInfo{
		ID:    h2.ID(),
		Addrs: h2.Addrs(),
	}

	// Connect both h1 and h3 to h2, but not to each other
	if err := h1.Connect(context.Background(), h2info); err != nil {
		panic(err)
	}
	if err := h3.Connect(context.Background(), h2info); err != nil {
		panic(err)
	}

	// Now, to test things, let's set up a protocol handler on h3
	h3.SetStreamHandler("/cats", func(s network.Stream) {
		fmt.Println("Meow! It worked!")
		s.Close()
	})

	_, err = h1.NewStream(context.Background(), h3.ID(), "/cats")
	if err == nil {
		fmt.Println("Didnt actually expect to get a stream here. What happened?")
		return
	}
	fmt.Println("Okay, no connection from h1 to h3: ", err)
	fmt.Println("Just as we suspected")

	// Creates a relay address
	relayaddr, err := ma.NewMultiaddr("/p2p-circuit/ipfs/" + h3.ID().Pretty())
	if err != nil {
		panic(err)
	}

	// Since we just tried and failed to dial, the dialer system will, by default
	// prevent us from redialing again so quickly. Since we know what we're doing, we
	// can use this ugly hack (it's on our TODO list to make it a little cleaner)
	// to tell the dialer "no, its okay, let's try this again"
	h1.Network().(*swarm.Swarm).Backoff().Clear(h3.ID())

	h3relayInfo := peer.AddrInfo{
		ID:    h3.ID(),
		Addrs: []ma.Multiaddr{relayaddr},
	}
	if err := h1.Connect(context.Background(), h3relayInfo); err != nil {
		panic(err)
	}

	// Woohoo! we're connected!
	s, err := h1.NewStream(context.Background(), h3.ID(), "/cats")
	if err != nil {
		fmt.Println("huh, this should have worked: ", err)
		return
	}

	s.Read(make([]byte, 1)) // block until the handler closes the stream
}
