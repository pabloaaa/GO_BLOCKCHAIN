package main

import (
	"context"
	"net"

	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"

	"google.golang.org/grpc"
)

type Node struct {
	pb.UnimplementedBlockchainServiceServer
	blockchain *Blockchain
	validator  *BlockValidator
	creator    *BlockCreator
}

func NewNode(blockchain *Blockchain, validator *BlockValidator, creator *BlockCreator) *Node {
	return &Node{
		blockchain: blockchain,
		validator:  validator,
		creator:    creator,
	}
}

func (n *Node) AddBlock(ctx context.Context, req *pb.BlockRequest) (*pb.BlockResponse, error) {
	block := BlockFromProto(req.Block)
	err := n.validator.ValidateAndAddBlock(block, n.blockchain)
	if err != nil {
		return &pb.BlockResponse{Success: false, Message: "Received block is invalid: " + err.Error()}, nil
	}

	return &pb.BlockResponse{Success: true, Message: "Block added successfully"}, nil
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
	go n.creator.Start(n.blockchain)

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
