# muxp2p — Share a port between libp2p and HTTP

This example demonstrates how to mount the libp2p WebSocket transport on
a shared `http.ServeMux`, alongside regular HTTP handlers on the same port.

```
:8080
 ├─ /healthz       → {"status":"ok"}
 ├─ /api/info       → server metadata
 ├─ /api/time       → current time
 └─ /p2p/<peerID>   → libp2p WebSocket transport
```

## Usage

```bash
# Start the server
go run .

# Or with a custom relay
RELAY=https://my-relay.example.com go run .
```

The peer ID is printed at startup. Other libp2p peers can dial the node
using the multiaddr shown in the logs.

## API

### `WebSocketHandler`

`WebSocketHandler()` returns an `http.Handler` that upgrades HTTP
connections to WebSocket for the libp2p transport. Mount it anywhere on
your mux.

To use it, capture the transport reference through a custom constructor.
`upgrader` and `rcmgr` are injected by libp2p's fx — you don't construct
them yourself:

```go
import (
    "github.com/btwiuse/wsport"
    "github.com/libp2p/go-libp2p/core/network"
    "github.com/libp2p/go-libp2p/core/transport"
)

var tpt *wsport.WebsocketTransport
host, _ := libp2p.New(
    libp2p.Transport(func(u transport.Upgrader, rcmgr network.ResourceManager) (*wsport.WebsocketTransport, error) {
        var err error
        tpt, err = wsport.New(u, rcmgr)
        return tpt, err
    }),
)
mux.Handle("/p2p", tpt.WebSocketHandler())
```

The handler is backed by an internal listener that receives its address
from the subsequent `Listen` call on the transport.

**Important:** When using `WebSocketHandler` or `WithMux`, calling
`host.Network().Listen(addr)` does **not** create a `net.Listener` or
bind to the port. It only registers the address with libp2p so that it
appears in `host.Addrs()` for peer discovery. The actual HTTP server is
started separately by the caller.

Call `Listen` exactly **once** — multiple calls overwrite the listener's
reported address and start duplicate upgrade goroutines that race on the
same incoming channel.

### `WithMux`

`WithMux(mux, path)` is a convenience option that calls
`WebSocketHandler` and registers the returned handler on the mux in one
step. Useful when the transport is created by go-libp2p's fx framework:

```go
host, _ := libp2p.New(
    libp2p.Transport(wsport.New, wsport.WithMux(mux, "/p2p")),
)
```

This is equivalent to:

```go
var tpt *wsport.WebsocketTransport
host, _ := libp2p.New(
    libp2p.Transport(func(u transport.Upgrader, rcmgr network.ResourceManager) (*wsport.WebsocketTransport, error) {
        var err error
        tpt, err = wsport.New(u, rcmgr)
        return tpt, err
    }),
)
mux.Handle("/p2p", tpt.WebSocketHandler())
```

### How it works

The transport creates a `muxListener` that implements both
`http.Handler` (upgrade WebSocket requests) and `manet.Listener`
(feed upgraded connections to the libp2p upgrader). No `net.Listener`
or `http.Server` is created by the transport — the caller owns the HTTP
server entirely.

### `Listen` behavior in handler mode

When using `WebSocketHandler` or `WithMux`, calling
`host.Network().Listen(addr)` registers the address with libp2p so it
appears in `host.Addrs()` for peer discovery. It does **not** bind a
port or create a `net.Listener` — the caller owns the HTTP server.

Multiple `Listen` calls are safe. Each returns a unique listener
scoped to its address, backed by the same upgrade pipeline. Swarm
tracks them independently and distributes incoming connections
correctly.
