package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	// We need to import libp2p's libraries that we use in this project.
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"

	"github.com/btwiuse/wsport"
	_ "github.com/btwiuse/x-parity-wss"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

var RELAY = getEnv("RELAY", "https://example.com")

// Protocol defines the libp2p protocol that we will use for the libp2p proxy
// service that we are going to provide. This will tag the streams used for
// this service. Streams are multiplexed and their protocol tag helps
// libp2p handle them to the right handler functions.
const Protocol = "/proxy-example/0.0.1"
const Connect = "/proxy-connect/0.0.1"

// makeRandomHost creates a libp2p host with a randomly generated identity.
// This step is described in depth in other tutorials.
func makeRandomHost(port int) host.Host {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	addr := fmt.Sprintf(RELAY+"/ws%d", port)
	log.Println("ListenAddr", addr)
	host, err := libp2p.New(
		libp2p.Transport(wsport.New),
		// libp2p.ListenAddrStrings(addr),
		wsport.ListenAddrStrings(addr),
	)
	if err != nil {
		log.Fatalln(err)
	}
	return host
}

// ProxyService provides HTTP proxying on top of libp2p by launching an
// HTTP server which tunnels the requests to a destination peer running
// ProxyService too.
type ProxyService struct {
	host      host.Host
	dest      peer.ID
	proxyAddr ma.Multiaddr
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
func NewProxyService(h host.Host, proxyAddr ma.Multiaddr, dest peer.ID) *ProxyService {
	// We let our host know that it needs to handle streams tagged with the
	// protocol id that we have defined, and then handle them to
	// our own streamHandling function.
	h.SetStreamHandler(Protocol, streamHandler)
	h.SetStreamHandler(Connect, connectHandler)

	fmt.Println("Proxy server is ready")
	fmt.Println("libp2p-peer addresses:")
	for _, a := range h.Addrs() {
		fmt.Printf("%s/p2p/%s\n", a, h.ID())
	}

	return &ProxyService{
		host:      h,
		dest:      dest,
		proxyAddr: proxyAddr,
	}
}

func (p *ProxyService) ServeConnect() {
	ln, err := net.Listen("tcp", ":9099")
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

// Serve listens on the ProxyService's proxy address. This effectively
// allows to set the listening address as http proxy.
func (p *ProxyService) Serve() {
	_, serveArgs, _ := manet.DialArgs(p.proxyAddr)
	fmt.Println("proxy listening on ", serveArgs)
	if p.dest != "" {
		http.ListenAndServe(serveArgs, p)
	}
}

func (p *ProxyService) ServeConn(conn net.Conn) {
	defer conn.Close()
	stream, err := p.host.NewStream(context.Background(), p.dest, Connect)
	if err != nil {
		log.Println(err)
		return
	}
	defer stream.Close()
	go io.Copy(stream, conn)
	io.Copy(conn, stream)
}

// ServeHTTP implements the http.Handler interface. WARNING: This is the
// simplest approach to a proxy. Therefore, we do not do any of the things
// that should be done when implementing a reverse proxy (like handling
// headers correctly). For how to do it properly, see:
// https://golang.org/src/net/http/httputil/reverseproxy.go?s=3845:3920#L121
//
// ServeHTTP opens a stream to the dest peer for every HTTP request.
// Streams are multiplexed over single connections so, unlike connections
// themselves, they are cheap to create and dispose of.
func (p *ProxyService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("proxying request for %s to peer %s\n", r.URL, p.dest)
	// We need to send the request to the remote libp2p peer, so
	// we open a stream to it
	stream, err := p.host.NewStream(context.Background(), p.dest, Protocol)
	// If an error happens, we write an error for response.
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	tee := io.MultiWriter(os.Stdout, stream)
	// r.Write() writes the HTTP request to the stream.
	err = r.Write(tee)
	if err != nil {
		stream.Reset()
		log.Println(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Now we read the response that was sent from the dest
	// peer
	buf := bufio.NewReader(stream)
	resp, err := http.ReadResponse(buf, r)
	if err != nil {
		stream.Reset()
		log.Println(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Copy any headers
	for k, v := range resp.Header {
		for _, s := range v {
			w.Header().Add(k, s)
		}
	}

	// Write response status and headers
	w.WriteHeader(resp.StatusCode)

	// Finally copy the body
	io.Copy(w, resp.Body)
	resp.Body.Close()
}

// addAddrToPeerstore parses a peer multiaddress and adds
// it to the given host's peerstore, so it knows how to
// contact it. It returns the peer ID of the remote peer.
func addAddrToPeerstore(h host.Host, addr string) peer.ID {
	// The following code extracts target's the peer ID from the
	// given multiaddress
	ipfsaddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		log.Fatalln(err)
	}
	pid, err := ipfsaddr.ValueForProtocol(ma.P_P2P)
	if err != nil {
		log.Fatalln(err)
	}

	peerid, err := peer.Decode(pid)
	if err != nil {
		log.Fatalln(err)
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s", peerid))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	// We have a peer ID and a targetAddr, so we add
	// it to the peerstore so LibP2P knows how to contact it
	h.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)
	return peerid
}

const help = `
This example creates a simple HTTP Proxy using two libp2p peers. The first peer
provides an HTTP server locally which tunnels the HTTP requests with libp2p
to a remote peer. The remote peer performs the requests and 
send the sends the response back.

Usage: Start remote peer first with:   ./proxy
       Then start the local peer with: ./proxy -d <remote-peer-multiaddress>

Then you can do something like: curl -x "localhost:9900" "http://ipfs.io".
This proxies sends the request through the local peer, which proxies it to
the remote peer, which makes it and sends the response back.`

func main() {
	flag.Usage = func() {
		fmt.Println(help)
		flag.PrintDefaults()
	}

	// Parse some flags
	destPeer := flag.String("d", "", "destination peer address")
	port := flag.Int("p", 9900, "proxy port")
	p2pport := flag.Int("l", 12000, "libp2p listen port")
	flag.Parse()

	// If we have a destination peer we will start a local server
	if *destPeer != "" {
		// We use p2pport+1 in order to not collide if the user
		// is running the remote peer locally on that port
		host := makeRandomHost(*p2pport + 1)
		// Make sure our host knows how to reach destPeer
		destPeerID := addAddrToPeerstore(host, *destPeer)
		proxyAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", *port))
		if err != nil {
			log.Fatalln(err)
		}
		// Create the proxy service and start the http server
		proxy := NewProxyService(host, proxyAddr, destPeerID)
		go proxy.ServeConnect()
		proxy.Serve() // serve hangs forever
	} else {
		host := makeRandomHost(*p2pport)
		// In this case we only need to make sure our host
		// knows how to handle incoming proxied requests from
		// another peer.
		_ = NewProxyService(host, nil, "")
		select {} // hang forever
	}

}
