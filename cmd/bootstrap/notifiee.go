package main

import (

	// We need to import libp2p's libraries that we use in this project.

	"github.com/libp2p/go-libp2p/core/network"
	ma "github.com/multiformats/go-multiaddr"
)

type Notifiee struct {
	onListen       func(network.Network, ma.Multiaddr)
	onListenClose  func(network.Network, ma.Multiaddr)
	onConnected    func(network.Network, network.Conn)
	onDisconnected func(network.Network, network.Conn)
}

func (N *Notifiee) Listen(n network.Network, a ma.Multiaddr) {
	if N.onListen != nil {
		N.onListen(n, a)
	}
}
func (N *Notifiee) ListenClose(n network.Network, a ma.Multiaddr) {
	if N.onListenClose != nil {
		N.onListenClose(n, a)
	}
}
func (N *Notifiee) Connected(n network.Network, c network.Conn) {
	if N.onConnected != nil {
		N.onConnected(n, c)
	}
}
func (N *Notifiee) Disconnected(n network.Network, c network.Conn) {
	if N.onDisconnected != nil {
		N.onDisconnected(n, c)
	}
}
