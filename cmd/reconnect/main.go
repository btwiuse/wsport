package main

import (
	"fmt"
	"log"
	"log/slog"
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

	notifiee := &network.NotifyBundle{
		ListenF: func(n network.Network, a ma.Multiaddr) {
			slog.Info(
				"[Listen]",
				"ma", fmt.Sprintf("%s/p2p/%s", a, host.ID()),
				// "localAddrs", host.Addrs(),
			)
			for i, addr := range host.Addrs() {
				log.Println("localAddr", i, addr)
			}
		},
		ListenCloseF: func(n network.Network, a ma.Multiaddr) {
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
		ConnectedF: func(n network.Network, c network.Conn) {
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
		DisconnectedF: func(n network.Network, c network.Conn) {
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
