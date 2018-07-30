package main

import (
	"context"
	"fmt"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	discovery "github.com/libp2p/go-libp2p/p2p/discovery"
	ma "github.com/multiformats/go-multiaddr"
	"log"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// creates host h1 and h2 listening on 0.0.0.0 on any interface device on port
	// 8000 and 8001 respectively
	host1 := newHost(8000, ctx)
	host2 := newHost(8001, ctx)

	// created host only know about themselves. their peerstore only keeps a
	// record of their own multiaddress
	printKnownPeers(host1)
	printKnownPeers(host2)

	// in order to find each other, the peers need to start a mDNS service which
	// will query and handle nDNS responses (https://tools.ietf.org/html/rfc6762)
	h1discService, err := discovery.NewMdnsService(ctx, host1, time.Second, "_host-discovery")
	if err != nil {
		log.Fatal(err)
	}
	defer h1discService.Close()

	// host 2 also starts a mDNS service
	h2discService, err := discovery.NewMdnsService(ctx, host2, time.Second, "_host-discovery")
	if err != nil {
		log.Fatal(err)
	}
	defer h2discService.Close()

	// when the mDNS discovery service receives a query response, it will call a
	// an handler which must implement the Notifee interface. The Notifee
	// interface expects `HandlePeerFound(pstore.PeerInfo)` to be implemented.
	// let's initiate and register a mDNS response handler in host2, so that when
	// its mDNS Service receives a response, it will Connect to the peerID which
	// responded. this way, both peers will register each other's multiaddress in
	// their peerstores.
	h2handler := &rspHandler{host2}
	h2discService.RegisterNotifee(h2handler)

	// let's wait a couple of seconds for the UDP packets to be exchanged
	time.Sleep(time.Second * 2)

	// now both peers know each other
	printKnownPeers(host1)
	printKnownPeers(host2)
}

func newHost(p int, ctx context.Context) host.Host {
	hma, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", p))
	h, err := libp2p.New(ctx, libp2p.ListenAddrs(hma))
	if err != nil {
		log.Fatal(err)
	}
	return h
}

type rspHandler struct {
	host host.Host
}

func (rh *rspHandler) HandlePeerFound(pi pstore.PeerInfo) {
	// Connect will add the host to the peerstore and dial up a new connection
	fmt.Println(fmt.Sprintf("\nhost %v connecting to %v... (blocking)", rh.host.ID(), pi.ID))

	err := rh.host.Connect(context.Background(), pi)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error when connecting peers: %v", err))
		return
	}
	fmt.Println("dial up OK")
}

func printKnownPeers(h host.Host) {
	fmt.Println(fmt.Sprintf("\nhost %v knows:", h.ID().Pretty()))
	for _, p := range h.Peerstore().Peers() {
		fmt.Println(fmt.Sprintf(" >> %v", p.Pretty()))
	}
}
