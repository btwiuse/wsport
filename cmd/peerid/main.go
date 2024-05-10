package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	// We need to import libp2p's libraries that we use in this project.
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"

	"github.com/btwiuse/wsport"
	"github.com/watjurk/ezlibp2p"
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
	addr := fmt.Sprintf(RELAY+"/ws%d", port)
	slog.Info("Listen", "addr", addr)

	identity, err := ezlibp2p.PersistentIdentity()
	if err != nil {
		log.Fatalln(err)
	}

	host, err := libp2p.New(
		libp2p.Transport(wsport.New),
		wsport.ListenAddrStrings(addr),
		identity,
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			return dht.New(
				context.Background(),
				h,
				dht.Mode(dht.ModeAutoServer),
			)
		}),
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
	select {}
}
