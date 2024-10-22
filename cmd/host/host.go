package main

import (
	"os"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"

	"github.com/btwiuse/p2pid"
	"github.com/btwiuse/wsport"
)

func newHost(addr string) (host.Host, error) {
	identity, err := p2pid.PersistentIdentity()
	if err != nil {
		return nil, err
	}

	return libp2p.New(
		identity,
		libp2p.ProtocolVersion(os.Getenv("PROTOCOL_VERSION")),
		libp2p.UserAgent(os.Getenv("USER_AGENT")),
		libp2p.Transport(wsport.New),
		// disable /libp2p/circuit/relay/0.2.0/stop
		// libp2p.DisableRelay(),
		wsport.ListenAddrStrings(addr),
	)
}
