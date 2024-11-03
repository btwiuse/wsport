build:
	go build -v -o /tmp/ ./basic
	go build -v -o /tmp/ ./bootstrap
	go build -v -o /tmp/ ./circuit_relay_hop
	go build -v -o /tmp/ ./dial
	go build -v -o /tmp/ ./host
	go build -v -o /tmp/ ./http
	go build -v -o /tmp/ ./peerid
	go build -v -o /tmp/ ./pid_seed
	go build -v -o /tmp/ ./proxy
	go build -v -o /tmp/ ./reconnect

patch:
	# patch for termux
	sudo sed -i -e 's,\tlog.Errorw,\t// log.Errorw,g' ~/go/pkg/mod/github.com/libp2p/go-libp2p\@v0.37.0/p2p/host/basic/basic_host.go
	cat ~/go/pkg/mod/github.com/libp2p/go-libp2p\@v0.37.0/p2p/host/basic/basic_host.go | grep log.Errorw
bas:
	# Create a libp2p.Host and print the peer id
	/tmp/basic

pid:
	# Create a libp2p.Host with a random peer id
	/tmp/pid_seed
	# Create a libp2p.Host with a custom peer id
	# Should see 12D3KooWPi4YSQCQrgFregoGHGidNKcgdGp3s4GruuZiLEkeHVDK
	PID_SEED=42 /tmp/pid_seed

hos:
	# A simple libp2p host with websocket transport and p2pid
	# - /ipfs/ping/1.0.0
	# - /libp2p/circuit/relay/0.2.0/stop
	# - /ipfs/id/1.0.0
	# - /ipfs/id/push/1.0.0
	# Note: listener will not reconnect when internet reconnects
	# see reconnect for that
	PID_SEED=host /tmp/host

pee:
	# Relay with peer discovery, with persistent peerid from env seed (PID_SEED)
	# - /ipfs/kad/1.0.0
	# Should see 12D3KooWSXdVD6y6zg28gXAnEU1CyofLSgdKbxN6z4ShjUj4XuYs
	PID_SEED=2 /tmp/peerid

hop:
	# Start a libp2p circuit v2 relay hop
	# - /libp2p/circuit/relay/0.2.0/hop
	PID_SEED=relay /tmp/circuit_relay_hop

rec:
	# Add event listeners to a host (Notifiers)
	# - OnListenClose: Reconnect when listener is closed
	# - OnListen: show listening addrs
	# - OnConnect:  show connected peer info
	# - OnDisconnect: show disconnected peer info
	#
	# Note: If relay address has no path, each reconnect will have a new subdomain
	/tmp/reconnect

htt:
	# libp2p HTTP server over websocket transport
	# - /http/1.1
	/tmp/http

pro:
	# libp2p gost proxy server
	/tmp/proxy

cli:
	# libp2p gost client
	$(eval ARG1 := $(filter-out $@,$(MAKECMDGOALS)))
	/tmp/proxy -d $(ARG1)

dia:
	# dial other peers and establish a connection
	$(eval ARG1 := $(filter-out $@,$(MAKECMDGOALS)))
	/tmp/dial $(ARG1)

boot:
	# keep connections alive to a list of bootstrap peers
	PID_SEED=42 /tmp/bootstrap

boo:
	# keep connections alive to a list of bootstrap peers
	$(eval ARG1 := $(filter-out $@,$(MAKECMDGOALS)))
	/tmp/bootstrap $(ARG1)

vara:
	# start a fake vara node
	# gear --chain vara --tmp --bootnodes /dns/example.com/tcp/443/wss/p2p/12D3KooWAFZAGPKdYL597PJcLtK2eMeBphZc8a9iZpAwFgwh87ph
	DOMAIN=/vara PID_SEED=vara DHT=/fe1b4c55fd4d668101126434206571a7838a8b6b93a6d1b95d607e78e6c53763/kad /tmp/bootstrap

dot:
	# start a fake dot node
	# polkadot --chain polkadot --tmp --bootnodes /dns/example.com/tcp/443/wss/p2p/12D3KooWS1x9zSqNtK5zSRzMCtihYvG6pSBzVwuaFfrgfMA79iRs
	DOMAIN=/dot PID_SEED=dot DHT=/91b171bb158e2d3848fa23a9f1c25182fb8e20313b2c1eb49219da7a70ce90c3/kad /tmp/bootstrap
