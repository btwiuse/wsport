package main

import (
	"fmt"
	"log"
	"os"
	"time"

	// We need to import libp2p's libraries that we use in this project.
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/btwiuse/wsport"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Run(args []string) error {
	relay, err := wsport.FromString(getEnv("RELAY", "https://example.com"))
	if err != nil {
		return err
	}

	host, err := libp2p.New(
		libp2p.Transport(wsport.New),
	)
	if err != nil {
		return err
	}

	notifiee := &Notifiee{
		onListen: func(n network.Network, a ma.Multiaddr) {
			log.Println("[Listen]", a)
			log.Println("host is online. libp2p-peer addresses:")
			for _, a := range host.Addrs() {
				log.Println(fmt.Sprintf("%s/p2p/%s", a, host.ID()))
			}
		},
		onListenClose: func(n network.Network, a ma.Multiaddr) {
			log.Println("[ListenClose]", a)
			for i := 0; ; i++ {
				err := host.Network().Listen(relay)
				if err == nil {
					break
				}
				log.Println(err, "retry in", i, "seconds")
				time.Sleep(time.Duration(i) * time.Second)
			}
		},
	}

	host.Network().Notify(notifiee)

	err = host.Network().Listen(relay)
	if err != nil {
		return err
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
