package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	// We need to import libp2p's libraries that we use in this project.

	"github.com/btwiuse/gost"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/net/gostream"
)

var AutoHandler = gost.AutoHandler()

var AutoProtocol protocol.ID = "/proxy-auto/0.0.1"

// ProxyService provides HTTP proxying on top of libp2p by launching an
// HTTP server which tunnels the requests to a destination peer running
// ProxyService too.
type ProxyServer struct {
	host.Host
}

type ProxyClient struct {
	host.Host
	port int
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
	fmt.Println("listening addresses:")
	for _, a := range h.Addrs() {
		fmt.Printf("- %s/p2p/%s\n", a, h.ID())
	}

	return &ProxyServer{
		Host: h,
	}
}

func (p *ProxyServer) Listen() (net.Listener, error) {
	log.Println("proxy server listening on", AutoProtocol)
	return gostream.Listen(p.Host, AutoProtocol)
}

func (p *ProxyServer) ListenAndServe() error {
	ln, err := p.Listen()
	if err != nil {
		return err
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go p.Handle(conn)
	}
}

func (p *ProxyServer) Handle(conn net.Conn) {
	defer conn.Close()
	log.Println("Got a new conn!")
	AutoHandler.Handle(conn)
	log.Println("conn handled!")
}

func NewProxyClient(h host.Host, port int, dest peer.ID) *ProxyClient {
	fmt.Println("listening addresses:")
	for _, a := range h.Addrs() {
		fmt.Printf("- %s/p2p/%s\n", a, h.ID())
	}

	return &ProxyClient{
		Host: h,
		port: port,
		dest: dest,
	}
}

func (p *ProxyClient) Listen() (net.Listener, error) {
	log.Println("proxy client listening on", p.port)
	return net.Listen("tcp", fmt.Sprintf(":%d", p.port))
}

func (p *ProxyClient) ListenAndServe() error {
	ln, err := p.Listen()
	if err != nil {
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go p.Handle(conn)
	}
	return nil
}

func (p *ProxyClient) Handle(conn net.Conn) {
	defer conn.Close()
	stream, err := p.NewStream(context.Background(), p.dest, AutoProtocol)
	if err != nil {
		log.Println(err)
		return
	}
	defer stream.Close()
	go io.Copy(stream, conn)
	io.Copy(conn, stream)
}
