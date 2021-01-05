package main

import (
    "fmt"
    ma "github.com/multiformats/go-multiaddr"
    "log"
    "testing"
)

// send an echo from port 7001 to port 7000
func TestMain(m *testing.M) {

    listenHost, err := makeBasicHost(7000, false, 0)
    if err != nil {
        log.Fatal(err)
    }
    // Build host multiaddress
    hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", listenHost.ID().Pretty()))
    addr := listenHost.Addrs()[0]
    fullAddr := addr.Encapsulate(hostAddr)

    pingHost, err := makeBasicHost(7001, false, 0)
    if err != nil {
        log.Fatal(err)
    }

    go echo("", listenHost)
    echo(fullAddr.String(), pingHost)
}
