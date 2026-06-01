package main

import (
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"

	"github.com/btwiuse/p2pid"
)

func newHost() (host.Host, error) {
	return libp2p.New(
		p2pid.FromEnv(p2pid.PID_SEED),
	)
}
