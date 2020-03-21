# p2p room chat app with libp2p [support peer discovery using rendezvous]

This program is a copy of the `chat-with-rendezvous` example with the ability to create a *room-chat* with more then two users.

## How to build this example?

```
cd room-chat-with-rendezvous/
go build -o chat
```

## Usage

Use different terminal windows to run

```
./chat -listen /ip4/127.0.0.1/tcp/6666
./chat -listen /ip4/127.0.0.1/tcp/6668
./chat -listen /ip4/127.0.0.1/tcp/6669

...

./chat -listen /ip4/127.0.0.1/tcp/$port
```

*note :* you can add __`-q`__ to launch in quiet mode.

## So how does it work?

### *difference with chat-with-rendezvous*

This example uses a custom [peerstore](https://github.com/jolatechno/mpi-peerstore).

On peer discovery, the first message sent is the host ID. Each peer check if they have a connection to the given peer and if not store the stream in a map where key are peer ID.

Upon sending a message, the message is sent to every stored stream.

When a peer disconnect, to ensure that other peer don't crash, channels are closed and removed when a peer doesn't respond.

## Authors
1. Joseph Touzet
