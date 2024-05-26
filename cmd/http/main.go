package main

import (
	"context"
	"log"
	"net/http"
	"os"

	// We need to import libp2p's libraries that we use in this project.
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
	p2phttp "github.com/libp2p/go-libp2p/p2p/http"
	"github.com/libp2p/go-libp2p/p2p/net/gostream"
	ma "github.com/multiformats/go-multiaddr"

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

	go ListenAndServe(host, p2phttp.ProtocolIDForMultistreamSelect, http.FileServer(http.Dir(".")))

	err = host.Network().Listen(relay)
	if err != nil {
		return err
	}

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

func ListenAndServe(host host.Host, p protocol.ID, handler http.Handler) error {
	ln, err := gostream.Listen(host, p)
	if err != nil {
		return err
	}

	return http.Serve(ln, handler)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := Run(os.Args[1:])
	if err != nil {
		log.Fatalln(err)
	}

	select {}
}
