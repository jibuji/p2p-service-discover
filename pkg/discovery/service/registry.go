package service

import (
	"context"
	"fmt"
	"sync"

	srpc "github.com/jibuji/go-stream-rpc"
	stream "github.com/jibuji/go-stream-rpc/stream/libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type registry struct {
	host host.Host
	mu   sync.RWMutex
	// Map protocol ID to client constructor function
	clientConstructors map[string]func(*srpc.RpcPeer) interface{}
}

func NewRegistry(h host.Host) ServiceRegistry {
	return &registry{
		host:               h,
		clientConstructors: make(map[string]func(*srpc.RpcPeer) interface{}),
	}
}

func (r *registry) RegisterService(handler ServiceHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	protocolID := handler.Protocol()

	// Set the stream handler
	r.host.SetStreamHandler(protocol.ID(protocolID), handler.HandleStream)

	return nil
}

func (r *registry) RegisterClientConstructor(protocol string, constructor func(*srpc.RpcPeer) interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clientConstructors[protocol] = constructor
}

func (r *registry) NewClient(ctx context.Context, ptcID string, targetPeer peer.ID) (interface{}, error) {
	r.mu.RLock()
	constructor, ok := r.clientConstructors[ptcID]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no client constructor registered for protocol: %s", ptcID)
	}

	s, err := r.host.NewStream(ctx, targetPeer, protocol.ID(ptcID))
	if err != nil {
		return nil, err
	}

	rpcPeer := srpc.NewRpcPeer(stream.NewLibP2PStream(s))
	return constructor(rpcPeer), nil
}
