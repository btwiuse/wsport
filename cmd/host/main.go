package main

import (
	"fmt"

	"github.com/btwiuse/wsport/cmd"
)

func main() {
	addr := cmd.RELAY
	fmt.Println("relay addr:", addr)

	host, err := newHost(addr)
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
