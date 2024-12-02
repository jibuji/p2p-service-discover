package types

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// ServiceInfo holds information about a registered service
type ServiceInfo struct {
	Topic string
	Peers map[peer.ID]PeerData
}

// PeerData holds information about a peer providing a service
type PeerData struct {
	LastSeen time.Time
	Addrs    []string
}
