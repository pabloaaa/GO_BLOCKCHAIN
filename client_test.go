package main

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"google.golang.org/grpc"
)

func setupClientTest(t *testing.T) (pb.BlockchainServiceClient, *grpc.ClientConn) {
	// Create a new blockchain and start the server
	blockchain := NewBlockchain()
	node := NewNode(blockchain)
	go func() {
		if err := node.Start("localhost:50051"); err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	// Create a client
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	client := pb.NewBlockchainServiceClient(conn)

	return client, conn
}

func TestClientsCompeteToAddBlocks(t *testing.T) {
	client, conn := setupClientTest(t)
	defer conn.Close()

	// Create 3 clients
	clients := make([]*Client, 3)
	for i := 0; i < 3; i++ {
		client, err := NewClient("localhost:50051")
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		clients[i] = client
	}

	// Each client tries to add blocks to the blockchain
	var wg sync.WaitGroup
	for i, client := range clients {
		wg.Add(1)
		go func(i int, client *Client) {
			defer wg.Done()
			for j := 0; j < 3; j++ {
				// Get the last block
				resp, err := client.GetBlockchain(context.Background(), &pb.Empty{})
				if err != nil {
					t.Errorf("Client %d failed to get blockchain: %v", i, err)
					return
				}
				lastBlock := resp.Blocks[len(resp.Blocks)-1]

				// Create a new block
				block := &Block{
					Index:        lastBlock.Index + 1,
					Timestamp:    uint64(time.Now().Unix()),
					PreviousHash: lastBlock.Hash,
				}

				// Find a nonce that makes the block's hash valid
				var hash string
				nonce := 0
				for {
					block.Data = uint64(nonce)
					block.calculateHash()
					hash = block.Hash
					if strings.HasPrefix(hash, "00") {
						break
					}
					nonce++
				}

				err = client.AddBlock(block)
				if err != nil {
					t.Errorf("Client %d failed to add block: %v", i, err)
				}

				// Check if this client has added 3 blocks
				resp, err = client.GetBlockchain(context.Background(), &pb.Empty{})
				if err != nil {
					t.Errorf("Client %d failed to get blockchain: %v", i, err)
					return
				}
				if len(resp.Blocks) == 4 { // 3 new blocks + genesis block
					t.Logf("Client %d won the race", i)
					return
				}
			}
		}(i, client)
	}

	// Wait for all clients to finish
	wg.Wait()

	// Check the length of the blockchain
	resp, err := client.GetBlockchain(context.Background(), &pb.Empty{})
	if err != nil {
		t.Fatalf("Failed to get blockchain: %v", err)
	}
	if len(resp.Blocks) != 4 { // 3 blocks from the winning client + genesis block
		t.Errorf("Expected blockchain length to be 4, got %d", len(resp.Blocks))
	}
}
