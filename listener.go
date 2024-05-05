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
	incoming chan *Conn
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

	// dial
	relayAddr := fmt.Sprintf("%s://%s%s?x-websocket-upgrade=1", scheme, lnaddr, parsed.path)
	nl, err := websocket.Listen(context.Background(), relayAddr)
	if err != nil {
		return nil, err
	}

	laddr, err := netAddr2Multiaddr(nl.Addr())
	if err != nil {
		return nil, err
	}

	ln := &listener{
		nl:       nl,
		isWss:    parsed.isWSS,
		laddr:    laddr,
		incoming: make(chan *Conn),
		closed:   make(chan struct{}),
	}
	return ln, nil
}

func netAddr2Multiaddr(addr net.Addr) (ma.Multiaddr, error) {
	u, err := url.Parse(addr.Network() + "://" + addr.String())
	if err != nil {
		return nil, err
	}

	port := u.Port()
	if port == "" {
		if u.Scheme == "ws" {
			port = "80"
		} else if u.Scheme == "wss" {
			port = "443"
		}
	}

	switch addr.Network() {
	case "ws":
		// convert to multiaddr, assume ws on 80
		if u.Path == "" {
			return ma.StringCast(fmt.Sprintf("/dns4/%s/tcp/%s/ws", u.Hostname(), port)), nil
		}

		return ma.StringCast(fmt.Sprintf("/dns4/%s/tcp/%s/x-parity-ws/%s", u.Hostname(), port, url.QueryEscape(u.Path))), nil
	case "wss":
		// convert to multiaddr, assume wss on 443
		if u.Path == "" {
			return ma.StringCast(fmt.Sprintf("/dns4/%s/tcp/%s/wss", u.Hostname(), port)), nil
		}

		return ma.StringCast(fmt.Sprintf("/dns4/%s/tcp/%s/x-parity-wss/%s", u.Hostname(), port, url.QueryEscape(u.Path))), nil
	}

	return nil, fmt.Errorf("unsupported network: %s", addr.Network())
}

func (l *listener) serve() {
	defer close(l.closed)
	http.Serve(l.nl, l)
}

func (l *listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// The upgrader writes a response for us.
		return
	}

	select {
	case l.incoming <- NewConn(c, l.isWss):
	case <-l.closed:
		c.Close()
	}
	// The connection has been hijacked, it's safe to return.
}

type MyConn struct {
	net.Conn
}

func (c *MyConn) LocalMultiaddr() ma.Multiaddr {
	return ma.StringCast("/ip4/127.0.0.1/tcp/418")
}

// fix websocket/unknown-unknown
func (c *MyConn) RemoteMultiaddr() ma.Multiaddr {
	return ma.StringCast("/ip4/127.0.0.1/tcp/404")
}

func (l *listener) Accept() (manet.Conn, error) {
	select {
	case c, ok := <-l.incoming:
		if !ok {
			return nil, transport.ErrListenerClosed
		}

		mnc := &MyConn{Conn: c}
		return mnc, nil
	case <-l.closed:
		return nil, transport.ErrListenerClosed
	}
}

func (l *listener) Addr() net.Addr {
	return l.nl.Addr()
}

func (l *listener) Close() error {
	err := l.nl.Close()
	<-l.closed
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
