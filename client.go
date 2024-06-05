package main

import (
	"context"
	"time"

	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"

	"google.golang.org/grpc"
)

type Client struct {
	nodeClient pb.BlockchainServiceClient
	blockchain *Blockchain
	creator    *BlockCreator
}

func NewClient(address string) (*Client, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	blockchain := NewBlockchain()
	validator := NewBlockValidator()
	creator := NewBlockCreator(validator)

	return &Client{
		nodeClient: pb.NewBlockchainServiceClient(conn),
		blockchain: blockchain,
		creator:    creator,
	}, nil
}

func (c *Client) Start() {
	c.creator.Start(c.blockchain)
}

func (c *Client) AddBlock(block *Block) error {
	// Notify the node about the new block
	blockRequest := &pb.BlockRequest{
		Block: block.ToProto(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.nodeClient.AddBlock(ctx, blockRequest)
	return err
}

func (c *Client) GetBlockchain(ctx context.Context, empty *pb.Empty, opts ...grpc.CallOption) (*pb.BlockchainResponse, error) {
	return c.nodeClient.GetBlockchain(ctx, empty, opts...)
}
