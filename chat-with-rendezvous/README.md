# p2p chat app with libp2p [support peer discovery]

This program demonstrates a simple p2p chat application. You will learn how to discover a peer in the network (using kad-dht), connect to it and open a chat stream. 

## How to build this example?

```
go get github.com/libp2p/go-libp2p-examples/chat-with-rendezvous

go build *.go
```

## Usage

Use two different terminal windows to run

```
./chat -listen /ip4/127.0.0.1/tcp/6666
./chat -listen /ip4/127.0.0.1/tcp/6668
```
## So how does it work?

1. **Configure a p2p host**
```go
ctx := context.Background()

// libp2p.New constructs a new libp2p Host.
// Other options can be added here.
host, err := libp2p.New(ctx)
```
[libp2p.New](https://godoc.org/github.com/libp2p/go-libp2p#New) is the constructor for libp2p node. It creates a host with given configuration. Right now, all the options are default, documented [here](https://godoc.org/github.com/libp2p/go-libp2p#New)

2. **Set a default handler function for incoming connections.**

This function is called on the local peer when a remote peer initiate a connection and starts a stream with the local peer.
```go
// Set a function as stream handler.
host.SetStreamHandler("/chat/1.1.0", handleStream)
```

```handleStream``` is executed for each new stream incoming to the local peer. ```stream``` is used to exchange data between local and remote peer. This example uses non blocking functions for reading and writing from this stream.

```go
func handleStream(stream net.Stream) {

    // Create a buffer stream for non blocking read and write.
    rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

    go readData(rw)
    go writeData(rw)

    // 'stream' will stay open until you close it (or the other side closes it).
}
```

3. **Initiate a new DHT Client with ```host``` as local peer.**


```go
dht, err := dht.New(ctx, host)
```

4. **Connect to IPFS bootstrap nodes.**

These nodes are used to find nearby peers using DHT.

```go
for _, addr := range bootstrapPeers {

    iaddr, _ := ipfsaddr.ParseString(addr)

    peerinfo, _ := peerstore.InfoFromP2pAddr(iaddr.Multiaddr())

    if err := host.Connect(ctx, *peerinfo); err != nil {
        fmt.Println(err)
    } else {
        fmt.Println("Connection established with bootstrap node: ", *peerinfo)
    }
}
```

5. **Announce your presence using a rendezvous point.**

[dht.Provide](https://godoc.org/github.com/libp2p/go-libp2p-kad-dht#IpfsDHT.Provide) makes this node announce that it can provide a value for the given key. Where a key in this case is ```rendezvousString```. Other peers will hit the same key to find other peers. 

```go
routingDiscovery := discovery.NewRoutingDiscovery(kademliaDHT)
discovery.Advertise(ctx, routingDiscovery, config.RendezvousString)
```

6. **Find peers nearby.**

[dht.FindProviders](https://godoc.org/github.com/libp2p/go-libp2p-kad-dht#IpfsDHT.FindProviders) will return all the peers who have announced their presence before.

```go
peerChan, err := routingDiscovery.FindPeers(ctx, config.RendezvousString)
```

**Note:** Although [dht.Provide](https://godoc.org/github.com/libp2p/go-libp2p-kad-dht#IpfsDHT.Provide) and [dht.FindProviders](https://godoc.org/github.com/libp2p/go-libp2p-kad-dht#IpfsDHT.FindProviders) works for a rendezvous peer discovery, this is not the right way of doing it. Libp2p is currently working on an actual rendezvous protocol ([libp2p/specs#56](https://github.com/libp2p/specs/pull/56)) which can be used for bootstrap purposes, real time peer discovery and application specific routing.

7. **Open streams to peers found.**

Finally we open stream to the peers we found, as we find them

```go
go func() {
		for peer := range peerChan {
			if peer.ID == host.ID() {
				continue
			}
			fmt.Println("Found peer:", peer)

			fmt.Println("Connecting to:", peer)
			stream, err := host.NewStream(ctx, peer.ID, protocol.ID(config.ProtocolID))

			if err != nil {
				fmt.Println("Connection failed:", err)
				continue
			} else {
				rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

				go writeData(rw)
				go readData(rw)
			}

			fmt.Println("Connected to:", peer)
		}
	}()
```

## Authors
1. Abhishek Upperwal
2. Mantas Vidutis