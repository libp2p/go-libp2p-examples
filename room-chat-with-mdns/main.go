package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"

	"github.com/jolatechno/libp2p-peerstore"

	"github.com/multiformats/go-multiaddr"
)

func handleStream(p peerstore.Peerstore) func(network.Stream){
	return func(stream network.Stream) {
		// Create a buffer stream for non blocking read and write.
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		addr, err := rw.ReadString('\n')
		addr = strings.Replace(addr, "\n", "", -1)

		if err != nil {
			return //errors here shloud just disconnect the handler
		}

		if !p.Has(addr) {
			ID, _ := peer.IDHexDecode(addr)
			fmt.Println("received connection from:", ID)

			p.Add(addr, writeData(rw))
			go readData(rw)
		}
		// 'stream' will stay open until you close it (or the other side closes it).
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
	help := flag.Bool("help", false, "Display Help")
	cfg := parseFlags()

	if *help {
		fmt.Printf("Simple example for peer discovery using mDNS. mDNS is great when you have multiple peers in local LAN.")
		fmt.Printf("Usage: \n   Run './chat-with-mdns'\nor Run './chat-with-mdns -host [host] -port [port] -rendezvous [string] -pid [proto ID]'\n")

		os.Exit(0)
	}

	fmt.Printf("[*] Listening on: %s with port: %d\n", cfg.listenHost, cfg.listenPort)

	ctx := context.Background()
	r := rand.Reader

	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", cfg.listenHost, cfg.listenPort))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	host, err := libp2p.New(
		ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)

	if err != nil {
		panic(err)
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
	// Set a function as stream handler.
	// This function is called when a peer initiates a connection and starts a stream with this peer.
	host.SetStreamHandler(protocol.ID(cfg.ProtocolID), handleStream(p))

	fmt.Printf("\n[*] Your Multiaddress Is: /ip4/%s/tcp/%v/p2p/%s\n", cfg.listenHost, cfg.listenPort, host.ID().Pretty())

	peerChan := initMDNS(ctx, host, cfg.RendezvousString)

	for {
		Peer := <-peerChan // will block untill we discover a peer
		if !p.Has(peer.IDHexEncode(Peer.ID)) {
			if err := host.Connect(ctx, Peer); err != nil {
				fmt.Println("Connection failed:", err)
			}

			// open a stream, this stream will be handled by handleStream other end
			stream, err := host.NewStream(ctx, Peer.ID, protocol.ID(cfg.ProtocolID))

			if err != nil {
				fmt.Println("Stream open failed", err)
			} else {
				rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
				err := writeData(rw, peer.IDHexEncode(host.ID()))
				if err != nil {
					continue
				}

				p.Add(peer.IDHexEncode(Peer.ID), writeData(rw))
				go readData(rw)

				fmt.Println("Connected to:", Peer.ID)
			}
		}
	}
}
