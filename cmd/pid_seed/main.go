package main

import (
	"fmt"

	"github.com/btwiuse/p2pid"
	"github.com/libp2p/go-libp2p"
)

func main() {
	host, err := libp2p.New(p2pid.FromEnv(p2pid.PID_SEED))
	if err != nil {
		panic(err)
	}

	// Print the host's ID
	fmt.Println("Host ID:", host.ID())
}
