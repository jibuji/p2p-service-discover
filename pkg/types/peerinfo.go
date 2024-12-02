package types

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

type PeerInfo struct {
	ID       peer.ID
	Addrs    []string
	LastSeen time.Time
}
