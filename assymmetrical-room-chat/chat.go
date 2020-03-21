package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-discovery"

	"github.com/jolatechno/mpi-peerstore"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	multiaddr "github.com/multiformats/go-multiaddr"
	logging "github.com/whyrusleeping/go-logging"

	"github.com/ipfs/go-log"
)

var logger = log.Logger("rendezvous")

func handleStream(p peerstore.Peerstore) func(network.Stream){
	return func(stream network.Stream) {
		// Create a buffer stream for non blocking read and write.
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
		go readData(rw)

		logger.Info("Got a new stream !")
	}
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			return //errors here shloud just disconnect the reader
		}

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m", str)
		}

	}
}

func writeData(rw *bufio.ReadWriter) func(string) error{
	return func(str string) error{
		_, err := rw.WriteString(fmt.Sprintf("%s\n", str))
		if err != nil {
			return err
		}
		err = rw.Flush()
		if err != nil {
			return err
		}

		return nil
	}
}

func main() {
	log.SetAllLoggers(logging.WARNING)
	log.SetLogLevel("rendezvous", "info")
	help := flag.Bool("h", false, "Display Help")
	config, err := ParseFlags()
	if err != nil {
		panic(err)
	}

	if *help {
		fmt.Println("This program demonstrates a simple p2p chat application using libp2p")
		fmt.Println()
		fmt.Println("Usage: Run './chat in two different terminals. Let them connect to the bootstrap nodes, announce themselves and connect to the peers")
		flag.PrintDefaults()
		return
	}

	ctx := context.Background()

	// libp2p.New constructs a new libp2p Host. Other options can be added
	// here.
	host, err := libp2p.New(ctx,
		libp2p.ListenAddrs([]multiaddr.Multiaddr(config.ListenAddresses)...),
	)
	if err != nil {
		panic(err)
	}

	logger.Info("Host created. We are:", host.ID())
	if !config.quiet {
		logger.Info(host.Addrs())
	}

	p := peerstore.NewPeerstore(writeData)

	go func(){
		stdReader := bufio.NewReader(os.Stdin)

		for{
			str, err := stdReader.ReadString('\n')

			if err != nil {
				continue
			}

			p.WriteAll(str)
		}
	}()
	// Set a function as stream handler. This function is called when a peer
	// initiates a connection and starts a stream with this peer.
	host.SetStreamHandler(protocol.ID(config.ProtocolID), handleStream(p))

	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := dht.New(ctx, host)
	if err != nil {
		panic(err)
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	if !config.quiet {
		logger.Debug("Bootstrapping the DHT")
	}
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
			panic(err)
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	for _, peerAddr := range config.BootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := host.Connect(ctx, *peerinfo); err != nil {
				if !config.quiet {
					logger.Warning(err)
				}
			} else {
				if !config.quiet {
					logger.Info("Connection established with bootstrap node:", *peerinfo)
				}
			}
		}()
	}
	wg.Wait()

	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	if !config.quiet {
		logger.Info("Announcing ourselves...")
	}
	routingDiscovery := discovery.NewRoutingDiscovery(kademliaDHT)
	discovery.Advertise(ctx, routingDiscovery, config.RendezvousString)
	if !config.quiet {
		logger.Debug("Successfully announced!")

		// Now, look for others who have announced
		// This is like your friend telling you the location to meet you.
		logger.Debug("Searching for other peers...")
	}

	for {
		peerChan, err := routingDiscovery.FindPeers(ctx, config.RendezvousString)
		if err != nil {
			panic(err)
		}

		for Peer := range peerChan {
			if Peer.ID == host.ID() || p.Has(peer.IDHexEncode(Peer.ID)) {
				continue
			}
			if !config.quiet {
				logger.Debug("Found peer:", Peer)
				logger.Debug("Connecting to:", Peer)
			}
			stream, err := host.NewStream(ctx, Peer.ID, protocol.ID(config.ProtocolID))

			if err != nil {
				if !config.quiet {
					logger.Warning("Connection failed:", err)
				}
				continue
			}

			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
			p.Add(peer.IDHexEncode(Peer.ID), writeData(rw))

			logger.Info("Connected to:", Peer.ID)
		}
	}
}
