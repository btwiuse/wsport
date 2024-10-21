package main

import (
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/btwiuse/p2pid"
)

func main() {
	identity, err := p2pid.PersistentIdentity()
	if err != nil {
		panic(err)
	}

	host, err := libp2p.New(identity)
	if err != nil {
		panic(err)
	}

	// Print the host's ID
	fmt.Println("Host ID:", host.ID())
}
