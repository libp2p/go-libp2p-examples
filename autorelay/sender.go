package main

import (
	"context"

	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-peer"
)

var sendLogger = logging.Logger("arSender")

func main() {
	logging.SetLogLevel("relay", "DEBUG")

	ctx := context.Background()
	config, err := parseFlags()
	if err != nil {
		sendLogger.Fatal("could not parse flags:", err)
	}

	// Have the sending host enable content routing.
	// Have the sending host enable autorelay
	senderHost, err := libp2p.New(ctx, libp2p.EnableRelay())
	if err != nil {
		sendLogger.Fatal("could not create the host:", err)
	}
	sendLogger.Info("Sender: ", senderHost.ID(), senderHost.ID().Pretty())

	// discover and connect to the relay node
	dhtClient := dht.NewDHTClient(ctx, senderHost, datastore.NewMapDatastore())
	relayPeerInfo, err := dhtClient.FindPeer(ctx, peer.ID(config.RelayID))
	if err != nil {
		sendLogger.Fatal("could not find the relay node:", err)
	}
	if err := senderHost.Connect(ctx, relayPeerInfo); err != nil {
		sendLogger.Fatal("could not connect to relay node:", err)
	}

	// find the NAT listening host
	routingDiscovery := discovery.NewRoutingDiscovery(dhtClient)
	peerChan, err := routingDiscovery.FindPeers(ctx, config.RendezvousString)
	if err != nil {
		sendLogger.Fatal("could not find peers", err)
	}

	for p := range peerChan {
		s, err := senderHost.NewStream(ctx, p.ID, "/cats")
		if err != nil {
			sendLogger.Warning("Connection failed:", err)
			continue
		} else {
			s.Read(make([]byte, 1))
		}
	}
}