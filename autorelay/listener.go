package main

import (
	"context"
	"time"

	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-kad-dht"
	inet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-routing"
)

var listenLogger = logging.Logger("arListener")

func main() {
	logging.SetLogLevel("autorelay", "DEBUG")
	logging.SetLogLevel("relay", "DEBUG")
	logging.SetLogLevel("autonat", "DEBUG")

	ctx := context.Background()
	config, err := parseFlags()
	if err != nil {
		listenLogger.Fatal("couldn't parse flags:", err)
	}

	dhtFactory := func (h host.Host) (routing.PeerRouting, error) {
		return dht.New(ctx, h)
	}

	listenerHost, err := libp2p.New(ctx,
		libp2p.EnableRelay(),
		libp2p.Routing(dhtFactory),
		libp2p.EnableAutoRelay())
	if err != nil {
		listenLogger.Fatal("could not create host:", err)
	}
	listenLogger.Info("listener: ", listenerHost.ID(), listenerHost.ID().Pretty())

	// Now, to test things, let's set up a protocol handler on listenerHost
	listenerHost.SetStreamHandler("/cats", func(s inet.Stream) {
		listenLogger.Info("meow! It worked!")
		s.Close()
	})

	// discover and connect to the relay node
	dhtClient := dht.NewDHTClient(ctx, listenerHost, datastore.NewMapDatastore())
	relayPeerInfo, err := dhtClient.FindPeer(ctx, peer.ID(config.RelayID))
	if err != nil {
		listenLogger.Fatal("could not get relay node peer info:", err)
	}
	if err := listenerHost.Connect(ctx, relayPeerInfo); err != nil {
		listenLogger.Fatal("could not connect to relay node:", err)
	}

	// repeatedly advertise the rendezvous string through the DHT so our node can be found
	routingDiscovery := discovery.NewRoutingDiscovery(dhtClient)
	go func() {
		for {
			discovery.Advertise(ctx, routingDiscovery, config.RendezvousString)
			time.Sleep(time.Minute)
		}
	}()
}