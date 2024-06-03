package main

import (
	"context"
	"net"
	"testing"
	"time"

	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestNewNode(t *testing.T) {
	bc := NewBlockchain()
	validator := NewBlockValidator()
	creator := NewBlockCreator(validator)

	node := NewNode(bc, validator, creator)

	if node.blockchain != bc {
		t.Errorf("Expected node blockchain to be same as bc, but got different value")
	}

	if node.validator != validator {
		t.Errorf("Expected node validator to be same as validator, but got different value")
	}

	if node.creator != creator {
		t.Errorf("Expected node creator to be same as creator, but got different value")
	}
}

func TestGetBlockchain(t *testing.T) {
	bc := NewBlockchain()
	validator := NewBlockValidator()
	creator := NewBlockCreator(validator)

	node := NewNode(bc, validator, creator)

	resp, err := node.GetBlockchain(context.Background(), &pb.Empty{})
	if err != nil {
		t.Errorf("Expected GetBlockchain to not return error, but got %v", err)
	}

	if len(resp.Blocks) != len(bc.GetBlocks()) {
		t.Errorf("Expected GetBlockchain to return initial blockchain, but got different value")
	}
}

func TestStart(t *testing.T) {
	bc := NewBlockchain()
	validator := NewBlockValidator()
	creator := NewBlockCreator(validator)

	node := NewNode(bc, validator, creator)

	go func() {
		if err := node.Start("bufnet"); err != nil {
			t.Errorf("Expected Start to not return error, but got %v", err)
		}
	}()

	// Wait for the server to start
	time.Sleep(time.Second)

	// Connect to the server
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Errorf("Expected to connect to node, but got error: %v", err)
	}
	defer conn.Close()

	client := pb.NewBlockchainServiceClient(conn)

	// Add a block
	block := NewBlock(1, uint64(time.Now().Unix()), make([]Transaction, 0), "0", 0)
	_, err = client.AddBlock(context.Background(), &pb.BlockRequest{Block: block.ToProto()})
	if err != nil {
		t.Errorf("Expected to add block, but got error: %v", err)
	}

	// Get the blockchain
	resp, err := client.GetBlockchain(context.Background(), &pb.Empty{})
	if err != nil {
		t.Errorf("Expected to get blockchain, but got error: %v", err)
	}

	if len(resp.Blocks) != 2 {
		t.Errorf("Expected blockchain length to be 2, but got %d", len(resp.Blocks))
	}
}
