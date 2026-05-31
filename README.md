# wsport — WebSocket transport for go-libp2p

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/btwiuse/wsport?tab=doc)
[![Go 1.25+](https://img.shields.io/github/go-mod/go-version/btwiuse/wsport)](https://golang.org/dl/)
[![License](https://img.shields.io/github/license/webteleport/webteleport?color=%23000&style=flat-round)](https://github.com/btwiuse/wsport/blob/main/LICENSE)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/btwiuse/wsport)

A fork of [go-ws-transport](https://github.com/libp2p/go-ws-transport) with support for mounting the WebSocket upgrade handler on an existing `http.ServeMux`.

## Usage

```go
import "github.com/btwiuse/wsport"
```

### As a libp2p transport

```go
host, _ := libp2p.New(
    libp2p.Transport(wsport.New),
)
```

### Share a port with HTTP

Mount the WebSocket handler on the same `http.ServeMux` as your API routes:

```go
mux := http.NewServeMux()
mux.HandleFunc("/api", apiHandler)

host, _ := libp2p.New(
    libp2p.Transport(wsport.New, wsport.WithMux(mux, "/p2p")),
)

ln, _ := net.Listen("tcp", ":8080")
host.Network().Listen(ma.StringCast("/ip4/0.0.0.0/tcp/8080/x-parity-ws/%2Fp2p"))
http.Serve(ln, mux)
```

### Or with `WebSocketHandler` for full control

```go
mux.Handle("/p2p", tpt.WebSocketHandler())
```

See [cmd/muxp2p](cmd/muxp2p/README.md) for a complete example.
