package main

import (
	"net"

	"github.com/libp2p/go-libp2p/core/network"
)

// conn is an implementation of net.Conn which wraps
// libp2p streams.
type conn struct {
	network.Stream
}

// newConn creates a conn given a libp2p stream
func newConn(s network.Stream) net.Conn {
	return &conn{s}
}

// LocalAddr returns the local network address.
func (c *conn) LocalAddr() net.Addr {
	return &addr{c.Stream.Conn().LocalPeer()}
}

// RemoteAddr returns the remote network address.
func (c *conn) RemoteAddr() net.Addr {
	return &addr{c.Stream.Conn().RemotePeer()}
}
