package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	ps "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

// helper method - create a lib-p2p host to listen on a port
func makeRandomNode(port int, done chan bool) *Node {
	// Ignoring most errors for brevity
	// See echo example for more details and better implementation
	priv, _, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	listen, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
	host, _ := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(listen),
		libp2p.Identity(priv),
	)

	return NewNode(host, done)
}

func main() {
	// Choose random ports between 10000-10100
	rand.Seed(666)
	port1 := rand.Intn(100) + 10000
	port2 := port1 + 1

	done := make(chan bool, 1)

	// Make 2 hosts
	h1 := makeRandomNode(port1, done)
	h2 := makeRandomNode(port2, done)
	h1.Peerstore().AddAddrs(h2.ID(), h2.Addrs(), ps.PermanentAddrTTL)
	h2.Peerstore().AddAddrs(h1.ID(), h1.Addrs(), ps.PermanentAddrTTL)

	log.Printf("This is a conversation between %s and %s\n", h1.ID().Pretty(), h2.ID().Pretty())

	// send messages using the protocols
	h1.Ping(h2.Host)
	h2.Ping(h1.Host)
	h1.Echo(h2.Host)
	h2.Echo(h1.Host)

	// block until all responses have been processed
	for i := 0; i < 4; i++ {
		<-done
	}
}
