# Service Protocol Documentation

## Overview

This document describes how to implement and use services in the P2P Service Discovery library. The library uses Protocol Buffers and go-stream-rpc for service definitions and RPC communication.

## Service Definition

### 1. Protocol Buffer Definition

Services are defined using Protocol Buffers. Here's an example:

```protobuf
syntax = "proto3";

package myservice;

service MyService {
    rpc DoSomething(Request) returns (Response);
    rpc StreamData(StreamRequest) returns (stream StreamResponse);
}

message Request {
    string data = 1;
}

message Response {
    string result = 1;
}
```

### 2. Service Implementation

Services must implement both the generated interface and the ServiceHandler interface:

```go
type MyService struct {
    proto.UnimplementedMyServiceServer
    *baseservice.BaseService
}

func NewMyService() *MyService {
    svc := &MyService{}
    svc.BaseService = baseservice.NewBaseService("/myservice/1.0.0", svc)
    return svc
}

// Implement RPCService interface
func (s *MyService) RegisterWithPeer(peer *rpc.RpcPeer) {
    proto.RegisterMyServiceServer(peer, s)
}

// Implement service methods
func (s *MyService) DoSomething(ctx context.Context, req *proto.Request) *proto.Response {
    // Implementation
    return &proto.Response{Result: "Done"}
}
```

## Protocol Versioning

### Version Format

Protocol IDs should follow this format:
```
/{service-name}/{major-version}.{minor-version}.{patch}
```

Example:
```go
const ProtocolID = "/myservice/1.0.0"
```

### Version Compatibility

- Major version changes indicate breaking changes
- Minor version changes add backward-compatible features
- Patch version changes fix bugs without API changes

## Service Registration

### 1. Register Service Handler

```go
// Create service instance
myService := NewMyService()

// Register with node
err := node.RegisterServiceHandler(myService)
if err != nil {
    log.Fatal(err)
}
```

### 2. Register Client Constructor

```go
node.Registry().RegisterClientConstructor(
    MyServiceProtocolID,
    func(peer *rpc.RpcPeer) interface{} {
        return proto.NewMyServiceClient(peer)
    },
)
```

## Using Services

### 1. Finding Service Providers

```go
// Find peers providing the service
peers, err := node.FindPeers("/myservice/1.0.0")
if err != nil {
    log.Fatal(err)
}

for _, peer := range peers {
    fmt.Printf("Found provider: %s\n", peer.ID)
}
```

### 2. Creating Service Client

```go
// Create client for a specific peer
client, err := node.NewServiceClient(ctx, "/myservice/1.0.0", peerID)
if err != nil {
    log.Fatal(err)
}

// Cast to specific service client
myClient := client.(*proto.MyServiceClient)

// Use the service
response := myClient.DoSomething(&proto.Request{Data: "hello"})
```

## Stream Handling

### 1. Server-side Streaming

```go
func (s *MyService) StreamData(req *proto.StreamRequest, stream proto.MyService_StreamDataServer) error {
    for {
        select {
        case <-stream.Context().Done():
            return nil
        default:
            if err := stream.Send(&proto.StreamResponse{
                Data: "data",
            }); err != nil {
                return err
            }
        }
    }
}
```

### 2. Client-side Streaming

```go
stream, err := client.StreamData(ctx, &proto.StreamRequest{})
if err != nil {
    log.Fatal(err)
}

for {
    resp, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Received: %s\n", resp.Data)
}
```

## Error Handling

### 1. Service Errors

```go
// Define service-specific errors
var (
    ErrInvalidInput = errors.New("invalid input")
    ErrNotFound     = errors.New("resource not found")
)

// Use in service implementation
func (s *MyService) DoSomething(ctx context.Context, req *proto.Request) *proto.Response {
    if req.Data == "" {
        return nil, ErrInvalidInput
    }
    // Implementation
}
```

### 2. Client Error Handling

```go
response, err := client.DoSomething(&proto.Request{Data: ""})
if err != nil {
    switch err {
    case ErrInvalidInput:
        // Handle invalid input
    case ErrNotFound:
        // Handle not found
    default:
        // Handle other errors
    }
}
```

## Best Practices

1. Protocol Design
   - Use semantic versioning
   - Keep messages focused and small
   - Consider backward compatibility

2. Implementation
   - Implement proper context handling
   - Use appropriate timeouts
   - Handle stream cleanup

3. Error Handling
   - Define clear error types
   - Provide meaningful error messages
   - Handle network errors gracefully

4. Testing
   - Test service implementation
   - Test client usage
   - Test error conditions