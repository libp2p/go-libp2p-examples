package main

import (
	"context"
	"fmt"
	"os"

	peer "github.com/libp2p/go-libp2p-peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const pubsubTopic = "/libp2p/example/chat/1.0.0"

func pubsubMessageHandler(id peer.ID, msg *SendMessage) {
	fmt.Printf("%s: %s\n", id.ShortString(), msg.Data)
}

func pubsubUpdateHandler(id peer.ID, msg *UpdatePeer) {

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
