# `go-libp2p` examples and tutorials

[![](https://img.shields.io/badge/made%20by-Protocol%20Labs-blue.svg?style=flat-square)](https://protocol.ai)
[![](https://img.shields.io/badge/project-libp2p-yellow.svg?style=flat-square)](https://libp2p.io/)
[![](https://img.shields.io/badge/freenode-%23libp2p-yellow.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23libp2p)
[![Discourse posts](https://img.shields.io/discourse/https/discuss.libp2p.io/posts.svg)](https://discuss.libp2p.io)

In this folder, you can find a variety of examples to help you get started in using go-libp2p. Every example as a specific purpose and some of each incorporate a full tutorial that you can follow through, helping you expand your knowledge about libp2p and p2p networks in general.

Let us know if you find any issue or if you want to contribute and add a new tutorial, feel welcome to submit a pr, thank you!

## Examples and Tutorials

- [P2P chat room with rendezvous peer discovery](./room-chat-with-rendezvous)
- [P2P chat room with peer discovery using mdns](./room-chat-with-mdns)

For js-libp2p examples, check https://github.com/libp2p/js-libp2p/tree/master/examples

## Troubleshooting

When building the examples ensure you have a clean `$GOPATH`. If you have checked out and built other `libp2p` repos then you may get errors similar to the one below when building the examples. Note that the use of the `gx` package manager **is not required** to run the examples or to use `libp2p`.
```
$:~/go/src/github.com/libp2p/go-libp2p-examples/libp2p-host$ go build host.go
# command-line-arguments
./host.go:36:18: cannot use priv (type "github.com/libp2p/go-libp2p-crypto".PrivKey) as type "gx/ipfs/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto".PrivKey in argument to libp2p.Identity:
        "github.com/libp2p/go-libp2p-crypto".PrivKey does not implement "gx/ipfs/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto".PrivKey (wrong type for Equals method)
                have Equals("github.com/libp2p/go-libp2p-crypto".Key) bool
                want Equals("gx/ipfs/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto".Key) bool
```

To obtain a clean `$GOPATH` execute the following:
```
> mkdir /tmp/libp2p-examples
> export GOPATH=/tmp/libp2p/examples
```

---

The last gx published version of this module was:
