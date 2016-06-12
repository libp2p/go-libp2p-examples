# libp2p host example

## What this does
This example can be started in either listen mode, or dial mode.
When running as a listener, it will sit and wait for incoming connections on
the `/hello/1.0.0` protocol. Whenever it receives a stream, it will 
write the message "hello libp2p" over the stream and close it.
When run in dial mode, the node will start up, connect to the given
address, open a stream to the target peer, and read a message on the
protocol `/hello/1.0.0`

## Building
```
$ go build
```

## Usage
In one terminal:
```
$ ./hosts -l 4737
```

In another, copy the address printed by the listener and do:

```
$ ./hosts -d ADDRESS
```

