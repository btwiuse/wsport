package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	// We need to import libp2p's libraries that we use in this project.
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	// "github.com/libp2p/go-libp2p/core/routing"

	ds "github.com/ipfs/go-datastore"
	dsync "github.com/ipfs/go-datastore/sync"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"

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

func Run(args []string) error {
	relayURL := RELAY

	options := []libp2p.Option{
		libp2p.Transport(wsport.New),
	}

	if len(args) == 0 {
		identity, err := ezlibp2p.PersistentIdentity()
		if err != nil {
			return err
		}

		options = append(options, identity)

		/*
			router := libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
				return dht.New(context.Background(), h)
			})
		*/

		// options = append(options, router)

		relayURL += "/bootstrap"
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

		// Make the DHT
		ipfsdht, err = dht.New(
			context.Background(),
			host,
			dht.Datastore(dstore),
			dht.Mode(dht.ModeServer),
		)
		if err != nil {
			return err
		}

		host = rhost.Wrap(host, ipfsdht)
	}

	Notify(host, relay)

	err = host.Network().Listen(relay)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return nil
	}

	for _, peerAddr := range args {
		log.Println("Connecting to bootstrap", peerAddr)
		peerMa, err := ma.NewMultiaddr(peerAddr)
		if err != nil {
			return err
		}

		peerInfo, err := peer.AddrInfoFromP2pAddr(peerMa)
		if err != nil {
			return err
		}

		err = host.Connect(context.Background(), *peerInfo)
		if err != nil {
			return err
		}
	}

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

func Notify(host host.Host, relay ma.Multiaddr) {
	notifiee := &Notifiee{
		onListen: func(n network.Network, a ma.Multiaddr) {
			slog.Info(
				"[Listen]",
				"ma", fmt.Sprintf("%s/p2p/%s", a, host.ID()),
				// "localAddrs", host.Addrs(),
			)
			for i, addr := range host.Addrs() {
				log.Println("localAddr", i, addr)
			}
		},
		onListenClose: func(n network.Network, a ma.Multiaddr) {
			slog.Info(
				"[ListenClose]",
				"ma", fmt.Sprintf("%s/p2p/%s", a, host.ID()),
				// "localAddrs", host.Addrs(),
			)
			for i, addr := range host.Addrs() {
				log.Println("localAddr", i, addr)
			}
			for i := 0; ; i++ {
				err := n.Listen(relay)
				if err == nil {
					break
				}
				log.Println(err, "retry in", i, "seconds")
				time.Sleep(time.Duration(i) * time.Second)
			}
		},
		onConnected: func(n network.Network, c network.Conn) {
			slog.Info(
				"[Connected]",
				"connId", c.ID(),
				"connRemotePeerId", c.RemotePeer(),
				"direction", c.Stat().Direction.String(),
				// "opened", c.Stat().Opened,
				// "peers", host.Peerstore().Peers(),
				// "connLocalPeerId", c.LocalPeer(),
				// "connLocalMa", c.LocalMultiaddr(),
				// "connRemoteMa", c.RemoteMultiaddr(),
			)
			for i, addr := range host.Peerstore().Peers() {
				log.Println("peer", i, addr, n.Connectedness(addr).String())
			}
		},
		onDisconnected: func(n network.Network, c network.Conn) {
			slog.Info(
				"[Disconnected]",
				"connId", c.ID(),
				"connRemotePeerId", c.RemotePeer(),
				"direction", c.Stat().Direction.String(),
				"opened", c.Stat().Opened,
				// "peers", host.Peerstore().Peers(),
				// "connLocalPeerId", c.LocalPeer(),
				// "connLocalMa", c.LocalMultiaddr(),
				// "connRemoteMa", c.RemoteMultiaddr(),
			)
			for i, addr := range host.Peerstore().Peers() {
				log.Println("peer", i, addr, n.Connectedness(addr).String())
			}
		},
	}

	host.Network().Notify(notifiee)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := Run(os.Args[1:])
	if err != nil {
		log.Fatalln(err)
	}

	select {}
}
