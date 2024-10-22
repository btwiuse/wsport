package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"

	"github.com/btwiuse/wsport/cmd"
	ma "github.com/multiformats/go-multiaddr"
)

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

type App struct {
	destPeer *string
	port     *int
}

func (a *App) IsClient() bool {
	return *a.destPeer != ""
}

func (a *App) ProxyClient(relay string) *ProxyClient {
	host, err := newHost(relay)
	if err != nil {
		log.Fatalln(err)
	}
	destPeerMA, err := ma.NewMultiaddr(*a.destPeer)
	if err != nil {
		log.Fatalln(err)
	}
	destPeerID := addAddrToPeerstore(host, destPeerMA)
	return NewProxyClient(host, *a.port, destPeerID)
}

func (a *App) ProxyServer(relay string) *ProxyServer {
	host, err := newHost(relay)
	if err != nil {
		log.Fatalln(err)
	}
	return NewProxyServer(host)
}

func Parse(args []string) (*App, error) {
	flagSet := flag.NewFlagSet("proxy", flag.ContinueOnError)
	flagSet.Usage = func() {
		fmt.Println(help)
		flagSet.PrintDefaults()
	}
	app := &App{
		destPeer: flagSet.String("d", "", "destination peer address. If empty, run as server, otherwise run as client"),
		port:     flagSet.Int("p", 9900, "proxy port"),
	}
	if err := flagSet.Parse(args); err != nil {
		return nil, err
	}
	return app, nil
}

func (a *App) Run() error {
	relay := cmd.RELAY

	if a.IsClient() {
		fmt.Println("client relay addr:", relay)

		return a.ProxyClient(relay).ListenAndServe()
	}

	fmt.Println("server relay addr:", relay)

	return a.ProxyServer(relay).ListenAndServe()
}

func Run(args []string) error {
	app, err := Parse(args)
	if err != nil {
		return err
	}
	return app.Run()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	if err := Run(os.Args[1:]); err != nil {
		log.Fatalln(err)
	}
}
