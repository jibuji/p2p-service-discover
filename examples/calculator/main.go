package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	rpc "github.com/jibuji/go-stream-rpc"
	"github.com/jibuji/p2p-service-discover/examples/calculator/proto"
	"github.com/jibuji/p2p-service-discover/examples/calculator/proto/service"
	"github.com/jibuji/p2p-service-discover/pkg/discovery"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create first node (service provider)
	host1, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
	if err != nil {
		log.Fatal(err)
	}

	config := discovery.DefaultConfig()
	node1, err := discovery.NewServiceNode(ctx, host1, *config)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Calculator service node started with ID: %s\n", node1.Host().ID().String())

	// Create and register calculator service
	calcHandler := service.NewCalculatorService()
	if err := node1.RegisterServiceHandler(calcHandler); err != nil {
		log.Fatal(err)
	}

	// Register client constructor
	node1.Registry().RegisterClientConstructor(
		service.CalculatorProtocolID,
		func(peer *rpc.RpcPeer) interface{} {
			return proto.NewCalculatorClient(peer)
		},
	)

	// Create second node (client)
	host2, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
	if err != nil {
		log.Fatal(err)
	}

	node2, err := discovery.NewServiceNode(ctx, host2, *config)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Client node started with ID: %s\n", node2.Host().ID().String())

	// Connect node2 to node1
	node1Info := peer.AddrInfo{
		ID:    node1.Host().ID(),
		Addrs: node1.Host().Addrs(),
	}
	if err := node2.Host().Connect(ctx, node1Info); err != nil {
		log.Fatal(err)
	}

	// Register the same service on client node to enable discovery
	if err := node2.RegisterServiceHandler(calcHandler); err != nil {
		log.Fatal(err)
	}
	node2.Registry().RegisterClientConstructor(
		service.CalculatorProtocolID,
		func(peer *rpc.RpcPeer) interface{} {
			return proto.NewCalculatorClient(peer)
		},
	)

	// Wait for discovery
	time.Sleep(2 * time.Second)

	// Create calculator client
	client, err := node2.NewServiceClient(ctx, service.CalculatorProtocolID, node1.Host().ID())
	if err != nil {
		log.Fatal(err)
	}
	calcClient := client.(*proto.CalculatorClient)

	// Use the calculator service
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				// Test Add
				addResp := calcClient.Add(&proto.AddRequest{A: 5, B: 3})

				fmt.Printf("5 + 3 = %d\n", addResp.Result)

				// Test Multiply
				mulResp := calcClient.Multiply(&proto.MultiplyRequest{A: 4, B: 6})
				fmt.Printf("4 * 6 = %d\n", mulResp.Result)
			}
		}
	}()

	// Wait for interrupt signal
	fmt.Println("\nRunning... Press Ctrl+C to exit")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
