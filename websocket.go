package wsport

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/transport"

	"github.com/btwiuse/wsdial"
	wsx "github.com/btwiuse/x-parity-wss"
	ma "github.com/multiformats/go-multiaddr"
	mafmt "github.com/multiformats/go-multiaddr-fmt"
	manet "github.com/multiformats/go-multiaddr/net"
)

var dialMatcher = mafmt.And(
	mafmt.Or(mafmt.IP, mafmt.DNS),
	mafmt.Base(ma.P_TCP),
	mafmt.Or(
		mafmt.Base(ma.P_WS),
		mafmt.Base(wsx.P_WS_WITH_PATH),
		mafmt.And(
			mafmt.Or(
				mafmt.And(
					mafmt.Base(ma.P_TLS),
					mafmt.Base(ma.P_SNI)),
				mafmt.Base(ma.P_TLS),
			),
			mafmt.Base(ma.P_WS)),
		mafmt.And(
			mafmt.Or(
				mafmt.And(
					mafmt.Base(ma.P_TLS),
					mafmt.Base(ma.P_SNI)),
				mafmt.Base(ma.P_TLS),
			),
			mafmt.Base(wsx.P_WS_WITH_PATH)),
		mafmt.Base(wsx.P_WSS_WITH_PATH),
		mafmt.Base(ma.P_WSS)))

var (
	wssComponent   = ma.StringCast("/wss")
	tlsWsComponent = ma.StringCast("/tls/ws")
	tlsComponent   = ma.StringCast("/tls")
	wsComponent    = ma.StringCast("/ws")
)

func init() {
	manet.RegisterFromNetAddr(ParseWebsocketNetAddr, "websocket")
	manet.RegisterToNetAddr(ConvertWebsocketMultiaddrToNetAddr, "ws")
	manet.RegisterToNetAddr(ConvertWebsocketMultiaddrToNetAddr, "wss")
	manet.RegisterToNetAddr(ConvertWebsocketMultiaddrToNetAddr, "x-parity-ws")
	manet.RegisterToNetAddr(ConvertWebsocketMultiaddrToNetAddr, "x-parity-wss")
}

type Option func(*WebsocketTransport) error

// WithTLSClientConfig sets a TLS client configuration on the WebSocket Dialer. Only
// relevant for non-browser usages.
//
// Some useful use cases include setting InsecureSkipVerify to `true`, or
// setting user-defined trusted CA certificates.
func WithTLSClientConfig(c *tls.Config) Option {
	return func(t *WebsocketTransport) error {
		t.tlsClientConf = c
		return nil
	}
}

// WithTLSConfig sets a TLS configuration for the WebSocket listener.
func WithTLSConfig(conf *tls.Config) Option {
	return func(t *WebsocketTransport) error {
		t.tlsConf = conf
		return nil
	}
}

// Mux is any type that accepts http.Handler registration by pattern.
// http.ServeMux implements this interface.
type Mux interface {
	Handle(pattern string, handler http.Handler)
}

// WithMux is a convenience option that calls WebSocketHandler and registers
// the returned handler on the given Mux at the given path.
func WithMux(mux Mux, path string) Option {
	return func(t *WebsocketTransport) error {
		mux.Handle(path, t.WebSocketHandler())
		return nil
	}
}

// WebsocketTransport is the actual go-libp2p transport
type WebsocketTransport struct {
	upgrader transport.Upgrader
	rcmgr    network.ResourceManager

	tlsClientConf *tls.Config
	tlsConf       *tls.Config

	handlerLn       *muxListener
	upgradeListener transport.Listener
}

var _ transport.Transport = (*WebsocketTransport)(nil)

func New(u transport.Upgrader, rcmgr network.ResourceManager, opts ...Option) (*WebsocketTransport, error) {
	if rcmgr == nil {
		rcmgr = &network.NullResourceManager{}
	}
	t := &WebsocketTransport{
		upgrader:      u,
		rcmgr:         rcmgr,
		tlsClientConf: &tls.Config{},
	}
	for _, opt := range opts {
		if err := opt(t); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (t *WebsocketTransport) CanDial(a ma.Multiaddr) bool {
	return dialMatcher.Matches(a)
}

func (t *WebsocketTransport) Protocols() []int {
	return []int{ma.P_WS, ma.P_WSS, wsx.P_WS_WITH_PATH, wsx.P_WSS_WITH_PATH}
}

func (t *WebsocketTransport) Proxy() bool {
	return false
}

func (t *WebsocketTransport) SkipResolve(_ context.Context, _ ma.Multiaddr) bool {
	return true
}

func (t *WebsocketTransport) Resolve(_ context.Context, maddr ma.Multiaddr) ([]ma.Multiaddr, error) {
	parsed, err := parseWebsocketMultiaddr(maddr)
	if err != nil {
		return nil, err
	}

	if !parsed.isWSS {
		// No /tls/ws component, this isn't a secure websocket multiaddr. We can just return it here
		return []ma.Multiaddr{maddr}, nil
	}

	if parsed.sni == nil {
		var err error
		// We don't have an sni component, we'll use dns/dnsaddr
		ma.ForEach(parsed.restMultiaddr, func(c ma.Component) bool {
			switch c.Protocol().Code {
			case ma.P_DNS, ma.P_DNS4, ma.P_DNS6:
				// err shouldn't happen since this means we couldn't parse a dns hostname for an sni value.
				parsed.sni, err = ma.NewComponent("sni", c.Value())
				return false
			}
			return true
		})
		if err != nil {
			return nil, err
		}
	}

	if parsed.sni == nil {
		// we didn't find anything to set the sni with. So we just return the given multiaddr
		return []ma.Multiaddr{maddr}, nil
	}

	return []ma.Multiaddr{parsed.toMultiaddr()}, nil
}

func (t *WebsocketTransport) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (transport.CapableConn, error) {
	connScope, err := t.rcmgr.OpenConnection(network.DirOutbound, true, raddr)
	if err != nil {
		return nil, err
	}
	c, err := t.dialWithScope(ctx, raddr, p, connScope)
	if err != nil {
		connScope.Done()
		return nil, err
	}
	return c, nil
}

func (t *WebsocketTransport) dialWithScope(ctx context.Context, raddr ma.Multiaddr, p peer.ID, connScope network.ConnManagementScope) (transport.CapableConn, error) {
	macon, err := t.maDial(ctx, raddr)
	if err != nil {
		return nil, err
	}
	conn, err := t.upgrader.Upgrade(ctx, t, macon, network.DirOutbound, p, connScope)
	if err != nil {
		return nil, err
	}
	return &capableConn{CapableConn: conn}, nil
}

func (t *WebsocketTransport) maDial(ctx context.Context, raddr ma.Multiaddr) (manet.Conn, error) {
	wsurl, err := parseMultiaddr(raddr)
	if err != nil {
		return nil, err
	}
	isWss := wsurl.Scheme == "wss"

	sni, err := raddr.ValueForProtocol(ma.P_SNI)
	if err == nil {
		wsurl.Host = sni + ":" + wsurl.Port()
	}

	wsconn, err := wsdial.Dial(ctx, wsurl, nil)
	if err != nil {
		return nil, err
	}

	mnc := &MyConn{Conn: wsconn, Secure: isWss}
	return mnc, nil
}

// WebSocketHandler returns an http.Handler that upgrades HTTP connections
// to WebSocket for use with this transport. Use it to mount the WebSocket
// upgrade handler on your own HTTP server or mux:
//
//	mux.Handle("/p2p", tpt.WebSocketHandler())
//	http.Serve(ln, mux)
//
// The upgrade pipeline starts immediately — connections work even without
// a subsequent Listen call. Listen is optional and only registers the
// listener address for libp2p peer discovery.
func (t *WebsocketTransport) WebSocketHandler() http.Handler {
	ln := newMuxListener()
	t.handlerLn = ln
	if t.upgrader != nil {
		t.upgradeListener = t.upgrader.UpgradeListener(t, ln)
	}
	return ln
}

func (t *WebsocketTransport) maListen(a ma.Multiaddr) (manet.Listener, error) {
	if t.handlerLn != nil {
		if err := t.handlerLn.updateAddr(a); err != nil {
			return nil, err
		}
		return t.handlerLn, nil
	}
	l, err := newListener(a, t.tlsConf)
	if err != nil {
		return nil, err
	}
	go l.serve()
	return l, nil
}

func (t *WebsocketTransport) Listen(a ma.Multiaddr) (transport.Listener, error) {
	if t.handlerLn != nil {
		// Handler mode: return a listener scoped to this address so swarm
		// can track multiple Listen calls independently.
		return &addrListener{Listener: t.upgradeListener, addr: a}, nil
	}
	malist, err := t.maListen(a)
	if err != nil {
		return nil, err
	}
	return &transportListener{Listener: t.upgrader.UpgradeListener(t, malist)}, nil
}
