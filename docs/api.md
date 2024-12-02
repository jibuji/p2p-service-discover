# API Reference

## Core Components

### ServiceNode

The main entry point for creating a P2P service node.

```go
type ServiceNode struct {
    // Contains host, DHT, PubSub and other internal components
}

// Create a new service node
func NewServiceNode(ctx context.Context, host host.Host, config Config) (*ServiceNode, error)

// Register a service handler
func (n *ServiceNode) RegisterServiceHandler(handler ServiceHandler) error

// Find peers providing a specific service
func (n *ServiceNode) FindPeers(serviceTopic string) ([]types.PeerInfo, error)

// Create a new service client
func (n *ServiceNode) NewServiceClient(ctx context.Context, protocol string, peer peer.ID) (interface{}, error)

// Access the service registry
func (n *ServiceNode) Registry() ServiceRegistry
```

### ServiceHandler

Interface for implementing service handlers.

```go
type ServiceHandler interface {
    // Get the protocol ID for this service
    Protocol() string
    
    // Handle incoming streams
    HandleStream(stream network.Stream)
}
```

### ServiceRegistry

Registry for managing services and clients.

```go
type ServiceRegistry interface {
    // Register a service handler
    RegisterService(handler ServiceHandler) error
    
    // Register a client constructor for a protocol
    RegisterClientConstructor(protocol string, constructor func(*rpc.RpcPeer) interface{})
    
    // Create a new client
    NewClient(ctx context.Context, protocol string, peer peer.ID) (interface{}, error)
}
```

### Configuration

Options for configuring the service node.

```go
type Config struct {
    EnableDHT          bool
    EnablePubSub       bool
    EnablePeerExchange bool
    PeerTTL            time.Duration
}

// Create default configuration
func DefaultConfig() *Config

// Configuration options
func WithDHT(enable bool) Option
func WithPubSub(enable bool) Option
func WithPeerTTL(ttl time.Duration) Option
func WithPeerExchange(enable bool) Option
```

## Service Implementation

### BaseService

Base implementation for services.

```go
type BaseService struct {
    protocolID string
    handler    ServiceHandler
}

// Create a new base service
func NewBaseService(protocolID string, handler ServiceHandler) *BaseService
```

### RPCService

Interface for RPC-based services.

```go
type RPCService interface {
    // Register with an RPC peer
    RegisterWithPeer(peer *rpc.RpcPeer)
}
```

## Types

### PeerInfo

Information about a peer providing a service.

```go
type PeerInfo struct {
    ID       peer.ID
    Addrs    []multiaddr.Multiaddr
    LastSeen time.Time
}
```

## Error Handling

Common errors returned by the library:

```go
var (
    ErrServiceNotFound     = errors.New("service not found")
    ErrPeerNotFound        = errors.New("peer not found")
    ErrProtocolNotSupported = errors.New("protocol not supported")
)