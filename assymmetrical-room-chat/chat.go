package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"sync"
	"os"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p-core/host"

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

func discoveryHandler(ctx context.Context, config Config, host host.Host) func(*peerstore.Peerstore,peer.ID){
	return func (p *peerstore.Peerstore, id peer.ID){
		Protocol := protocol.ID(config.RendezvousString + "//" + config.ProtocolID)
		stream, err := host.NewStream(ctx, id, Protocol)

		if err != nil {
			if !config.quiet {
				logger.Warning("Connection failed:", err)
			}
			return
		}

		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		w := writeData(rw)
		p.Add(peer.IDB58Encode(id), &w)

		logger.Info("Connected to:", id)
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

	p := peerstore.NewPeerstore(&host, discovery.NewRoutingDiscovery(kademliaDHT), config.RendezvousString)

	p.SetHostId()
	p.Annonce(ctx)
	p.Listen(ctx, discoveryHandler(ctx, config, host))
	p.SetStreamHandler(protocol.ID(config.ProtocolID), handleStream(p))

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

	select {}
}
