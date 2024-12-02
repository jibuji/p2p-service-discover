package peerexchange

import (
	srpc "github.com/jibuji/go-stream-rpc"
	"github.com/jibuji/p2p-service-discover/internal/protocol/proto"
	"github.com/jibuji/p2p-service-discover/internal/protocol/proto/service"
	interfaces "github.com/jibuji/p2p-service-discover/pkg/discovery/interfaces"
	baseservice "github.com/jibuji/p2p-service-discover/pkg/discovery/service"
)

const PeerExchangeProtocolID = "/peer-exchange/1.0.1"

type Handler struct {
	*baseservice.BaseService
	node interfaces.ServiceDiscovery
}

func NewHandler(node interfaces.ServiceDiscovery) *Handler {
	h := &Handler{node: node}
	h.BaseService = baseservice.NewBaseService(PeerExchangeProtocolID, h)
	return h
}

// RegisterWithPeer implements RPCService interface
func (h *Handler) RegisterWithPeer(peer *srpc.RpcPeer) {
	proto.RegisterServicePeerServer(peer, service.NewServicePeerService(h.node))
}
