package node

import (
	"fmt"
	"log"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
)

func NewNode() (host.Host, error) {
	// Create a connection manager
	connManager, err := connmgr.NewConnManager(
		100, // Lowwater
		400, // HighWater
		connmgr.WithGracePeriod(20),
	)
	if err != nil {
		log.Fatalf("failed to create connection manager: %v", err)
	}

	// Create a new libp2p Host with a random TCP port
	host, err := libp2p.New(
		libp2p.ConnectionManager(connManager),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
	)

	if err != nil {
		log.Fatalf("failed to create libp2p host: %v", err)
	}

	// Print the host's addresses
	fmt.Println("Node's multiaddresses:")
	for _, addr := range host.Addrs() {
		fmt.Printf("%s/p2p/%s\n", addr, host.ID().String())
	}

	return host, nil
}
