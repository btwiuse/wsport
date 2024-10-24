package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	// We need to import libp2p's libraries that we use in this project.
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"

	// "github.com/libp2p/go-libp2p/core/routing"

	ds "github.com/ipfs/go-datastore"
	dsync "github.com/ipfs/go-datastore/sync"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"

	"github.com/btwiuse/p2pid"
	"github.com/btwiuse/wsport"
	"github.com/btwiuse/wsport/cmd"
)

func Run(args []string) error {
	relayURL := cmd.RELAY

	identity, err := p2pid.PersistentIdentity()
	if err != nil {
		return err
	}

	options := []libp2p.Option{
		libp2p.ProtocolVersion(os.Getenv("PROTOCOL_VERSION")),
		libp2p.UserAgent(os.Getenv("USER_AGENT")),
		libp2p.Transport(wsport.New),
		identity,
	}

	if len(args) == 0 {
		/*
			router := libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
				return dht.New(context.Background(), h)
			})
		*/

		// options = append(options, router)

		path := "/bootstrap"
		if domain := os.Getenv("DOMAIN"); domain != "" {
			path = domain
		}

		relayURL += path
	}

	relay, err := wsport.FromString(relayURL)
	if err != nil {
		return err
	}

	var host host.Host
	host, err = libp2p.New(
		options...,
	)
	if err != nil {
		return err
	}

	var ipfsdht *dht.IpfsDHT
	// if len(args) != 0 {
	if true {
		// Construct a datastore (needed by the DHT). This is just a simple, in-memory thread-safe datastore.
		dstore := dsync.MutexWrap(ds.NewMapDatastore())

		customOptions := []dht.Option{
			dht.Datastore(dstore),
			dht.Mode(dht.ModeServer),
		}

		dhtOverride := os.Getenv("DHT")
		if dhtOverride != "" {
			customOptions = append(customOptions, dht.V1ProtocolOverride(protocol.ID(dhtOverride)))
		}

		// Make the DHT
		ipfsdht, err = dht.New(
			context.Background(),
			host,
			customOptions...,
		)
		if err != nil {
			return err
		}

		host = rhost.Wrap(host, ipfsdht)
	}

	cmd.Notify(host, relay)

	MatchUnknownProtocol(host)

	err = host.Network().Listen(relay)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return nil
	}

	go KeepBootnode(host, args)

	if len(args) > 1 {
		// Bootstrap the dht
		err = ipfsdht.Bootstrap(context.Background())
		if err != nil {
			return err
		}
	}

	return nil

	go func() {
		for {
			rt := ipfsdht.RoutingTable()
			for i, p := range rt.ListPeers() {
				log.Println("dht peer", i, p)
			}
			time.Sleep(5 * time.Second)
			continue
			println("refreshing routing table")
			if err := <-ipfsdht.ForceRefresh(); err != nil {
				log.Println(err)
				time.Sleep(5 * time.Second)
			}
		}
	}()

	return nil
}

func MatchUnknownProtocol(h host.Host) {
	matcherFunc := func(p protocol.ID) bool {
		fmt.Printf("Catch-all matcher called for protocol: %s\n", p)
		return true
	}

	handlerFunc := func(s network.Stream) {
		s.Reset()
	}

	h.SetStreamHandlerMatch("/", matcherFunc, handlerFunc)

}

func KeepBootnode(host host.Host, addrs []string) {
	for {
		err := Bootnode(host, addrs)
		if err != nil {
			log.Println("KeepBootnode", err)
		}
		time.Sleep(5 * time.Second)
	}
}

func Bootnode(host host.Host, addrs []string) error {
	for _, peerAddr := range addrs {
		peerMa, err := ma.NewMultiaddr(peerAddr)
		if err != nil {
			return err
		}

		_, peerID := peer.SplitAddr(peerMa)

		if host.Network().Connectedness(peerID) == network.Connected {
			continue
		}

		log.Println("Connecting to bootstrap", peerAddr)

		peerInfo, err := peer.AddrInfoFromP2pAddr(peerMa)
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
