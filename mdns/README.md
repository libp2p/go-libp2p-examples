## libp2p peer discovery with mDNS

[go-libp2p](https://github.com/libp2p/go-libp2p) currently implements the [mDNS](https://tools.ietf.org/html/rfc6762)
for zero configuration peer discovery in local networks. mDNS solves the
problem of resolving and maintaining a mapping table between peerIDs and local
ports on local networks. With mDNS, it is possible for peers to discover others
in the local network without knowing of their existence previously and without the 
need to rely on [pre-configured bootstrapping schemes](https://github.com/ipfs/ipfs/issues/30)
or any other previous configuration.

1) First, we create 2 libp2p hosts in the same local network, listening on all 
interfaces on `0.0.0.0` and in different ports;

```go
host1 := newHost(8000, ctx)
host2 := newHost(8001, ctx)

//...

func newHost(p int, ctx context.Context) host.Host {
  hma, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", p))
  h, err := libp2p.New(ctx, libp2p.ListenAddrs(hma))
  if err != nil {
      log.Fatal(err)
  }
  return h
}
```

2) Next we verify that the newly created hosts don't know about each other by
printing their peerstore.

3) Let's configure now the mDNS service in both hosts. The mDNS Service will 
multicast discovery requests at a certain interval of time while listening to 
discovery requests from other peers and reply to them. 

```go
// in order to find each other, the peers need to start a mDNS service which
// will query and handle nDNS responses
(https://tools.ietf.org/html/rfc6762)

h1discService, err := discovery.NewMdnsService(ctx, host1, time.Second,
"_host-discovery")
  if err != nil {
     log.Fatal(err)
  }
  defer h1discService.Close()

// host 2 also starts a mDNS service
h2discService, err := discovery.NewMdnsService(ctx, host2, time.Second,
"_host-discovery")
  if err != nil {
    log.Fatal(err)
  }
defer h2discService.Close()
```

We can inspect what happens on a package level when the mDNS service starts in 
both peers using a packet inspection tool:

![mDNS in action](http://www.giphy.com/gifs/1k2WcHlUpT7oCPPIaK)

First, each of the peers send a package with a `QM` query to the multicast address
224.0.0.251. Once the queries as received by each host, they reply with a `PTR`
response which contains the multihash of the host and in which port he is
listening to. Note that each host replies also to its own query. This is fine
and will the response is ignored by the go-libp2p mDNS service.

4) We need to let the mDNS service know what to do when receiving a discovery
reply from other peers. For the sake of this example, we'll set it only for one
peer, which will then be responsible to reach out for the new discovered peer.

```go
// when the mDNS discovery service receives a query response, it will call a
// an handler which must implement the Notifee interface. The Notifee
// interface expects `HandlePeerFound(pstore.PeerInfo)` to be implemented.
// let's initiate and register a mDNS response handler in host2, so that
when
// its mDNS Service receives a response, it will Connect to the peerID which
// responded. this way, both peers will register each other's multiaddress
in
// their peerstores.
h2handler := &rspHandler{host2}
h2discService.RegisterNotifee(h2handler)

func (rh *rspHandler) HandlePeerFound(pi pstore.PeerInfo) {
// Connect will add the host to the peerstore and dial up a new connection
fmt.Println(fmt.Sprintf("\nhost %v connecting to %v... (blocking)",
rh.host.ID(), pi.ID))

  err := rh.host.Connect(context.Background(), pi)
  if err != nil {
  fmt.Println(fmt.Sprintf("Error when connecting peers: %v", err))
    return
  }
  fmt.Println("dial up OK")
}
```

Once `host2` receives a mDNS discovery package, it will try to connect to it.
The `host.Connect` primitive will take care of adding the peers to the
respective peerstores.


5) Now, we can check again the peerstores of each peer to show that the hosts 
know about each other and that the peer discovery was successful.
