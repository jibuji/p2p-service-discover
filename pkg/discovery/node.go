package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"

	srpc "github.com/jibuji/go-stream-rpc"
	"github.com/jibuji/p2p-service-discover/internal/protocol/proto"
	"github.com/jibuji/p2p-service-discover/internal/protocol/proto/service/peerexchange"
	"github.com/jibuji/p2p-service-discover/pkg/discovery/service"
	"github.com/jibuji/p2p-service-discover/pkg/types"
)

type ServiceNode struct {
	host            host.Host
	dht             *dht.IpfsDHT
	pubsub          *pubsub.PubSub
	ctx             context.Context
	cancel          context.CancelFunc
	services        map[string]*types.ServiceInfo
	mu              sync.RWMutex
	peerTTL         time.Duration
	serviceRegistry service.ServiceRegistry
}

func (n *ServiceNode) initProtocols(cfg Config) error {
	// Initialize DHT if enabled
	if cfg.EnableDHT {
		kdht, err := dht.New(n.ctx, n.host)
		if err != nil {
			return err
		}
		n.dht = kdht
	}

	// Initialize PubSub if enabled
	if cfg.EnablePubSub {
		ps, err := pubsub.NewGossipSub(n.ctx, n.host)
		if err != nil {
			return err
		}
		n.pubsub = ps
	}

	// Initialize peer exchange if enabled
	if cfg.EnablePeerExchange {
		handler := peerexchange.NewHandler(n)
		if err := n.RegisterServiceHandler(handler); err != nil {
			return err
		}

		n.Registry().RegisterClientConstructor(
			peerexchange.PeerExchangeProtocolID,
			func(peer *srpc.RpcPeer) interface{} {
				return proto.NewServicePeerClient(peer)
			},
		)
	}

	return nil
}

// NewServiceNode creates a new service discovery node
func NewServiceNode(ctx context.Context, h host.Host, cfg Config) (*ServiceNode, error) {
	ctx, cancel := context.WithCancel(ctx)

	node := &ServiceNode{
		host:     h,
		ctx:      ctx,
		cancel:   cancel,
		services: make(map[string]*types.ServiceInfo),
		peerTTL:  cfg.PeerTTL,
	}

	node.serviceRegistry = service.NewRegistry(h)

	// Initialize DHT and PubSub if enabled
	if err := node.initProtocols(cfg); err != nil {
		cancel()
		return nil, err
	}

	return node, nil
}

// RegisterService registers a new service with the given topic
func (n *ServiceNode) RegisterService(serviceTopic string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if _, exists := n.services[serviceTopic]; exists {
		return fmt.Errorf("service %s already registered", serviceTopic)
	}

	service := &types.ServiceInfo{
		Topic: serviceTopic,
		Peers: make(map[peer.ID]types.PeerData),
	}

	// Setup DHT advertising if enabled
	if n.dht != nil {
		routingDiscovery := routing.NewRoutingDiscovery(n.dht)
		routingDiscovery.Advertise(n.ctx, serviceTopic)

		// Start DHT discovery loop
		go n.dhtDiscoveryLoop(serviceTopic)
	}

	// Setup PubSub if enabled
	if n.pubsub != nil {
		topic, err := n.pubsub.Join(serviceTopic)
		if err != nil {
			return fmt.Errorf("failed to join topic: %w", err)
		}

		// Start pubsub announcement and discovery routines
		go n.announceLoop(serviceTopic, topic)
		go n.pubsubDiscoveryLoop(serviceTopic, topic)
	}

	n.services[serviceTopic] = service
	return nil
}

// FindPeers returns a list of peers that provide the specified service
func (n *ServiceNode) FindPeers(serviceTopic string) ([]types.PeerInfo, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	service, ok := n.services[serviceTopic]
	if !ok {
		return nil, fmt.Errorf("service not found: %s", serviceTopic)
	}

	var peers []types.PeerInfo
	now := time.Now()
	for p, data := range service.Peers {
		if now.Sub(data.LastSeen) < n.peerTTL {
			// Get peer's multiaddresses
			peerAddrs := n.host.Peerstore().Addrs(p)
			addrStrings := make([]string, len(peerAddrs))
			for i, addr := range peerAddrs {
				addrStrings[i] = addr.String()
			}

			peers = append(peers, types.PeerInfo{
				ID:       p,
				Addrs:    addrStrings,
				LastSeen: data.LastSeen,
			})
		}
	}
	return peers, nil
}

// CheckServiceProvider verifies if a peer provides a specific service
func (n *ServiceNode) CheckServiceProvider(ctx context.Context, peerID peer.ID, serviceTopic string) (bool, error) {
	n.mu.RLock()
	service, ok := n.services[serviceTopic]
	n.mu.RUnlock()

	if !ok {
		return false, fmt.Errorf("service not found: %s", serviceTopic)
	}

	data, exists := service.Peers[peerID]
	if !exists {
		return false, nil
	}

	return time.Since(data.LastSeen) < n.peerTTL, nil
}

// Host returns the libp2p host
func (n *ServiceNode) Host() host.Host {
	return n.host
}

// Close shuts down the node and all its services
func (n *ServiceNode) Close() error {
	n.cancel()
	if n.dht != nil {
		if err := n.dht.Close(); err != nil {
			return err
		}
	}
	return n.host.Close()
}

// ListServices returns a list of all registered service topics
func (n *ServiceNode) ListServices() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	services := make([]string, 0, len(n.services))
	for topic := range n.services {
		services = append(services, topic)
	}
	return services
}

// ListPeers is an alias for FindPeers
func (n *ServiceNode) ListPeers(serviceTopic string) ([]types.PeerInfo, error) {
	return n.FindPeers(serviceTopic)
}

// RegisterServiceHandler registers a service handler and automatically registers it for discovery
func (n *ServiceNode) RegisterServiceHandler(handler service.ServiceHandler) error {
	// Register the service for discovery
	if err := n.RegisterService(handler.Protocol()); err != nil {
		return err
	}

	// Register the service handler
	return n.serviceRegistry.RegisterService(handler)
}

// NewServiceClient creates a client for a remote service
func (n *ServiceNode) NewServiceClient(ctx context.Context, protocol string, peer peer.ID) (interface{}, error) {
	return n.serviceRegistry.NewClient(ctx, protocol, peer)
}

// Registry returns the service registry
func (n *ServiceNode) Registry() service.ServiceRegistry {
	return n.serviceRegistry
}
