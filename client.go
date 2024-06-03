package main

import (
	"context"

	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"

	"google.golang.org/grpc"
)

type Client struct {
	nodeClient pb.BlockchainServiceClient
}

func NewClient(address string) (*Client, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := pb.NewBlockchainServiceClient(conn)

	return &Client{nodeClient: client}, nil
}

func (c *Client) AddBlock(block *Block) error {
	_, err := c.nodeClient.AddBlock(context.Background(), &pb.BlockRequest{Block: block.ToProto()})
	return err
}

func (c *Client) SubscribeNewBlocks() error {
	stream, err := c.nodeClient.SubscribeNewBlocks(context.Background(), &pb.Empty{})
	if err != nil {
		return err
	}

	for {
		block, err := stream.Recv()
		if err != nil {
			return err
		}

		// Handle the new block
		_ = BlockFromProto(block)
	}

	return nil
}
