package main

import (
	"context"
	"fmt"
	"log"
	"os"

	ma "github.com/multiformats/go-multiaddr"

	"github.com/btwiuse/wsport"
	"github.com/btwiuse/wsport/cmd"
)

func Run(args []string) error {
	host, err := newHost()
	if err != nil {
		return err
	}

	addr := cmd.RELAY
	fmt.Println("relay addr:", addr)

	relayMa, err := wsport.FromString(addr)
	if err != nil {
		return err
	}

	cmd.Notify(host, relayMa)

	err = host.Network().Listen(relayMa)
	if err != nil {
		return err
	}

	fmt.Println("registered protocols:")
	for _, protocol := range host.Mux().Protocols() {
		fmt.Println("-", protocol)
	}

	// Connect to the specified peers
	for _, addr := range args {
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return err
		}
		peerInfo, err := AddrInfo(maddr)
		if err != nil {
			return err
		}
		err = host.Connect(context.Background(), *peerInfo)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := Run(os.Args[1:])
	if err != nil {
		log.Fatalln(err)
	}

	select {}
}
