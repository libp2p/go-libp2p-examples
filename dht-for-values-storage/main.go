package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	levelds "github.com/ipfs/go-ds-leveldb"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-core/routing"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	mplex "github.com/libp2p/go-libp2p-mplex"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	yamux "github.com/libp2p/go-libp2p-yamux"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	tcp "github.com/libp2p/go-tcp-transport"
	ws "github.com/libp2p/go-ws-transport"
)

type mdnsNotifee struct {
	h   host.Host
	ctx context.Context
}

func (m *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	m.h.Connect(m.ctx, pi)
}

func main() {

	// Exection flags
	help := flag.Bool("h", false, "Display Help")
	config := parseFlags()

	if *help {
		fmt.Println("This programs demonstrates the use of a pubsub pattern to broadcast values in the network. Get them stored into some peer (not always the one adding them) and being able to retrieve them from any node by have the CID of the value wanted.")
		fmt.Println()
		fmt.Println("Usage: Run './start")
		flag.PrintDefaults()
		return
	}

	// Starting a new context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//  ListenAddrStrings configures libp2p to listen on the given (unparsed) addresses.
	//listenAddrs := libp2p.ListenAddrStrings("/ip4/" + config.listenHost + "/tcp/" + strconv.Itoa(config.listenPort))
	listenAddrs := libp2p.ListenAddrStrings(
		"/ip4/0.0.0.0/tcp/0",
		"/ip4/0.0.0.0/tcp/0/ws",
	)

	// Package dht implements a distributed hash table that satisfies the ipfs routing interface. This DHT is modeled after kademlia with S/Kademlia modifications.
	var dht *kaddht.IpfsDHT
	newDHT := func(h host.Host) (routing.PeerRouting, error) {
		var err error
		ds, err := levelds.NewDatastore("", nil)
		if err != nil {
			fmt.Println(err)
		}

		dht = kaddht.NewDHT(ctx, h, ds)
		return dht, err
	}

	routing := libp2p.Routing(newDHT)

	// Choosing transport protocol
	transports := libp2p.ChainOptions(
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(ws.New),
	)

	// Choosing muxers for multiple streams
	muxers := libp2p.ChainOptions(
		libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
		libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
	)

	// Creating the host with the previous parameters
	host, err := libp2p.New(
		ctx,
		transports,
		listenAddrs,
		muxers,
		routing,
	)
	if err != nil {
		panic(err)
	}

	// Setting the function 'handleStream' to handle incoming connections
	host.SetStreamHandler(protocol.ID("tcp"), handleStream)

	fmt.Printf("\n[*] Topic: %s\n\n", config.topic)
	multiAddress := "/ip4/" + config.listenHost + "/tcp/" + strconv.Itoa(config.listenPort) + "/p2p/" + host.ID().Pretty()
	fmt.Printf("[*] Your MultiAddrress Is: %s\n\n", multiAddress)
	fmt.Printf("[*] Your DHT ID is : %s\n\n", dht.PeerID())

	// New gossip pubsub for the host
	ps, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		panic(err)
	}

	// subscribing to the topic
	sub, err := ps.Subscribe(pubsubTopic)
	if err != nil {
		panic(err)
	}
	go pubsubHandler(ctx, sub)

	// Printing the listening interfaces
	for _, addr := range host.Addrs() {
		fmt.Println("-> Listening on", addr)
	}

	fmt.Println("\n[*] How to:\nUse \x1b[32m/peers\x1b[0m to see the peers list\nUse \x1b[32m/name\x1b[0m to set your peer name for message delivery\nUse \x1b[32m/put\x1b[0m to put a value in the DHT\nUse \x1b[32m/get\x1b[0m to get Value from the DHT\n\n")

	// Use mdns for discovery
	mdns, err := discovery.NewMdnsService(ctx, host, time.Second*10, pubsubTopic)
	if err != nil {
		panic(err)
	}
	mdns.RegisterNotifee(&mdnsNotifee{h: host, ctx: ctx})

	err = dht.Bootstrap(ctx)
	if err != nil {
		panic(err)
	}

	// Setting some global variables used in the code
	setter(ctx, dht, ps, multiAddress)
	donec := make(chan struct{}, 1)
	go chatInputLoop(ctx, host, ps, donec, dht)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT)

	select {
	case <-stop:
		host.Close()
		os.Exit(0)
	case <-donec:
		host.Close()
	}
}
