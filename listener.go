package wsport

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/libp2p/go-libp2p/core/transport"
	"github.com/webteleport/webteleport/transport/websocket"
	"github.com/webteleport/utils"

	"github.com/btwiuse/wsconn"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

type listener struct {
	nl net.Listener
	// The Go standard library sets the http.Server.TLSConfig no matter if this is a WS or WSS,
	// so we can't rely on checking if server.TLSConfig is set.
	isWss bool

	laddr ma.Multiaddr

	closed   chan struct{}
	incoming chan *ConnAddr
}

type ConnAddr struct {
	Conn net.Conn
	Addr string
}

func (pwma *parsedWebsocketMultiaddr) toMultiaddr() ma.Multiaddr {
	if !pwma.isWSS {
		return pwma.restMultiaddr.Encapsulate(wsComponent)
	}

	if pwma.sni == nil {
		return pwma.restMultiaddr.Encapsulate(tlsComponent).Encapsulate(wsComponent)
	}

	return pwma.restMultiaddr.Encapsulate(tlsComponent).Encapsulate(pwma.sni).Encapsulate(wsComponent)
}

// newListener creates a new listener from a raw net.Listener.
// tlsConf may be nil (for unencrypted websockets).
func newListener(a ma.Multiaddr, tlsConf *tls.Config) (*listener, error) {
	parsed, err := parseWebsocketMultiaddr(a)
	if err != nil {
		return nil, err
	}

	_, lnaddr, err := manet.DialArgs(parsed.restMultiaddr)
	if err != nil {
		return nil, err
	}
	scheme := "ws"
	if parsed.isWSS {
		scheme = "wss"
	}

	var nl net.Listener
	var localListen bool = strings.HasPrefix(lnaddr, "0.0.0.0") || strings.HasPrefix(lnaddr, "[::]") || strings.HasPrefix(lnaddr, "127.0.0.1") || strings.HasPrefix(lnaddr, "::1") || strings.HasPrefix(lnaddr, "localhost")

	if localListen {
		// listen locally
		nl, err = net.Listen("tcp", lnaddr)
		if err != nil {
			return nil, err
		}
	} else {
		// dial remote
		relayAddr := fmt.Sprintf("%s://%s%s?x-websocket-upgrade=1", scheme, lnaddr, parsed.path)
		nl, err = websocket.Listen(context.Background(), relayAddr)
		if err != nil {
			return nil, err
		}
	}

	lu, err := netAddr2URL(nl.Addr())
	if err != nil {
		return nil, err
	}

	laddr, err := url2Multiaddr(lu)
	if err != nil {
		return nil, err
	}

	first, _ := ma.SplitFirst(a)
	// Don't resolve dns addresses.
	// We want to be able to announce domain names, so the peer can validate the TLS certificate.
	if c := first.Protocol().Code; (c == ma.P_DNS || c == ma.P_DNS4 || c == ma.P_DNS6 || c == ma.P_DNSADDR) && first.Value() == "localhost" {
		_, last := ma.SplitFirst(laddr)
		laddr = first.Encapsulate(last)
	}

	ln := &listener{
		nl:       nl,
		isWss:    parsed.isWSS,
		laddr:    laddr,
		incoming: make(chan *ConnAddr),
		closed:   make(chan struct{}),
	}
	return ln, nil
}

func netAddr2URL(addr net.Addr) (*url.URL, error) {
	if addr.Network() == "tcp" {
		return url.Parse("http://" + addr.String())
	}
	return url.Parse(addr.Network() + "://" + addr.String())
}

// deprecated: superceded by FromURL
func url2Multiaddr(u *url.URL) (ma.Multiaddr, error) {
	return FromURL(u)
}

func (l *listener) serve() {
	http.Serve(l.nl, l)
	close(l.closed)
}

func (l *listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	realIP := utils.RealIP(r)
	c, err := wsconn.Wrconn(w, r)
	if err != nil {
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	select {
	case l.incoming <- &ConnAddr{c, realIP}:
	case <-l.closed:
		c.Close()
	}
	// The connection has been hijacked, it's safe to return.
}

type MyConn struct {
	net.Conn
	Addr   string
	Secure bool
}

func (c *MyConn) LocalMultiaddr() ma.Multiaddr {
	if !c.Secure {
		return ma.StringCast("/ip4/127.0.0.1/tcp/418/ws")
	}
	return ma.StringCast("/ip4/127.0.0.1/tcp/418/wss")
}

// fix websocket/unknown-unknown
func (c *MyConn) RemoteMultiaddr() ma.Multiaddr {
	addr := "127.0.0.1"
	if c.Addr != "" {
		addr = c.Addr
	}
	if !c.Secure {
		maddr := fmt.Sprintf("/ip4/%s/tcp/404/ws", addr)
		return ma.StringCast(maddr)
	}
	maddr := fmt.Sprintf("/ip4/%s/tcp/404/wss", addr)
	return ma.StringCast(maddr)
}

func (l *listener) Accept() (manet.Conn, error) {
	select {
	case c, ok := <-l.incoming:
		if !ok {
			return nil, transport.ErrListenerClosed
		}

		mnc := &MyConn{
			Conn:   c.Conn,
			Addr:   c.Addr,
			Secure: l.isWss,
		}
		return mnc, nil
	case <-l.closed:
		return nil, transport.ErrListenerClosed
	}
}

func (l *listener) Addr() net.Addr {
	return l.nl.Addr()
}

func (l *listener) Close() error {
	err := l.nl.Close() // cause serve() to return, closing l.closed
	<-l.closed
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "use of closed network connection") {
		return transport.ErrListenerClosed
	}
	return err
}

func (l *listener) Multiaddr() ma.Multiaddr {
	return l.laddr
}

type transportListener struct {
	transport.Listener
}

func (l *transportListener) Accept() (transport.CapableConn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &capableConn{CapableConn: conn}, nil
}
