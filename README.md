# P2P Service Discovery and RPC Library

A flexible and extensible peer-to-peer service discovery and RPC library built on libp2p, designed to make it easy to build distributed applications with service discovery capabilities.

## Features

- **Service Discovery**
  - DHT-based service discovery
  - PubSub-based real-time peer announcements
  - Automatic peer exchange protocol
  - Configurable peer TTL
  - Multi-protocol support

- **Service Registration**
  - Easy service registration with protocol versioning
  - Stream-based RPC using go-stream-rpc
  - Support for custom service handlers
  - Automatic protocol negotiation

- **Client Capabilities**
  - Dynamic service client creation
  - Automatic peer connection management
  - Support for multiple service protocols
  - Easy-to-use client interfaces

## Installation

```bash
go get github.com/jibuji/p2p-service-discover
```

## Quick Start

Here's a simple example of creating a calculator service:

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/jibuji/p2p-service-discover/pkg/discovery"
    "github.com/libp2p/go-libp2p"
)

func main() {
    ctx := context.Background()
    
    // Create libp2p host
    host, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
    if err != nil {
        log.Fatal(err)
    }

    // Create service node with default config
    config := discovery.DefaultConfig()
    node, err := discovery.NewServiceNode(ctx, host, *config)
    if err != nil {
        log.Fatal(err)
    }

    // Register calculator service
    calcService := NewCalculatorService()
    if err := node.RegisterServiceHandler(calcService); err != nil {
        log.Fatal(err)
    }

    // Register client constructor
    node.Registry().RegisterClientConstructor(
        CalculatorProtocolID,
        func(peer *rpc.RpcPeer) interface{} {
            return proto.NewCalculatorClient(peer)
        },
    )

    // Keep the node running
    select {}
}
```

## Creating a Service

1. Define your service protocol using Protocol Buffers:

```protobuf
syntax = "proto3";

package calculator;

service Calculator {
    rpc Add(AddRequest) returns (AddResponse);
    rpc Multiply(MultiplyRequest) returns (MultiplyResponse);
}

message AddRequest {
    int32 a = 1;
    int32 b = 2;
}

message AddResponse {
    int32 result = 1;
}
```

2. Implement your service:

```go
type CalculatorService struct {
    proto.UnimplementedCalculatorServer
    *baseservice.BaseService
}

func NewCalculatorService() *CalculatorService {
    svc := &CalculatorService{}
    svc.BaseService = baseservice.NewBaseService("/calculator/1.0.0", svc)
    return svc
}

func (s *CalculatorService) Add(ctx context.Context, req *proto.AddRequest) *proto.AddResponse {
    result := req.A + req.B
    return &proto.AddResponse{Result: result}
}
```

## Using a Service

```go
// Create client node
clientNode, err := discovery.NewServiceNode(ctx, clientHost, *config)
if err != nil {
    log.Fatal(err)
}

// Connect to service provider
client, err := clientNode.NewServiceClient(ctx, "/calculator/1.0.0", providerPeerID)
if err != nil {
    log.Fatal(err)
}

// Use the service
calcClient := client.(*proto.CalculatorClient)
response := calcClient.Add(&proto.AddRequest{A: 5, B: 3})
fmt.Printf("5 + 3 = %d\n", response.Result)
```

## Configuration

The library provides flexible configuration options:

```go
config := discovery.DefaultConfig()

// Customize configuration
config.EnableDHT = true
config.EnablePubSub = true
config.EnablePeerExchange = true
config.PeerTTL = 3 * time.Hour

// Or use functional options
config = discovery.DefaultConfig(
    discovery.WithDHT(true),
    discovery.WithPubSub(true),
    discovery.WithPeerTTL(3 * time.Hour),
)
```

## Examples

Check out the [examples](examples/) directory for complete working examples:

- [Calculator Service](examples/calculator/): A simple calculator service demonstration
- [Simple Discovery](examples/simple/): Basic peer discovery example

## Documentation

For detailed documentation, please visit our [documentation](docs/) directory:

- [API Reference](docs/api.md)
- [Architecture](docs/architecture.md)
- [Service Protocol](docs/service-protocol.md)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.