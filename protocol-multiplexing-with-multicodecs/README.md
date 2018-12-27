# Protocol Multiplexing using multicodecs with libp2p

This example shows how to use multicodecs (i.e. json) to encode and transmit information between libp2p hosts using libp2p Streams.

Multicodecs present a common interface, making it very easy to swap the codec implementation if needed.

This example expects that you area already familiar with the [echo example](https://github.com/libp2p/go-libp2p-examples/tree/master/echo).

## Build

From the `go-libp2p-examples` directory run the following:

```
> make deps
> cd protocol-multiplexing-with-multicodecs/
> go build -o multicodecs
```

## Usage

```
> ./multicodecs
```

## Details

The example creates two libp2p Hosts. Host1 opens a stream to Host2. Host2 has a `StreamHandler` to deal with the incoming stream. This is covered in the `echo` example.

Both hosts simulate a conversation. But rather than sending raw messages on the stream, each message in the conversation is encoded under a `json` object (using the `json` multicodec). For example:

```
{
  "Msg": "This is the message",
  "Index": 3,
  "HangUp": false
}
```

The stream lasts until one of the sides closes it when the HangUp field is `true`.
