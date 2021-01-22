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

	// Tell the host to be a Relay client i.e. the ability to "use a relay".
	h1, err := libp2p.New(context.Background(), libp2p.EnableRelay())
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
	// via p2p-circuit for our example.
	// In this example, h1 will connect to h3 via the Relay Server h2.
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

	// Creates a relay address for h3 that will dial relay server h2 and ask h2 for a connection to h3.
	var relayAddrs []ma.Multiaddr
	for _, a := range h2.Addrs() {
		relayAddr, err := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s/p2p-circuit", h2.ID().Pretty()))
		if err != nil {
			panic(err)
		}
		relayAddrs = append(relayAddrs, a.Encapsulate(relayAddr))
	}

	// Since we just tried and failed to dial, the dialer system will, by default
	// prevent us from redialing again so quickly. Since we know what we're doing, we
	// can use this ugly hack (it's on our TODO list to make it a little cleaner)
	// to tell the dialer "no, its okay, let's try this again"
	h1.Network().(*swarm.Swarm).Backoff().Clear(h3.ID())

	h3relayInfo := peer.AddrInfo{
		ID:    h3.ID(),
		Addrs: relayAddrs,
	}
	if err := h1.Connect(context.Background(), h3relayInfo); err != nil {
		panic(err)
	}

	// Woohoo! we're connected to h3 via a connection Relayed by h2.
	s, err := h1.NewStream(context.Background(), h3.ID(), "/cats")
	if err != nil {
		fmt.Println("huh, this should have worked: ", err)
		return
	}

	s.Read(make([]byte, 1)) // block until the handler closes the stream
}
