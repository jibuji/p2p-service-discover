package discovery

import (
	"context"

	pb "github.com/jibuji/p2p-service-discover/internal/protocol/proto"
	"github.com/jibuji/p2p-service-discover/pkg/types"
	"github.com/libp2p/go-libp2p/core/peer"
)

// ServiceDiscovery defines the core functionality for service discovery
type ServiceDiscovery interface {
	// RegisterService registers a new service with the given topic
	RegisterService(serviceTopic string) error

	// FindPeers returns a list of peers that provide the specified service
	FindPeers(serviceTopic string) ([]types.PeerInfo, error)

	// CheckServiceProvider verifies if a peer provides a specific service
	CheckServiceProvider(ctx context.Context, peerID peer.ID, serviceTopic string) (bool, error)
}

// PeerExchange defines the peer exchange protocol functionality
type PeerExchange interface {
	// FetchPeerList retrieves a paginated list of peers for a service from a remote peer
	FetchPeerList(ctx context.Context, remotePeer peer.ID, serviceTopic string, page, pageSize int32) ([]*pb.PeerInfo, error)
}
