package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/multiformats/go-multiaddr"
	"log"
	"os"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	libp2pdht "github.com/libp2p/go-libp2p-kad-dht"
	inet "github.com/libp2p/go-libp2p-net"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multihash"
)

func handleStream(stream inet.Stream) {
	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	go writeData(rw)

	// 'stream' will stay open until you close it (or the other side closes it).
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer")
			panic(err)
		}

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
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}

// Useful only when we're the first node to launch in a peer network. Will
// attempt to announce to the peer network that we're subscribed (and producing)
// on the chat topic. This is necessary because, as the first peer, we have no
// other peers to announce to when we start up. For peers joining an existing
// network, this will succeed in its first iteration.
func provide(dht *libp2pdht.IpfsDHT, rendezvousPoint cid.Cid) {
	fmt.Println("Announcing ourselves...")
	for {
		timeoutCtx, cancel := context.WithTimeout(dht.Context(), 10*time.Second)
		defer cancel()
		if err := dht.Provide(timeoutCtx, rendezvousPoint, true); err == nil {
			fmt.Println("Successfully announced!")
			break
		}
		time.Sleep(5 * time.Second)
	}
}

func main() {
	help := flag.Bool("h", false, "Display Help")
	config, err := ParseFlags()
	if err != nil {
		panic(err)
	}

	if *help {
		fmt.Printf("This program demonstrates a simple p2p chat application using libp2p\n\n")
		fmt.Printf("Usage: Run './chat in two different terminals. Let them connect to the bootstrap nodes, announce themselves and connect to the peers\n")

		os.Exit(0)
	}

	ctx := context.Background()

	// Configure p2p host
	privkey, _, err := crypto.GenerateKeyPair(crypto.RSA, 1024)
	if err != nil {
		panic(err)
	}
	transports := libp2p.DefaultTransports
	addrs := make([]multiaddr.Multiaddr, len(config.ListenAddresses))
	copy(addrs, config.ListenAddresses)

	options := []libp2p.Option{transports, libp2p.DefaultMuxers}
	options = append(options, libp2p.ListenAddrs(addrs...))
	options = append(options, libp2p.Identity(privkey))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	host, err := libp2p.New(ctx, options...)
	if err != nil {
		panic(err)
	}

	// Set a function as stream handler.
	// This function is called when a peer initiates a connection and starts a stream with this peer.
	host.SetStreamHandler(protocol.ID(config.ProtocolID), handleStream)

	// Start a DHT, for use in peer discovery. We can't just make a new DHT client
	// because we want each peer to maintain its own local copy of the DHT, so
	// that the bootstrapping node of the DHT can go down without inhibitting
	// future peer discovery.
	kademliaDHT, err := libp2pdht.New(ctx, host)
	if err != nil {
		panic(err)
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	fmt.Println("Bootstrapping the DHT")
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the other nodes in the network.
	for _, peerAddr := range config.BootstrapPeers {
		peerinfo, _ := peerstore.InfoFromP2pAddr(peerAddr)

		if err := host.Connect(ctx, *peerinfo); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Connection established with bootstrap node: ", *peerinfo)
		}
	}

	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	v1Builder := cid.V1Builder{Codec: cid.Raw, MhType: multihash.SHA2_256}
	rendezvousPoint, _ := v1Builder.Sum([]byte(config.RendezvousString))
	go provide(kademliaDHT, rendezvousPoint)

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	// 'FindProviders' will return 'PeerInfo' of all the peers which
	// have 'Provide' or announced themselves previously.
	fmt.Println("Searching for other peers...")
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	peers, err := kademliaDHT.FindProviders(timeoutCtx, rendezvousPoint)
	if err != nil {
		cancel()
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

		fmt.Println("Connecting to: ", p)
		stream, err := host.NewStream(ctx, p.ID, protocol.ID(config.ProtocolID))

		if err != nil {
			fmt.Println("Connection failed: ", err)
			continue
		} else {
			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

			go writeData(rw)
			go readData(rw)
		}

		fmt.Println("Connected to: ", p)
	}

	select {}
}
