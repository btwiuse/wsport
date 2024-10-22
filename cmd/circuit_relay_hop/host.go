package main

import (
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
		libp2p.Transport(wsport.New),
		wsport.ListenAddrStrings(addr),
	)
}
