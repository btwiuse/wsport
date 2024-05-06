package wsport

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	_ "github.com/btwiuse/x-parity-wss"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

// Addr is an implementation of net.Addr for WebSocket.
type Addr struct {
	*url.URL
}

var _ net.Addr = (*Addr)(nil)

// Network returns the network type for a WebSocket, "websocket".
func (addr *Addr) Network() string {
	return "websocket"
}

// NewAddr creates an Addr with `ws` scheme (insecure).
//
// Deprecated. Use NewAddrWithScheme.
func NewAddr(host string) *Addr {
	// Older versions of the transport only supported insecure connections (i.e.
	// WS instead of WSS). Assume that is the case here.
	return NewAddrWithScheme(host, false)
}

// NewAddrWithScheme creates a new Addr using the given host string. isSecure
// should be true for WSS connections and false for WS.
func NewAddrWithScheme(host string, isSecure bool) *Addr {
	scheme := "ws"
	if isSecure {
		scheme = "wss"
	}
	return &Addr{
		URL: &url.URL{
			Scheme: scheme,
			Host:   host,
		},
	}
}

func ConvertWebsocketMultiaddrToNetAddr(maddr ma.Multiaddr) (net.Addr, error) {
	url, err := parseMultiaddr(maddr)
	if err != nil {
		return nil, err
	}
	return &Addr{URL: url}, nil
}

func ParseWebsocketNetAddr(a net.Addr) (ma.Multiaddr, error) {
	wsa, ok := a.(*Addr)
	if !ok {
		return nil, fmt.Errorf("not a websocket address")
	}
	return FromURL(wsa.URL)
}

func FromURL(a *url.URL) (ma.Multiaddr, error) {
	wsa := &Addr{URL: a}

	var (
		tcpma ma.Multiaddr
		err   error
		port  = wsa.Port()
		host  = wsa.Hostname()
		path  = wsa.Path
		sche  string
	)

	switch wsa.Scheme {
	case "wss", "https":
		if port == "" {
			port = "443"
		}
		if path == "" {
			sche = "/wss"
		} else {
			sche = fmt.Sprintf("/x-parity-wss/%s", url.QueryEscape(path))
		}
	case "ws", "http":
		if port == "" {
			port = "80"
		}
		if path == "" {
			sche = "/ws"
		} else {
			sche = fmt.Sprintf("/x-parity-ws/%s", url.QueryEscape(path))
		}
	default:
		return nil, fmt.Errorf("invalid scheme in url: '%q'", wsa.URL)
	}

	iport, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("failed to parse port '%q': %s", port, err)
	}

	// NOTE: Ignoring IPv6 zones...
	// Detect if host is IP address or DNS
	if ip := net.ParseIP(host); ip != nil {
		// Assume IP address
		tcpma, err = manet.FromNetAddr(&net.TCPAddr{
			IP:   ip,
			Port: iport,
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Assume DNS name
		tcpma, err = ma.NewMultiaddr(fmt.Sprintf("/dns/%s/tcp/%s", host, port))
		if err != nil {
			return nil, err
		}
	}

	wsma, err := ma.NewMultiaddr(sche)
	if err != nil {
		return nil, err
	}
	return tcpma.Encapsulate(wsma), nil
}

func parseMultiaddr(maddr ma.Multiaddr) (*url.URL, error) {
	parsed, err := parseWebsocketMultiaddr(maddr)
	if err != nil {
		return nil, err
	}

	scheme := "ws"
	if parsed.isWSS {
		scheme = "wss"
	}

	network, host, err := manet.DialArgs(parsed.restMultiaddr)
	if err != nil {
		return nil, err
	}
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil, fmt.Errorf("unsupported websocket network %s", network)
	}
	return &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   parsed.path,
	}, nil
}

type parsedWebsocketMultiaddr struct {
	isWSS bool
	// sni is the SNI value for the TLS handshake, and for setting HTTP Host header
	sni *ma.Component
	// the rest of the multiaddr before the /tls/sni/example.com/ws or /ws or /wss
	restMultiaddr ma.Multiaddr
	// the rest of the multiaddr after /x-parity-ws or /x-parity-wss
	path string
}

func parseWebsocketMultiaddr(a ma.Multiaddr) (parsedWebsocketMultiaddr, error) {
	out := parsedWebsocketMultiaddr{}

	// First check if we have a WSS component. If so we'll canonicalize it into a /tls/ws
	withoutWss := a.Decapsulate(wssComponent)
	if !withoutWss.Equal(a) {
		a = withoutWss.Encapsulate(tlsWsComponent)
	}

	// then check if we have a x-parity-wss
	withoutWss, c := ma.SplitLast(a)
	if !withoutWss.Equal(a) && c.Protocol().Name == "x-parity-wss" {
		a = withoutWss.Encapsulate(ma.StringCast(fmt.Sprintf("/tls/x-parity-ws/%s", c.Value())))
	}

	// Remove the ws component
	withoutWs, c := ma.SplitLast(a)
	if c.Protocol().Name != "ws" && c.Protocol().Name != "x-parity-ws" {
		return out, fmt.Errorf("not a websocket multiaddr")
	}
	if c.Protocol().Name == "x-parity-ws" {
		// percent decode the path
		v, err := url.QueryUnescape(c.Value())
		if err != nil {
			return out, err
		}
		out.path = v
	}

	rest := withoutWs
	// If this is not a wss then withoutWs is the rest of the multiaddr
	out.restMultiaddr = withoutWs
	for {
		var head *ma.Component
		rest, head = ma.SplitLast(rest)
		if head == nil || rest == nil {
			break
		}

		if head.Protocol().Code == ma.P_SNI {
			out.sni = head
		} else if head.Protocol().Code == ma.P_TLS {
			out.isWSS = true
			out.restMultiaddr = rest
			break
		}
	}

	return out, nil
}
