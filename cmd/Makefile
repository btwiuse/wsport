BUILD_DIR := out

build:
	go build -v -o out/ ./basic
	go build -v -o out/ ./bootstrap
	go build -v -o out/ ./circuit_relay_hop
	go build -v -o out/ ./dial
	go build -v -o out/ ./host
	go build -v -o out/ ./http
	go build -v -o out/ ./peerid
	go build -v -o out/ ./pid_seed
	go build -v -o out/ ./proxy
	go build -v -o out/ ./reconnect
	go build -v -o out/ ./wsdial
	go build -v -o out/ ./wsdial2

build-wasm:
	GOOS=js GOARCH=wasm go build -v -o out/basic.wasm ./basic/
	GOOS=js GOARCH=wasm go build -v -o out/bootstrap.wasm ./bootstrap/
	GOOS=js GOARCH=wasm go build -v -o out/circuit_relay_hop.wasm ./circuit_relay_hop/
	GOOS=js GOARCH=wasm go build -v -o out/dial.wasm ./dial/
	GOOS=js GOARCH=wasm go build -v -o out/host.wasm ./host/
	GOOS=js GOARCH=wasm go build -v -o out/http.wasm ./http/
	GOOS=js GOARCH=wasm go build -v -o out/peerid.wasm ./peerid/
	GOOS=js GOARCH=wasm go build -v -o out/pid_seed.wasm ./pid_seed/
	GOOS=js GOARCH=wasm go build -v -o out/proxy.wasm ./proxy/
	GOOS=js GOARCH=wasm go build -v -o out/reconnect.wasm ./reconnect/
	GOOS=js GOARCH=wasm go build -v -o out/wsdial.wasm ./wsdial/
	GOOS=js GOARCH=wasm go build -v -o out/wsdial2.wasm ./wsdial2/

patch:
	# patch for termux
	sudo sed -i -e 's,\tlog.Errorw,\t// log.Errorw,g' ~/go/pkg/mod/github.com/libp2p/go-libp2p\@v0.37.0/p2p/host/basic/basic_host.go
	cat ~/go/pkg/mod/github.com/libp2p/go-libp2p\@v0.37.0/p2p/host/basic/basic_host.go | grep log.Errorw

wsd:
	# Compare native ws transport and wsport, useful for comparison / debugging
	# tcping -l /ip4/0.0.0.0/tcp/60003/ws
	go run ./wsdial/ /ip4/127.0.0.1/tcp/60003/ws/p2p/12D3KooWK3YY84JVi2VjvGLsNDnruqRoi8xfDVoQFpW246gSyMf1
	go run ./wsdial2/ /ip4/127.0.0.1/tcp/60003/ws/p2p/12D3KooWK3YY84JVi2VjvGLsNDnruqRoi8xfDVoQFpW246gSyMf1

bas:
	# Create a libp2p.Host and print the peer id
	./out/basic

pid:
	# Create a libp2p.Host with a random peer id
	./out/pid_seed
	# Create a libp2p.Host with a custom peer id
	# Should see 12D3KooWPi4YSQCQrgFregoGHGidNKcgdGp3s4GruuZiLEkeHVDK
	PID_SEED=42 ./out/pid_seed

hos:
	# A simple libp2p host with websocket transport and p2pid
	# - /ipfs/ping/1.0.0
	# - /libp2p/circuit/relay/0.2.0/stop
	# - /ipfs/id/1.0.0
	# - /ipfs/id/push/1.0.0
	# Note: listener will not reconnect when internet reconnects
	# see reconnect for that
	PID_SEED=host ./out/host

pee:
	# Relay with peer discovery, with persistent peerid from env seed (PID_SEED)
	# - /ipfs/kad/1.0.0
	# Should see 12D3KooWSXdVD6y6zg28gXAnEU1CyofLSgdKbxN6z4ShjUj4XuYs
	PID_SEED=2 ./out/peerid

hop:
	# Start a libp2p circuit v2 relay hop
	# - /libp2p/circuit/relay/0.2.0/hop
	PID_SEED=relay ./out/circuit_relay_hop

rec:
	# Add event listeners to a host (Notifiers)
	# - OnListenClose: Reconnect when listener is closed
	# - OnListen: show listening addrs
	# - OnConnect:  show connected peer info
	# - OnDisconnect: show disconnected peer info
	#
	# Note: If relay address has no path, each reconnect will have a new subdomain
	./out/reconnect

htt:
	# libp2p HTTP server over websocket transport
	# - /http/1.1
	./out/http

pro:
	# libp2p gost proxy server
	./out/proxy

cli:
	# libp2p gost client
	$(eval ARG1 := $(filter-out $@,$(MAKECMDGOALS)))
	./out/proxy -d $(ARG1)

dia:
	# dial other peers and establish a connection
	$(eval ARG1 := $(filter-out $@,$(MAKECMDGOALS)))
	./out/dial $(ARG1)

boot:
	# keep connections alive to a list of bootstrap peers
	PID_SEED=42 ./out/bootstrap

boo:
	# keep connections alive to a list of bootstrap peers
	$(eval ARG1 := $(filter-out $@,$(MAKECMDGOALS)))
	./out/bootstrap $(ARG1)

# https://github.com/polkadot-api/polkadot-api/tree/main/packages/known-chains/specs

vara:
	# start a fake vara node
	# gear --chain vara --tmp --bootnodes /dns/example.com/tcp/443/wss/p2p/12D3KooWAFZAGPKdYL597PJcLtK2eMeBphZc8a9iZpAwFgwh87ph
	DOMAIN=/vara PID_SEED=vara DHT=/fe1b4c55fd4d668101126434206571a7838a8b6b93a6d1b95d607e78e6c53763/kad ./out/bootstrap

dot:
	# start a fake dot node
	# polkadot --chain polkadot --tmp --bootnodes /dns/example.com/tcp/443/wss/p2p/12D3KooWS1x9zSqNtK5zSRzMCtihYvG6pSBzVwuaFfrgfMA79iRs
	DOMAIN=/dot PID_SEED=dot DHT=/91b171bb158e2d3848fa23a9f1c25182fb8e20313b2c1eb49219da7a70ce90c3/kad ./out/bootstrap

pas:
	# start a fake paseo asset hub node
	DOMAIN=/pas PID_SEED=pas DHT=/d6eec26135305a8ad257a20d003357284c8aa03d0bdb2b357ab0a22371e11ef2/kad ./out/bootstrap /dns/asset-hub-paseo-bootnode.turboflakes.io/tcp/30430/wss/p2p/12D3KooWJzfVkdDnKfn2hQ1c3ysrbmReTjVKrEBHkdwgZThbB1BM

# identify protocol
#
# /ipfs/id/1.0.0
# /ipfs/id/push/1.0.0
#
# Get remote peer advertised addresses

# constella for terminal
#
# /xterm/<pid>
