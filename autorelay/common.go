package main

import (
	"flag"

	"github.com/libp2p/go-libp2p-peer"
)

type peerID peer.ID

func (pid *peerID) String() string {
	return pid.String()
}


func (pid *peerID) Set(value string) error {
	newPID, err := peer.IDFromString(value)
	if err != nil {
		return err
	}
	*pid = peerID(newPID)
	return nil
}

type Config struct {
	RelayID          peerID
	RendezvousString string
}

func parseFlags() (Config, error) {
	config := Config{}
	flag.Var(&config.RelayID, "relayID", "The peer ID of the relay node")
	flag.StringVar(&config.RendezvousString, "rendezvous", "nat-node", "A string to advertise this node for discovery in the DHT")
	flag.Parse()

	return config, nil
}

