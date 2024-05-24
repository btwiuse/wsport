package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	// We need to import libp2p's libraries that we use in this project.
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"

	"github.com/btwiuse/wsport"
	ma "github.com/multiformats/go-multiaddr"
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
	relay := fmt.Sprintf(RELAY+"/ws%d", port)
	log.Println("ListenAddr", relay)
	host, err := libp2p.New(
		libp2p.Transport(wsport.New),
		// libp2p.ListenAddrStrings(addr),
		wsport.ListenAddrStrings(relay),
	)
	if err != nil {
		log.Fatalln(err)
	}
	relayMA, err := wsport.FromString(relay)
	if err != nil {
		log.Fatalln(err)
	}
	Notify(host, relayMA)
	return host
}

// addAddrToPeerstore parses a peer multiaddress and adds
// it to the given host's peerstore, so it knows how to
// contact it. It returns the peer ID of the remote peer.
func addAddrToPeerstore(h host.Host, addr ma.Multiaddr) peer.ID {
	addrinfo, err := AddrInfo(addr)
	if err != nil {
		log.Fatalln(err)
	}

	for _, addr := range addrinfo.Addrs {
		h.Peerstore().AddAddr(addrinfo.ID, addr, peerstore.PermanentAddrTTL)
	}

	return addrinfo.ID
}

const help = `
This example creates a simple HTTP Proxy using two libp2p peers. The first peer
provides an HTTP server locally which tunnels the HTTP requests with libp2p
to a remote peer. The remote peer performs the requests and 
send the sends the response back.

Usage: Start remote peer first with:   ./proxy
       Then start the local peer with: ./proxy -d <remote-peer-multiaddress>

Then you can do something like: curl -x "localhost:9900" "http://ipfs.io".
This proxies sends the request through the local peer, which proxies it to
the remote peer, which makes it and sends the response back.`

func main() {
	flag.Usage = func() {
		fmt.Println(help)
		flag.PrintDefaults()
	}

	// Parse some flags
	destPeer := flag.String("d", "", "destination peer address")
	port := flag.Int("p", 9900, "proxy port")
	p2pport := flag.Int("l", 12000, "libp2p listen port")
	flag.Parse()

	// If we have a destination peer we will start a local client
	if *destPeer != "" {
		// We use p2pport+1 in order to not collide if the user
		// is running the remote peer locally on that port
		host := makeRandomHost(*p2pport + 1)
		destPeerMA, err := ma.NewMultiaddr(*destPeer)
		if err != nil {
			log.Fatalln(err)
		}
		// Make sure our host knows how to reach destPeer
		destPeerID := addAddrToPeerstore(host, destPeerMA)
		// Create the proxy client and start the http server
		client := NewProxyClient(host, destPeerID)
		client.ServeAuto(*port) // hangs forever
	} else {
		host := makeRandomHost(*p2pport)
		// In this case we only need to make sure our host
		// knows how to handle incoming proxied requests from
		// another peer.
		_ = NewProxyServer(host)
		select {} // hang forever
	}
}
