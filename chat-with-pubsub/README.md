# p2p chat app with libp2p [with pubsub]

- This program demonstrates a simple p2p chat application. 
- It can work among at least 2 peers if the these peers are in the same network.
- Peers all subscribe to a topic `chat-with-pubsub` so that peers can chat.
- This example is evolved from the [ipfs-camp-2019 example](https://github.com/libp2p/go-libp2p-examples/tree/master/ipfs-camp-2019/08-End) and [chat example](https://github.com/libp2p/go-libp2p-examples/tree/master/chat). Credit to authors of these two examples!

## Build

From the `go-libp2p-examples` directory run the following:

```
> make deps
> cd chat-with-pubsub
> go build
```

## Usage

### On node 'A'

```
> ./chat-with-pubsub
Run ./chat-with-pubsub -d /ip4/127.0.0.1/tcp/52324/p2p/QmYeFvW9wkEmfqWGLREu6tFCHqnaH2G3MnPtYsjitWmzJV

> hello
<peer.ID Qm*tWmzJV>: hello

<peer.ID Qm*X1aTJo>: world

> /name juinc
<peer.ID Qm*tWmzJV> -> juinc

> hello from juinc
juinc: hello from juinc
```

### On node 'B'

```
> ./chat -d /ip4/127.0.0.1/tcp/52324/p2p/QmYeFvW9wkEmfqWGLREu6tFCHqnaH2G3MnPtYsjitWmzJV
Connected to QmYeFvW9wkEmfqWGLREu6tFCHqnaH2G3MnPtYsjitWmzJV

<peer.ID Qm*tWmzJV>: hello

> world
<peer.ID Qm*RiVbkZ>: world

<peer.ID Qm*tWmzJV> -> juinc
juinc: hello from juinc
```

## Author
[Juin Chiu](https://github.com/juinc)
