package main

import (
	"context"
	"os"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"

	"github.com/btwiuse/p2pid"
	"github.com/btwiuse/wsport"
)

func newHost(addr string) (host.Host, error) {
	return libp2p.New(
		libp2p.ProtocolVersion(os.Getenv("PROTOCOL_VERSION")),
		libp2p.UserAgent(os.Getenv("USER_AGENT")),
		p2pid.FromEnv(p2pid.PID_SEED),
		libp2p.Transport(wsport.New),
		wsport.ListenAddrStrings(addr),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			return dht.New(
				context.Background(),
				h,
				dht.Mode(dht.ModeAutoServer),
			)
		}),
	)
}
