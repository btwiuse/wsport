package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/btwiuse/wsport"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <multiaddr>", os.Args[0])
	}

	host, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.Transport(wsport.New),
	)
	if err != nil {
		log.Fatalln("failed to create host:", err)
	}
	defer host.Close()

	host.Network().Notify(&network.NotifyBundle{
		DisconnectedF: func(n network.Network, c network.Conn) {
			log.Printf("[Disconnected] peer=%s duration=%s", c.RemotePeer(), time.Since(c.Stat().Opened))
		},
	})

	for _, arg := range os.Args[1:] {
		maddr, err := ma.NewMultiaddr(arg)
		if err != nil {
			log.Fatalln("invalid multiaddr:", err)
		}

		addrInfo, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			log.Fatalln("parse addr info:", err)
		}

		err = host.Connect(context.Background(), *addrInfo)
		if err != nil {
			log.Fatalln("connect:", err)
		}
		fmt.Println("connected to", addrInfo)
	}

	select {}
}
