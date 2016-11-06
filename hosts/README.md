# Echo client/server with libp2p

This example can be started in either listen mode (server), or dial mode (client).

In listen mode, it will sit and wait for incoming connections on
the `/echo/1.0.0` protocol. Whenever it receives a stream, it will
write whatever message it received, and close the stream.

In dial mode, it will connect to the target peer on the given address.
It then opens a stream, writes a short message on the same protocol,
and print whatever reply it receives.

## Build

```
> make deps
> go build ./examples/hosts
```

## Usage

In one terminal:

```
> ./hosts -l 4737
2016/11/06 04:37:00 I am /ip4/127.0.0.1/tcp/4737/ipfs/QmXzbaXtBw6mU29WoeYrCtcRLVbT8asWCcEFVuDy4w6pdq
2016/11/06 04:37:00 listening for connections
2016/11/06 04:37:01 got a new stream
2016/11/06 04:37:01 read request: "Hello, world!"
```

In another, copy the address printed by the listener and do:

```
> ./hosts -d /ip4/127.0.0.1/tcp/4737/ipfs/QmXzbaXtBw6mU29WoeYrCtcRLVbT8asWCcEFVuDy4w6pdq
2016/11/06 04:37:01 I am /ip4/127.0.0.1/tcp/0/ipfs/QmeMNYMmkgoyd8M7y925r4yVVDjKtiYtU4rNCyj7wDWzk1
2016/11/06 04:37:01 connecting to target
2016/11/06 04:37:01 opening stream
2016/11/06 04:37:01 read reply: "Hello, world!"
>
```
