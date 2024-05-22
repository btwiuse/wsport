package main

import (
	"fmt"

	"github.com/libp2p/go-libp2p"
)

func main() {
	host, err := libp2p.New()
	if err != nil {
		panic(err)
	}

	// Print the host's ID
	fmt.Println("Host ID:", host.ID())
}
