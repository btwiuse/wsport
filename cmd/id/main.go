package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"log/slog"
	"os"

	// We need to import libp2p's libraries that we use in this project.
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/btwiuse/wsport"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getPrivKey() (crypto.PrivKey, error) {
	if pk := getEnv("PRIV_KEY", ""); pk != "" {
		keyBytes, err := crypto.ConfigDecodeKey(pk)
		if err != nil {
			return nil, err
		}
		return crypto.UnmarshalPrivateKey(keyBytes)
	}

	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

var RELAY = getEnv("RELAY", "https://example.com")

// makeRandomHost creates a libp2p host with a randomly generated identity.
// This step is described in depth in other tutorials.
func makeRandomHost(port int) host.Host {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	privKey, _ := getPrivKey()
	addr := fmt.Sprintf(RELAY+"/ws%d", port)
	maddr, _ := wsport.FromString(addr)
	peerID, _ := peer.IDFromPublicKey(privKey.GetPublic())
	maddr = maddr.Encapsulate(ma.StringCast("/p2p/" + peerID.String()))
	slog.Info("Listen", "addr", addr, "maddr", maddr)
	host, err := libp2p.New(
		libp2p.Transport(wsport.New),
		wsport.ListenAddrStrings(addr),
		// libp2p.ListenAddrs(maddr),
		libp2p.Identity(privKey),
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
