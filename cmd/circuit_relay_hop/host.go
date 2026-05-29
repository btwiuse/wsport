package main

import (
	"os"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"

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
	)
}
