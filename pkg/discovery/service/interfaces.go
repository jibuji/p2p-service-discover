package service

import (
	"context"
	"log"

	srpc "github.com/jibuji/go-stream-rpc"
	stream "github.com/jibuji/go-stream-rpc/stream/libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// ServiceHandler represents a service implementation
type ServiceHandler interface {
	// Protocol returns the protocol ID for this service
	Protocol() string

	// HandleStream is called when a new stream is received
	HandleStream(stream network.Stream)
}

// RPCService represents a service that can be registered with an RPC peer
type RPCService interface {
	// RegisterWithPeer registers the service with the given RPC peer
	RegisterWithPeer(peer *srpc.RpcPeer)
}

// ServiceRegistry manages service registration and client creation
type ServiceRegistry interface {
	// RegisterService registers a service handler
	RegisterService(handler ServiceHandler) error

	// NewClient creates a client for the given service and peer
	NewClient(ctx context.Context, protocol string, peer peer.ID) (interface{}, error)

	// RegisterClientConstructor registers a constructor function for creating service clients
	RegisterClientConstructor(protocol string, constructor func(*srpc.RpcPeer) interface{})
}

// BaseService provides common stream handling functionality
type BaseService struct {
	protocolID string
	service    RPCService
}

// NewBaseService creates a new base service
func NewBaseService(protocolID string, service RPCService) *BaseService {
	return &BaseService{
		protocolID: protocolID,
		service:    service,
	}
}

func (b *BaseService) Protocol() string {
	return b.protocolID
}

// HandleStream implements the common stream handling pattern
func (b *BaseService) HandleStream(s network.Stream) {
	peer := srpc.NewRpcPeer(stream.NewLibP2PStream(s))
	defer peer.Close()

	b.service.RegisterWithPeer(peer)

	// Handle stream closure
	done := make(chan struct{})
	peer.OnStreamClose(func(err error) {
		if err != nil {
			log.Printf("Stream error: %v\n", err)
		}
		close(done)
	})
	<-done
}
