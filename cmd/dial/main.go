package main

import (
	"context"
	"fmt"
	"log"
	"os"

	// We need to import libp2p's libraries that we use in this project.
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
	tptu "github.com/libp2p/go-libp2p/p2p/net/upgrader"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
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

// makeRandomHost creates a libp2p host with a randomly generated identity.
// This step is described in depth in other tutorials.
func makeRandomHost(port int) host.Host {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	addr := RELAY
	log.Println("ListenAddr", addr)
	host, err := libp2p.New(
		libp2p.Transport(wsport.New),
		wsport.ListenAddrStrings(addr),
		// it is a failed attempt to disable peer id check
		// taken from https://github.com/quochoandkh/go-spacemesh/blob/a06f7e8eff9fa2ce7d9bccdc15e8b40b21495c3b/p2p/host.go#L222
		// safe to delete later
		libp2p.Security(
			noise.ID,
			func(id protocol.ID, privkey crypto.PrivKey, muxers []tptu.StreamMuxer) (*noise.SessionTransport, error) {
				tp, err := noise.New(id, privkey, muxers)
				if err != nil {
					return nil, err
				}
				return tp.WithSessionOptions(
				// noise.DisablePeerIDCheck(),
				)
			},
		),
	)
	if err != nil {
		log.Fatalln(err)
	}
	return host
}

func main() {
	host := makeRandomHost(1000)
	fmt.Println("host is ready")
	fmt.Println("libp2p-peer addresses:")
	for _, a := range host.Addrs() {
		fmt.Printf("%s/p2p/%s\n", a, host.ID())
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
