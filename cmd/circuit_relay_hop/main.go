package main

import (
	"fmt"

	"github.com/btwiuse/wsport/cmd"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
)

func main() {
	addr := cmd.RELAY
	fmt.Println("relay addr:", addr)

	host, err := newHost(addr)
	if err != nil {
		panic(err)
	}

	// register new protocol libp2p circuitv2 relay
	// /libp2p/circuit/relay/0.2.0/hop
	// returns *relay.Relay
	_, err = relay.New(host)
	if err != nil {
		panic(err)
	}

	fmt.Println("listening addresses:")
	for _, a := range host.Addrs() {
		fmt.Printf("- %s/p2p/%s\n", a, host.ID())
	}

	fmt.Println("registered protocols:")
	for _, protocol := range host.Mux().Protocols() {
		fmt.Println("-", protocol)
	}

	select {}
}
