package main

import (
	"context"
	"fmt"
	"log"
	"os"

	ma "github.com/multiformats/go-multiaddr"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	addr := RELAY
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

	for _, arg := range os.Args[1:] {
		maddr, err := ma.NewMultiaddr(arg)
		if err != nil {
			log.Fatalln(err)
		}

		addrInfo, err := AddrInfo(maddr)
		if err != nil {
			log.Fatalln(err)
		}

		err = host.Connect(context.Background(), *addrInfo)
		if err != nil {
			log.Fatalln(err)
		} else {
			fmt.Println("Connected to", addrInfo)
		}
	}

	select {}
}
