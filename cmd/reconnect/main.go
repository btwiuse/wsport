package main

import (
	"log"
	"os"

	// We need to import libp2p's libraries that we use in this project.
	"github.com/libp2p/go-libp2p"

	"github.com/btwiuse/wsport"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

var RELAY = getEnv("RELAY", "https://example.com")

func Run(args []string) error {
	relay, err := wsport.FromString(RELAY)
	if err != nil {
		return err
	}

	host, err := libp2p.New(
		libp2p.Transport(wsport.New),
	)
	if err != nil {
		return err
	}

	Notify(host, relay)

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
