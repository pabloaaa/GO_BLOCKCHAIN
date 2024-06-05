package main

import (
	"context"
	"log"
	"net"
	"sync"

	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"

	"google.golang.org/grpc"
)

type Node struct {
	pb.UnimplementedBlockchainServiceServer
	blockchain  *Blockchain
	subscribers []pb.BlockchainService_SubscribeNewBlocksServer
	lock        sync.Mutex
}

func NewNode(blockchain *Blockchain) *Node {
	return &Node{
		blockchain: blockchain,
	}
}

func (n *Node) SubscribeNewBlocks(_ *pb.Empty, stream pb.BlockchainService_SubscribeNewBlocksServer) error {
	n.subscribers = append(n.subscribers, stream)
	return nil
}

func (n *Node) GetBlockchain(ctx context.Context, req *pb.Empty) (*pb.BlockchainResponse, error) {
	blocks := n.blockchain.GetBlocks() // You need to implement this method in Blockchain

	pbBlocks := make([]*pb.Block, len(blocks))
	for i, block := range blocks {
		pbBlocks[i] = block.ToProto()
	}

	return &pb.BlockchainResponse{Blocks: pbBlocks}, nil
}

func (n *Node) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	pb.RegisterBlockchainServiceServer(server, n) // You need to implement the required methods in Node

	if err := server.Serve(listener); err != nil {
		return err
	}

	return nil
}

func (n *Node) AddBlock(ctx context.Context, req *pb.BlockRequest) (*pb.BlockResponse, error) {
	n.lock.Lock()
	defer n.lock.Unlock()

	// Get the last block again after acquiring the lock
	lastBlock := n.blockchain.Last()

	// Convert the protobuf block to your Block type
	block := BlockFromProto(req.GetBlock())

	// Update the index and previous hash of the new block based on the last block
	block.Index = lastBlock.Index + 1
	block.PreviousHash = lastBlock.Hash

	// Validate the block before adding it to the blockchain
	validator := NewBlockValidator()
	if err := validator.ValidateAndAddBlock(block, n.blockchain); err != nil {
		return &pb.BlockResponse{Success: false, Message: "Failed to add block"}, err
	}

	// Notify the subscribers about the new block
	n.NotifySubscribers(block)

	return &pb.BlockResponse{Success: true, Message: "Block added successfully", Block: block.ToProto()}, nil
}

func (n *Node) NotifySubscribers(block *Block) {
	for _, subscriber := range n.subscribers {
		// Convert your block to protobuf format
		pbBlock := block.ToProto()

		// Create a new Block to send to the subscriber
		newBlock := &pb.Block{
			Index:        pbBlock.Index,
			Timestamp:    pbBlock.Timestamp,
			Transactions: pbBlock.Transactions,
			PreviousHash: pbBlock.PreviousHash,
			Hash:         pbBlock.Hash,
			Data:         pbBlock.Data,
		}

		// Send the updated block to the subscriber
		if err := subscriber.Send(newBlock); err != nil {
			// Handle error
			log.Printf("Failed to send updated block to subscriber: %v", err)
		}
	}
}
