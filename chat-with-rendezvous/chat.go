package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multihash"
)

// IPFS bootstrap nodes. Used to find other peers in the network.
var bootstrapPeers = []string{
	"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
	"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
	"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
	"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
}

var rendezvous = "meet me here"

func handleStream(stream net.Stream) {
	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	go writeData(rw)

	// stream 's' will stay open until you close it (or the other side closes it).
}
func readData(rw *bufio.ReadWriter) {
	for {
		str, _ := rw.ReadString('\n')

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')

		if err != nil {
			panic(err)
		}

		rw.WriteString(fmt.Sprintf("%s\n", sendData))
		rw.Flush()
	}
}

func main() {

	help := flag.Bool("h", false, "Display Help")
	rendezvousString := flag.String("r", rendezvous, "Unique string to identify group of nodes. Share this with your friends to let them connect with you")
	flag.Parse()

	if *help {
		fmt.Printf("This program demonstrates a simple p2p chat application using libp2p\n\n")
		fmt.Printf("Usage: Run './chat in two different terminals. Let them connect to the bootstrap nodes, announce themselves and connect to the peers\n")

		os.Exit(0)
	}

	ctx := context.Background()

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	host, err := libp2p.New(ctx)
	if err != nil {
		panic(err)
	}

	// Set a function as stream handler.
	// This function  is called when a peer initiate a connection and starts a stream with this peer.
	host.SetStreamHandler("/chat/1.1.0", handleStream)

	dht, err := dht.New(ctx, host)
	if err != nil {
		panic(err)
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the other nodes in the network.
	for _, addr := range bootstrapPeers {

		iaddr, _ := ipfsaddr.ParseString(addr)

		peerinfo, _ := peerstore.InfoFromP2pAddr(iaddr.Multiaddr())

		if err := host.Connect(ctx, *peerinfo); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Connection established with bootstrap node: ", *peerinfo)
		}
	}

	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	rendezvousPoint, _ := cid.NewPrefixV1(cid.Raw, multihash.SHA2_256).Sum([]byte(*rendezvousString))

	fmt.Println("announcing ourselves...")
	tctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if err := dht.Provide(tctx, rendezvousPoint, true); err != nil {
		panic(err)
	}

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	// 'FindProviders' will return 'PeerInfo' of all the peers which
	// have 'Provide' or announced themselves previously.
	fmt.Println("searching for other peers...")
	tctx, cancel = context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	peers, err := dht.FindProviders(tctx, rendezvousPoint)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d peers!\n", len(peers))

	for _, p := range peers {
		fmt.Println("Peer: ", p)
	}

	for _, p := range peers {

		if p.ID == host.ID() || len(p.Addrs) == 0 {
			// No sense connecting to ourselves or if addrs are not available
			continue
		}

		stream, err := host.NewStream(ctx, p.ID, "/chat/1.1.0")

		if err != nil {
			fmt.Println(err)
		} else {
			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

			go writeData(rw)
			go readData(rw)
		}

		fmt.Println("Connected to: ", p)
	}

	select {}
}
