package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	// We need to import libp2p's libraries that we use in this project.

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

const Auto = "/proxy-auto/0.0.1"

// ProxyService provides HTTP proxying on top of libp2p by launching an
// HTTP server which tunnels the requests to a destination peer running
// ProxyService too.
type ProxyServer struct {
	host.Host
}

type ProxyClient struct {
	host.Host
	dest peer.ID
}

// NewProxyService attaches a proxy service to the given libp2p Host.
// The proxyAddr parameter specifies the address on which the
// HTTP proxy server listens. The dest parameter specifies the peer
// ID of the remote peer in charge of performing the HTTP requests.
//
// ProxyAddr/dest may be nil/"" it is not necessary that this host
// provides a listening HTTP server (and instead its only function is to
// perform the proxied http requests it receives from a different peer.
//
// The addresses for the dest peer should be part of the host's peerstore.
func NewProxyServer(h host.Host) *ProxyServer {
	// We let our host know that it needs to handle streams tagged with the
	// protocol id that we have defined, and then handle them to
	// our own streamHandling function.
	h.SetStreamHandler(Auto, AutoHandler)

	fmt.Println("Proxy server is ready")
	fmt.Println("libp2p-peer addresses:")
	for _, a := range h.Addrs() {
		fmt.Printf("%s/p2p/%s\n", a, h.ID())
	}

	return &ProxyServer{
		Host: h,
	}
}

func NewProxyClient(h host.Host, dest peer.ID) *ProxyClient {
	fmt.Println("Proxy client is ready")
	fmt.Println("libp2p-peer addresses:")
	for _, a := range h.Addrs() {
		fmt.Printf("%s/p2p/%s\n", a, h.ID())
	}

	return &ProxyClient{
		Host: h,
		dest: dest,
	}
}

func (p *ProxyClient) ServeAuto(port int) {
	fmt.Println("proxy listening on ", port)
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalln(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go p.ServeConn(conn)
	}
}

func (p *ProxyClient) ServeConn(conn net.Conn) {
	defer conn.Close()
	stream, err := p.NewStream(context.Background(), p.dest, Auto)
	if err != nil {
		log.Println(err)
		return
	}
	defer stream.Close()
	go io.Copy(stream, conn)
	io.Copy(conn, stream)
}
