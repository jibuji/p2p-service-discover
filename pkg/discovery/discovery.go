package discovery

import (
	"encoding/json"
	"log"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/multiformats/go-multiaddr"

	"github.com/jibuji/p2p-service-discover/pkg/types"
)

type announcement struct {
	PeerID    string    `json:"peer_id"`
	Timestamp time.Time `json:"timestamp"`
}

func convertAddrs(addrs []multiaddr.Multiaddr) []string {
	result := make([]string, len(addrs))
	for i, addr := range addrs {
		result[i] = addr.String()
	}
	return result
}

func (n *ServiceNode) dhtDiscoveryLoop(serviceTopic string) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	routingDiscovery := routing.NewRoutingDiscovery(n.dht)

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			peers, err := routingDiscovery.FindPeers(n.ctx, serviceTopic)
			if err != nil {
				continue
			}

			n.mu.Lock()
			service := n.services[serviceTopic]
			now := time.Now()
			for p := range peers {
				if p.ID != n.host.ID() { // Don't add self
					service.Peers[p.ID] = types.PeerData{
						LastSeen: now,
						Addrs:    convertAddrs(p.Addrs),
					}
				}
			}
			n.mu.Unlock()
		}
	}
}

func (n *ServiceNode) announceLoop(serviceTopic string, topic *pubsub.Topic) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			ann := announcement{
				PeerID:    n.host.ID().String(),
				Timestamp: time.Now(),
			}

			data, err := json.Marshal(ann)
			if err != nil {
				continue
			}

			if err := topic.Publish(n.ctx, data); err != nil {
				continue
			}
		}
	}
}

func (n *ServiceNode) pubsubDiscoveryLoop(serviceTopic string, topic *pubsub.Topic) {
	sub, err := topic.Subscribe()
	if err != nil {
		return
	}
	defer sub.Cancel()

	for {
		select {
		case <-n.ctx.Done():
			return
		default:
			msg, err := sub.Next(n.ctx)
			if err != nil {
				continue
			}

			// Skip messages from self
			if msg.ReceivedFrom == n.host.ID() {
				continue
			}

			var ann announcement
			if err := json.Unmarshal(msg.Data, &ann); err != nil {
				continue
			}

			peerID, err := peer.Decode(ann.PeerID)
			if err != nil {
				continue
			}

			// Query the peerstore for known addresses
			addrs := n.host.Peerstore().Addrs(peerID)
			if len(addrs) == 0 {
				// Optionally, you can initiate a discovery mechanism or log the absence of addresses
				log.Printf("No addresses found for peer %s", peerID)
				continue
			}
			convertedAddrs := convertAddrs(addrs)

			n.mu.Lock()
			service := n.services[serviceTopic]
			service.Peers[peerID] = types.PeerData{
				LastSeen: ann.Timestamp,
				Addrs:    convertedAddrs,
			}
			n.mu.Unlock()
		}
	}
}
