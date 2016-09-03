package main

import (
	"context"
	"fmt"
	"net"
	"os"

	transport "github.com/ipfs/go-libp2p-transport"
	ma "github.com/jbenet/go-multiaddr"
	smux "github.com/jbenet/go-stream-muxer"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
)

func fatal(i interface{}) {
	fmt.Println(i)
	os.Exit(1)
}

type NullMux struct{}

type NullMuxConn struct {
	net.Conn
}

func (c *NullMuxConn) AcceptStream() (smux.Stream, error) {
	panic("We don't do this")
}

func (c *NullMuxConn) IsClosed() bool {
	return false
}

func (c *NullMuxConn) OpenStream() (smux.Stream, error) {
	panic("if only you could see how disappointed i am in you right now")
}

func (c *NullMuxConn) Serve(_ smux.StreamHandler) {
}

func (nm NullMux) NewConn(c net.Conn, server bool) (smux.Conn, error) {
	return &NullMuxConn{c}, nil
}

var _ smux.Transport = (*NullMux)(nil)

func main() {
	laddr, err := ma.NewMultiaddr("/ip4/0.0.0.0/tcp/5555")
	if err != nil {
		fatal(err)
	}

	swarm.PSTransport = new(NullMux)

	s := swarm.NewBlankSwarm(context.Background(), "bob", nil)

	s.AddTransport(transport.NewTCPTransport())

	err = s.AddListenAddr(laddr)
	if err != nil {
		fatal(err)
	}

	s.SetConnHandler(func(c *swarm.Conn) {
		fmt.Println("CALLED OUR CONN HANDLER!")
		defer c.Close()
		buf := make([]byte, 1024)
		for {
			n, err := c.RawConn().Read(buf)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("read: %q\n", string(buf[:n]))

			_, err = c.RawConn().Write(buf[:n])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	})

	<-make(chan bool)
}
