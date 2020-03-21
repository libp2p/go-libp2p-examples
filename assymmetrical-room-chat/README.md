# p2p assymetrical room chat app with libp2p [support peer discovery using rendezvous]

This program is a similar to the `room-chat-with-rendezvous` but with asymmetrical peerstore as described bellow

## How to build this example?

```
cd assymmetrical-room-chat/
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

### *difference with room-chat-with-rendezvous*

This example uses a custom [peerstore](https://github.com/jolatechno/mpi-peerstore).

This time, the host only listen to peer that connected to it (using the `handleStream` function), and only write to peer it discovered.

This remoove the need for the host to send its address first into a new stream.

## Authors
1. Joseph Touzet
