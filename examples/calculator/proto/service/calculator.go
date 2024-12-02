package service

import (
	"context"

	srpc "github.com/jibuji/go-stream-rpc"
	"github.com/jibuji/p2p-service-discover/examples/calculator/proto"
	baseservice "github.com/jibuji/p2p-service-discover/pkg/discovery/service"
)

const CalculatorProtocolID = "/calculator/1.0.0"

type CalculatorService struct {
	proto.UnimplementedCalculatorServer
	*baseservice.BaseService
}

func NewCalculatorService() *CalculatorService {
	svc := &CalculatorService{}
	svc.BaseService = baseservice.NewBaseService(CalculatorProtocolID, svc)
	return svc
}

// RegisterWithPeer implements RPCService interface
func (s *CalculatorService) RegisterWithPeer(peer *srpc.RpcPeer) {
	proto.RegisterCalculatorServer(peer, s)
}

func (s *CalculatorService) Add(ctx context.Context, req *proto.AddRequest) *proto.AddResponse {
	result := req.A + req.B
	return &proto.AddResponse{Result: result}
}

func (s *CalculatorService) Multiply(ctx context.Context, req *proto.MultiplyRequest) *proto.MultiplyResponse {
	result := req.A * req.B
	return &proto.MultiplyResponse{Result: result}
}

