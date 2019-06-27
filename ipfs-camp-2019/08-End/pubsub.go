package main

import (
	"context"
	"fmt"
	"os"

	peer "github.com/libp2p/go-libp2p-peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

var handles = map[string]string{}

const pubsubTopic = "/libp2p/example/chat/1.0.0"

func pubsubMessageHandler(id peer.ID, msg *SendMessage) {
	handle, ok := handles[id.String()]
	if !ok {
		handle = id.ShortString()
	}
	fmt.Printf("%s: %s\n", handle, msg.Data)
}

func pubsubUpdateHandler(id peer.ID, msg *UpdatePeer) {
	oldHandle, ok := handles[id.String()]
	if !ok {
		oldHandle = id.ShortString()
	}
	handles[id.String()] = string(msg.UserHandle)
	fmt.Printf("%s -> %s\n", oldHandle, msg.UserHandle)
}

func pubsubHandler(ctx context.Context, sub *pubsub.Subscription) {
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		req := &Request{}
		err = req.Unmarshal(msg.Data)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		switch *req.Type {
		case Request_SEND_MESSAGE:
			pubsubMessageHandler(msg.GetFrom(), req.SendMessage)
		case Request_UPDATE_PEER:
			pubsubUpdateHandler(msg.GetFrom(), req.UpdatePeer)
		}
	}
}
