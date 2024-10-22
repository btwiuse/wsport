package main

import (
	"net/http"

	// We need to import libp2p's libraries that we use in this project.

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/net/gostream"
)

func ListenAndServe(host host.Host, p protocol.ID, handler http.Handler) error {
	ln, err := gostream.Listen(host, p)
	if err != nil {
		return err
	}

	return http.Serve(ln, handler)
}
