package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"sync"
	"time"

	sha "crypto/sha256"
	c "github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	p "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	peer "github.com/libp2p/go-libp2p-peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"
	"strings"
)

var ctx context.Context
var dht *kaddht.IpfsDHT
var valueStore map[string]string
var ps *pubsub.PubSub
var wg sync.WaitGroup
var multiAddr string

// Initializing global Variables used
func setter(ct context.Context, d *kaddht.IpfsDHT, p *pubsub.PubSub, ma string) {
	ctx = ct
	dht = d
	valueStore = make(map[string]string)
	ps = p
	multiAddr = ma
}

// Broadcasts a normal message into the network
func sendMessage(ps *pubsub.PubSub, msg string) {

	msgId := make([]byte, 10)
	_, err := rand.Read(msgId)
	defer func() {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()
	if err != nil {
		return
	}
	now := time.Now().Unix()
	req := &Request{
		Type: Request_SEND_MESSAGE.Enum(),
		SendMessage: &SendMessage{
			Id:      msgId,
			Data:    []byte(msg),
			Created: &now,
		},
	}
	msgBytes, err := req.Marshal()
	if err != nil {
		return
	}
	err = ps.Publish(pubsubTopic, msgBytes)

}

// Changes peer name from short String ID to a name specified via '/name newName'
func updatePeer(ps *pubsub.PubSub, id peer.ID, handle string) {
	oldHandle, ok := handles[id.String()]
	if !ok {
		oldHandle = id.ShortString()
	}
	handles[id.String()] = handle

	req := &Request{
		Type: Request_UPDATE_PEER.Enum(),
		UpdatePeer: &UpdatePeer{
			UserHandle: []byte(handle),
		},
	}
	reqBytes, err := req.Marshal()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	err = ps.Publish(pubsubTopic, reqBytes)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	fmt.Printf("%s -> %s\n", oldHandle, handle)
}

// Input loop
func chatInputLoop(ctx context.Context, h host.Host, ps *pubsub.PubSub, donec chan struct{}, dht *kaddht.IpfsDHT) {

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		msg := scanner.Text()

		if strings.HasPrefix(msg, "/peers") { // Asking for list of peers '/peers'
			fmt.Println("[*] Peers:")
			fmt.Println(dht.RoutingTable().ListPeers())
			fmt.Println("\n\n")

		} else if strings.HasPrefix(msg, "/name ") { // Asking to change the name '/name newName'
			newHandle := strings.TrimPrefix(msg, "/name ")
			newHandle = strings.TrimSpace(newHandle)
			updatePeer(ps, h.ID(), newHandle)

		} else if strings.HasPrefix(msg, "/get ") { // Asking to get a value from the network '/get CID'
			newHandle := strings.TrimPrefix(msg, "/get ")
			newHandle = strings.TrimSpace(newHandle)
			getValue(dht, ctx, newHandle)

		} else if strings.HasPrefix(msg, "/put ") { // Put value in the network '/put value'
			newHandle := strings.TrimPrefix(msg, "/put ")
			newHandle = strings.TrimSpace(newHandle)
			addValue(dht, ctx, newHandle)

		} else { // Send a normal message
			sendMessage(ps, msg)

		}
	}
	donec <- struct{}{}
}

// Gets info of the node providing the value for the contentID
func getProviderInfo(dht *kaddht.IpfsDHT, ctx context.Context, contentID string) (_ p.AddrInfo) {

	cid, err := c.Decode(contentID)
	if err != nil {
		fmt.Println(err)
	}
	providers, err := dht.FindProviders(ctx, cid)

	if len(providers) == 0 {
		fmt.Println("No Providers Found...")
		return
	}

	closest_info, err3 := dht.FindPeer(ctx, providers[0].ID)
	if err3 != nil {
		return p.AddrInfo{ID: "me"}
	}
	return closest_info
}

// Find Peer that has the Closest ID to the CID of the value to add
func findClosestPeerInfo(dht *kaddht.IpfsDHT, ctx context.Context, cid string) p.AddrInfo {

	closest_peer_channel, err := dht.GetClosestPeers(ctx, cid)
	if err != nil { // Case that there is only one peer in the network
		fmt.Println("Couldnt find closest peer ... ", err)
		fmt.Println("Pushing in local ...")

		return p.AddrInfo{
			ID:    dht.Host().ID(),
			Addrs: dht.Host().Addrs(),
		}
	}

	closest_peer := <-closest_peer_channel

	closest_info, err3 := dht.FindPeer(ctx, closest_peer)
	if err3 != nil {
		fmt.Println("Couldnt find closest peer found ...", err3)
		return p.AddrInfo{}
	}

	return closest_info

}

// Function that gets executed when running '/get CID' to send a request to the peer providing the value. This peer will open a connection stream with the requester and sends it back the value
func getValue(dht *kaddht.IpfsDHT, ctx context.Context, cid_with_addr string) {

	// cid_with_addr is in the format "cid:addr"
	data := strings.Split(cid_with_addr, ":")
	contentID := data[0]

	// Getting informations of the node providing the value
	closest_info := getProviderInfo(dht, ctx, contentID)

	// If the node providing the value is itself the one requesting it.. It returns a 'peer.AddrInfo' with 'me' as ID
	if closest_info.ID == "me" {
		fmt.Printf("\x1b[32m%s\x1b[0m>\n ", getValueFromStore(contentID))
		return
	}

	// Opening a direct connection with the provider
	err := dht.Host().Connect(ctx, closest_info)
	if err != nil {
		fmt.Println(err)
		return
	}

	stream, err := dht.Host().NewStream(ctx, closest_info.ID, protocol.ID("tcp"))
	if err != nil {
		fmt.Println("Couldnt open stream")
	}

	getWriter(stream, cid_with_addr)
}

// Writes in the stream a valueget request in the format "get:cid:req_multiAdd"
func getWriter(stream network.Stream, contentID string) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	writeData(rw, "get:"+contentID+":"+multiAddr)
}

// When a peer asks for a value, the provider runs this function to send the value back to him
func getReader(req_multAddr string, contentID string) {
	value := getValueFromStore(contentID)

	fmt.Println("\n!! Sending this CID this CID: \x1b[32m", contentID, " \x1b[0m to somebody!")

	// Triming off the \n at the end
	req_multAddr = strings.TrimSuffix(req_multAddr, "\n")

	// Turn the destination into a multiaddr
	req_maddr, err := multiaddr.NewMultiaddr(req_multAddr)
	if err != nil {
		fmt.Println(err)
	}

	// Extract the peer ID from the multiaddr
	req_info, err := p.AddrInfoFromP2pAddr(req_maddr)
	if err != nil {
		fmt.Println(err)
	}

	// Add the address into the peer store to connect to it
	dht.Host().Peerstore().AddAddr(req_info.ID, req_info.Addrs[0], peerstore.PermanentAddrTTL)

	stream, err := dht.Host().NewStream(ctx, req_info.ID, protocol.ID("tcp"))
	if err != nil {
		fmt.Println("Couldnt open stream: ", err)
		return
	}
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	writeData(rw, "=> "+string(value))
}

// Functions get called when we execute '/put value'
func addValue(dht *kaddht.IpfsDHT, ctx context.Context, value string) {
	// Creating the sha2-256 of the value provided
	buf := []byte(value)
	buff := sha.Sum256(buf)
	mhbuf, _ := mh.EncodeName(buff[:], "sha2-256")

	// Generating a CID from that hash
	cid, err := c.Cast(mhbuf)
	fmt.Println("=> Get it by \x1b[32m/get", cid.String(), "\x1b[0m")
	if err != nil {
		fmt.Println(err)
	}

	pushValue(dht, ctx, cid.String(), value)
}

// function that searches for the closest peer to the CID of the value to be pushed and establishes a direct connection with it to send the value
func pushValue(dht *kaddht.IpfsDHT, ctx context.Context, cid string, value string) {

	// Find closest peer to the CID
	closest_info := findClosestPeerInfo(dht, ctx, cid)

	// If the closest peer to the value is the one pushing it to the network, it gets stored locally
	if closest_info.ID == dht.Host().ID() {
		fmt.Println("Pushing : \x1b[32m", cid, ":", value, "\x1b[0m")
		pushToStore(cid, value)

		c, err := c.Decode(cid)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Broadcasting that we are holding that value
		dht.Provide(ctx, c, true)
		return
	}

	err4 := dht.Host().Connect(ctx, closest_info)
	if err4 != nil {
		fmt.Println(err4)
		return
	}

	stream, err5 := dht.Host().NewStream(ctx, closest_info.ID, protocol.ID("tcp"))
	if err5 != nil {
		fmt.Println("Stream open failed", err5)
	} else {
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
		putWriter(rw, cid+":"+value)
	}

}

// Writes the pushValue data in the stream
func putWriter(rw *bufio.ReadWriter, data string) {
	writeData(rw, data)
}

// Gets executed when there is an incoming connection
func handleStream(stream network.Stream) {

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go readData(rw)

	wg.Add(1)
	wg.Wait()
	if stream.Close() != nil {
		fmt.Println("Couldnt close stream")
	}
}

// Reads the data from the stream opened from a requesting node
func readData(rw *bufio.ReadWriter) {
	defer wg.Done()

	str, err := rw.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading from buffer")
		panic(err)
	}

	if str == "" {
		return
	}
	if str != "\n" {
		// Green console colour: 	\x1b[32m
		// Reset console colour: 	\x1b[0m
		data := strings.Split(str, ":")

		if len(data) == 1 { // "value"

			fmt.Printf("\x1b[32m%s\x1b[0m>\n ", str)

		} else if len(data) == 3 { // "get:cid:multiAddr"

			getReader(data[2], data[1])

		} else if len(data) == 2 { // "cid:value"
			fmt.Printf("Someone just asked me to keep this with me: \x1b[32m%s\x1b[0m> ", data[0]+":"+data[1]+"\n")

			pushToStore(data[0], data[1])

			cid, _ := c.Decode(data[0])

			err := dht.Provide(ctx, cid, true)
			if err != nil {
				fmt.Println("Provider Problem: ", err)
			}
		}
	}

}

// Write data to the buffer to be send to the other person
func writeData(rw *bufio.ReadWriter, data string) {

	_, err := rw.WriteString(fmt.Sprintf("%s\n", data))
	if err != nil {
		fmt.Println("Error writing to buffer")
		panic(err)
	}
	err = rw.Flush()
	if err != nil {
		fmt.Println("Error flushing buffer")
		panic(err)
	}

}

// Access the map to get the value or to push them
func pushToStore(cid string, value string) {
	valueStore[cid] = value

}

func getValueFromStore(cid string) string {
	return valueStore[cid]
}
