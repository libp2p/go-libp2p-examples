package main

import (
	"context"
	"fmt"
	"time"

	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-autonat-svc"
	circuit "github.com/libp2p/go-libp2p-circuit"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-kad-dht"
	inet "github.com/libp2p/go-libp2p-net"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-routing"
)

func main() {
	logging.SetLogLevel("autorelay", "DEBUG")
	logging.SetLogLevel("relay", "DEBUG")
	logging.SetLogLevel("autonat", "DEBUG")

	dhtFactory := func (h host.Host) (routing.PeerRouting, error) {
		return dht.New(context.Background(), h)
	}

	// Todo: Figure out how to run this example without needing to set up a whole network

	// Create three libp2p hosts, enable relay client capabilities on all
	// of them.

	// Have the sending host enable content routing.
	// Have the sending host enable autorelay
	senderHost, err := libp2p.New(context.Background(),
		libp2p.EnableRelay(),
		libp2p.Routing(dhtFactory),
		libp2p.EnableAutoRelay())
	if err != nil {
		panic(err)
	}
	fmt.Println("Sender: ", senderHost.ID(), senderHost.ID().Pretty())

	// Have the relay host enable content routing.
	// Have the relay host enable autorelay
	// Have the relay host relay connections for other peers (The ability to *use*
	// a relay vs the ability to *be* a relay)
	relayHost, err := libp2p.New(context.Background(),
		libp2p.EnableRelay(circuit.OptHop),
		libp2p.Routing(dhtFactory),
		libp2p.EnableAutoRelay(),
		libp2p.ListenAddrStrings("/dns4/lvh.me/tcp/0"))
	if err != nil {
		panic(err)
	}
	_, err = autonat.NewAutoNATService(context.Background(), relayHost)
	if err != nil {
		panic(err)
	}
	fmt.Println("Relay: ", relayHost.ID(), relayHost.ID().Pretty())


	// Have the listening host only enable relay communication
	// Zero out the listen addresses for the listening host, so it can only
	// communicate via p2p-circuit for our example
	listenerHost, err := libp2p.New(context.Background(), libp2p.EnableRelay())
	if err != nil {
		panic(err)
	}
	fmt.Println("Listener: ", listenerHost.ID(), listenerHost.ID().Pretty())


	relayHostInfo := pstore.PeerInfo{
		ID:    relayHost.ID(),
		Addrs: relayHost.Addrs(),
	}

	// Connect both senderHost and listenerHost to relayHost, but not to each other
	if err := senderHost.Connect(context.Background(), relayHostInfo); err != nil {
		panic(err)
	}
	if err := listenerHost.Connect(context.Background(), relayHostInfo); err != nil {
		panic(err)
	}

	// Now, to test things, let's set up a protocol handler on listenerHost
	listenerHost.SetStreamHandler("/cats", func(s inet.Stream) {
		fmt.Println("Meow! It worked!")
		s.Close()
	})

	// Creates a relay address
	//relayaddr, err := ma.NewMultiaddr("/p2p-circuit/ipfs/" + listenerHost.ID().Pretty())
	//if err != nil {
	//	panic(err)
	//}

	//listenerHostRelayInfo := pstore.PeerInfo{
	//	ID:    listenerHost.ID(),
	//	Addrs: []ma.Multiaddr{relayaddr},
	//}

	time.Sleep(time.Minute*2)
	fmt.Println(senderHost.Addrs())

	senderHostRelayInfo := pstore.PeerInfo{
		ID:    senderHost.ID(),
		Addrs: senderHost.Addrs(),
	}

	// Connect the sending host to the listening host via the relay
	//if err := senderHost.Connect(context.Background(), listenerHostRelayInfo); err != nil {
	//	panic(err)
	//}
	if err := listenerHost.Connect(context.Background(), senderHostRelayInfo); err != nil {
		panic(err)
	}

	// Woohoo! we're connected!
	s, err := senderHost.NewStream(context.Background(), listenerHost.ID(), "/cats")
	if err != nil {
		fmt.Println("huh, this should have worked: ", err)
		return
	}

	s.Read(make([]byte, 1)) // block until the handler closes the stream
}