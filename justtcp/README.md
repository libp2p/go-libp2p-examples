# libp2p 'just tcp' example

## What this does
This example starts up a libp2p swarm that listens for tcp connections behind a
multistream muxer protocol of `/plaintext/1.0.0`. All connections made to it
will be echoed back.

## Building
```
$ go build
```

## Usage
```
$ ./justtcp
```
