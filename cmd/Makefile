bas:
	# Create a libp2p.Host and print the peer id
	go run ./basic

pid:
	# Create a libp2p.Host with a random peer id
	go run ./pid_seed
	# Create a libp2p.Host with a custom peer id
	# Should see 12D3KooWPi4YSQCQrgFregoGHGidNKcgdGp3s4GruuZiLEkeHVDK
	PID_SEED=42 go run ./pid_seed

hos:
	# A simple libp2p host with websocket transport and p2pid
	# - /ipfs/ping/1.0.0
	# - /libp2p/circuit/relay/0.2.0/stop
	# - /ipfs/id/1.0.0
	# - /ipfs/id/push/1.0.0
	PID_SEED=host go run ./host

pee:
	# Relay with peer discovery, with persistent peerid from env seed (PID_SEED)
	# - /ipfs/kad/1.0.0
	# Should see 12D3KooWSXdVD6y6zg28gXAnEU1CyofLSgdKbxN6z4ShjUj4XuYs
	PID_SEED=2 go run ./peerid

hop:
	# Start a libp2p circuit v2 relay hop
	# - /libp2p/circuit/relay/0.2.0/hop
	PID_SEED=relay go run ./circuit_relay_hop

rec:
	# Add event listeners to a host (Notifiers)
	# - OnListenClose: Reconnect when listener is closed
	# - OnListen: show listening addrs
	# - OnConnect:  show connected peer info
	# - OnDisconnect: show disconnected peer info
	#
	# Note: If relay address has no path, each reconnect will have a new subdomain
	go run ./reconnect

htt:
	# libp2p HTTP server over websocket transport
	# - /http/1.1
	go run ./http

pro:
	# libp2p gost proxy server
	go run ./proxy

cli:
	# libp2p gost client
	$(eval ARG1 := $(filter-out $@,$(MAKECMDGOALS)))
	go run ./proxy -d $(ARG1)

dia:
	# dial other peers and establish a connection
	go run ./dial

boo:
	# keep connections alive to a list of bootstrap peers
	go run ./bootstrap
