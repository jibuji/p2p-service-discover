package discovery

import (
	"context"

	"github.com/jibuji/p2p-service-discover/internal/protocol/proto"
	"github.com/libp2p/go-libp2p/core/peer"
)

const PeerExchangeProtocolID = "/peer-exchange/1.0.1"

// NewPeerExchangeClient creates a client for peer exchange
func NewPeerExchangeClient(node *ServiceNode, ctx context.Context, targetPeer peer.ID) (*proto.ServicePeerClient, error) {
	client, err := node.NewServiceClient(ctx, PeerExchangeProtocolID, targetPeer)
	if err != nil {
		return nil, err
	}
	return client.(*proto.ServicePeerClient), nil
}
