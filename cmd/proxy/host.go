package main

import (
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"

	"github.com/btwiuse/p2pid"
	"github.com/btwiuse/wsport"
)

func newHost(relay string) (host.Host, error) {
	identity, err := p2pid.PersistentIdentity()
	if err != nil {
		return nil, err
	}

	host, err := libp2p.New(
		libp2p.ProtocolVersion(os.Getenv("PROTOCOL_VERSION")),
		libp2p.UserAgent(os.Getenv("USER_AGENT")),
		identity,
		libp2p.Transport(wsport.New),
		wsport.ListenAddrStrings(relay),
	)
	if err != nil {
		return nil, err
	}

	relayMA, err := wsport.FromString(relay)
	if err != nil {
		return nil, err
	}

	Notify(host, relayMA)

	fmt.Println("registered protocols:")
	for _, protocol := range host.Mux().Protocols() {
		fmt.Println("-", protocol)
	}

	return host, nil
}
