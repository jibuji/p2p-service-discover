package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jibuji/p2p-service-discover/examples/node"
	"github.com/jibuji/p2p-service-discover/internal/protocol/proto"
	"github.com/jibuji/p2p-service-discover/pkg/discovery"
	"github.com/jibuji/p2p-service-discover/pkg/types"
	"github.com/libp2p/go-libp2p/core/peer"
)

func checkPeers(node *discovery.ServiceNode, nodeNum int, serviceTopic string) ([]types.PeerInfo, error) {
	peers, err := node.FindPeers(serviceTopic)
	if err != nil {
		log.Printf("Error listing peers for node %d: %v\n", nodeNum, err)
		return nil, err
	}
	fmt.Printf("Node %d found %d peers for service %s\n",
		nodeNum, len(peers), serviceTopic)

	// Print detailed peer information
	for _, p := range peers {
		fmt.Printf("  Peer ID: %s\n", p.ID)
		fmt.Printf("  Addresses: %v\n", p.Addrs)
		fmt.Printf("  Last seen: %s\n", p.LastSeen.Format(time.RFC3339))
	}
	return peers, nil
}

func startPeerChecking(ctx context.Context, nodes []*discovery.ServiceNode, serviceTopic string) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				fmt.Println("\n=== Checking peers for all nodes ===")
				for i, node := range nodes {
					peers, err := checkPeers(node, i+1, serviceTopic)
					if err != nil {
						log.Printf("Error checking peers for node %d: %v\n", i+1, err)
						continue
					}
					fmt.Printf("Node %d found %d peers\n", i+1, len(peers))
				}
				fmt.Println("=====================================")
			}
		}
	}()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create configuration with both DHT and PubSub enabled
	config := discovery.DefaultConfig()
	// Optionally disable peer exchange if needed:
	// config.EnablePeerExchange = false
	// or use the option:
	// config = discovery.DefaultConfig(discovery.WithPeerExchange(false))

	// create libp2p node with tcp port 0, which means the libp2p will choose a random port
	host1, _ := node.NewNode()
	// Create first node (bootstrap node)
	node1, err := discovery.NewServiceNode(ctx, host1, *config)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Node 1 (bootstrap) started with ID: %s\n", node1.Host().ID().String())

	// Create second node
	host2, _ := node.NewNode()
	node2, err := discovery.NewServiceNode(ctx, host2, *config)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Node 2 started with ID: %s\n", node2.Host().ID().String())

	// Create third node
	host3, _ := node.NewNode()
	node3, err := discovery.NewServiceNode(ctx, host3, *config)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Node 3 started with ID: %s\n", node3.Host().ID().String())

	// Connect node2 and node3 to node1 (bootstrap node)
	bootstrapInfo := peer.AddrInfo{
		ID:    node1.Host().ID(),
		Addrs: node1.Host().Addrs(),
	}

	if err := node2.Host().Connect(ctx, bootstrapInfo); err != nil {
		log.Fatal(err)
	}
	if err := node3.Host().Connect(ctx, bootstrapInfo); err != nil {
		log.Fatal(err)
	}

	// Register service on all nodes
	serviceTopic := "example-service"
	nodes := []*discovery.ServiceNode{node1, node2, node3}

	for i, node := range nodes {
		if err := node.RegisterService(serviceTopic); err != nil {
			log.Fatalf("Failed to register service on node %d: %v", i+1, err)
		}
		fmt.Printf("Service registered on node %d\n", i+1)
	}

	// Wait for initial peer discovery
	time.Sleep(5 * time.Second)

	// Create a new node to demonstrate peer fetching
	host4, _ := node.NewNode()
	node4, err := discovery.NewServiceNode(ctx, host4, *config)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nNode 4 (querying node) started with ID: %s\n", node4.Host().ID().String())

	// Connect node4 to bootstrap node
	if err := node4.Host().Connect(ctx, bootstrapInfo); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Node 4 connected to bootstrap node")

	if err := node4.RegisterService(serviceTopic); err != nil {
		log.Fatalf("Failed to register service on node %d: %v", 4, err)
	}
	time.Sleep(30 * time.Second)
	// Example of using FindPeers with multiaddrs
	peerInfos, err := node4.FindPeers(serviceTopic)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nFound peers with addresses:\n")
	for _, info := range peerInfos {
		fmt.Printf("Peer %s:\n", info.ID)
		for _, addr := range info.Addrs {
			fmt.Printf("  - %s\n", addr)
		}
	}

	go func() {
		time.Sleep(1 * time.Minute)
		// Create peer exchange client in one line
		pexClient, err := discovery.NewPeerExchangeClient(node4, ctx, bootstrapInfo.ID)
		if err != nil {
			log.Fatal(err)
		}

		// Use the client directly
		req := &proto.PeerListRequest{
			ServiceTopic: serviceTopic,
			Page:         0,
			PageSize:     10,
		}
		peers := pexClient.FetchPeerList(req)

		fmt.Printf("Fetched %d peers:\n", len(peers.Peers))

		provides := pexClient.CheckService(&proto.ServiceCheckRequest{
			ServiceTopic: serviceTopic,
		})
		fmt.Printf("Service %s provides: %v\n", serviceTopic, provides.ProvidesService)

		for _, p := range peers.Peers {
			peerID, err := peer.IDFromBytes(p.PeerId)
			if err != nil {
				continue
			}
			fmt.Printf("- Peer %s (last seen: %v)\n",
				peerID.String(),
				time.Unix(0, p.LastSeen).Format(time.RFC3339))
		}
	}()

	// Start periodic peer checking
	startPeerChecking(ctx, nodes, serviceTopic)

	// Wait for interrupt signal
	fmt.Println("\nRunning... Press Ctrl+C to exit")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
