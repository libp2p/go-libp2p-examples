# p2p Chat and storing values [using PubSub and DHT]
This programs demonstrates the use of a pubsub pattern to broadcast values in the network. Get them stored into some peer (not always the one adding them) and being able to retrieve them from any node by having the CID of the value wanted.

This example is based on the **ipfs-camp-2019 pubsub code**.

## How to build
 > go get -v -d ./...
 
 > go build .
 
 # Code explanation
 The functions added to the **ipfs-camp-2019** last code behave in the following way:
 
 
- /get cid:
    - getProviderrInfo (dht, ctx, cid)  peer.AddrInfo
    - getValue (dht, ctx, "cid:addr") : 
        - getProviderInfo 
        - Connect To it 
        - GetWriter 

- /put value: 
    - addValue(dht, ctx, value)
        - FindClosestPeerInfo(dht, ctx, cid) peer.AddrInfo
        - pushValue (dht, ctx, cid, value) :
            - FindClosestPeer 
            - Connect to it 
            - PutWriter 

- Handle Stream:
    - Reading/Writing into the stream
        - ReadData(stream) :
            - If it is “cid:value”:
                - pushToStore(cid:value)
                - provide it to the network
            - If it is “get:cid:addr”:
                - getReader(stream,cid+addr)
            - If it is “value”
                - PrintItToScreen
        - WriteData(stream, valueToBeWritten) 

- Put Writer:
    - WriteData(stream, cid:value)
- Get Writer:
    - Writedata(stream, get:cid:addr)
- Get Reader: 
    - getValueFromStore(cid) string
    - WriteData(stream, value) 

# How it works
  > ./dht-for-values-storage -h 
 
```
  This programs demonstrates the use of a pubsub pattern to broadcast values in the network. 
  Get them stored into some peer (not always the one adding them) and being able to retrieve them from any 
  node that have the CID of the value wanted.

Usage: Run './start
  -h	Display Help
  -host string
    	The bootstrap node host listen address
    	 (default "0.0.0.0")
  -pid string
    	Sets a protocol id for stream headers (default "/chat/1.1.0")
  -port int
    	node listen port (default 4001)
  -topic string
    	Unique string to identify group of nodes (default "/libp2p/example/chat/1.0.0")
```

#### Author
CHAMI Rachid
