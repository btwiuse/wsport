package main

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/btwiuse/p2pid"
	"github.com/btwiuse/wsport"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/transport"
	"github.com/webteleport/webteleport"
)

var (
	RELAY   = cmp.Or(os.Getenv("RELAY"), "https://example.com")
	p2pPath = "/"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"server":   "muxp2p-demo",
			"version":  "0.1.0",
			"p2p_path": p2pPath,
		})
	})

	mux.HandleFunc("/api/time", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"time": time.Now().Format(time.RFC3339),
		})
	})

	// Create libp2p host with a transport that we can reference.
	var tpt *wsport.WebsocketTransport
	host, err := libp2p.New(
		p2pid.FromEnv(p2pid.PID_SEED),
		libp2p.Transport(func(u transport.Upgrader, rcmgr network.ResourceManager) (*wsport.WebsocketTransport, error) {
			var err error
			tpt, err = wsport.New(u, rcmgr)
			return tpt, err
		}),
	)
	if err != nil {
		log.Fatalf("failed to create libp2p host: %v", err)
	}
	defer host.Close()

	p2pPath = "/p2p/" + host.ID().String()

	// Mount the WebSocket handler on the shared mux.
	mux.Handle(p2pPath, tpt.WebSocketHandler())

	ln, err := webteleport.Listen(context.Background(), RELAY)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	maddr := fmt.Sprintf("%s://%s%s", ln.Addr().Network(), ln.Addr(), p2pPath)
	listenMa, err := wsport.FromString(maddr)
	if err != nil {
		log.Fatalf("failed to parse multiaddr: %v", err)
	}

	if err := host.Network().Listen(listenMa); err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("peer ID: %s", host.ID())
	log.Printf("p2p WebSocket mounted at %s", p2pPath)
	log.Println(listenMa)
	log.Printf("listening addresses: %v", host.Addrs())
	log.Println("listening on", ln.Addr())

	if err := http.Serve(ln, mux); err != nil {
		log.Fatalf("http.Serve: %v", err)
	}
}
