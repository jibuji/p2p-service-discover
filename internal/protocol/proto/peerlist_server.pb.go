// Code generated by stream-rpc. DO NOT EDIT.
package proto

import (
	rpc "github.com/jibuji/go-stream-rpc"
	"context"
)

// UnimplementedCalculatorServer can be embedded to have forward compatible implementations
type UnimplementedServicePeerServer struct{}

type ServicePeerServer interface {
	FetchPeerList(context.Context, *PeerListRequest) *PeerListResponse

	CheckService(context.Context, *ServiceCheckRequest) *ServiceCheckResponse
}

type ServicePeerServerImpl struct {
	impl ServicePeerServer
}

func RegisterServicePeerServer(peer *rpc.RpcPeer, impl ServicePeerServer) {
	server := &ServicePeerServerImpl{impl: impl}
	peer.RegisterService("ServicePeer", server)
}

func (s *UnimplementedServicePeerServer) FetchPeerList(ctx context.Context, req *PeerListRequest) *PeerListResponse {
	return nil
}

func (s *UnimplementedServicePeerServer) CheckService(ctx context.Context, req *ServiceCheckRequest) *ServiceCheckResponse {
	return nil
}

func (s *ServicePeerServerImpl) FetchPeerList(ctx context.Context, req *PeerListRequest) *PeerListResponse {
	return s.impl.FetchPeerList(ctx, req)
}

func (s *ServicePeerServerImpl) CheckService(ctx context.Context, req *ServiceCheckRequest) *ServiceCheckResponse {
	return s.impl.CheckService(ctx, req)
}
