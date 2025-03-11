package tests

import (
	"testing"

	"github.com/pabloaaa/GO_BLOCKCHAIN/mocks"
	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	. "github.com/pabloaaa/GO_BLOCKCHAIN/src"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
	"github.com/stretchr/testify/mock"
)

func TestHandleBlockMessage_BlockchainSyncRequest(t *testing.T) {
	mockBlockchain := new(mocks.MockBlockchain)
	mockSender := new(mocks.MockTcpMessageSender)
	handler := NewBlockMessageHandler(mockBlockchain, mockSender, 8080)

	mockBlockchain.On("GetLatestBlock").Return(&types.Block{})

	msg := &pb.BlockMessage{
		BlockMessageType: &pb.BlockMessage_BlockchainSyncRequest{
			BlockchainSyncRequest: &pb.BlockchainSyncRequest{SenderAddress: []byte("8081")},
		},
	}

	mockSender.On("SendMsgToAddress", 8081, mock.Anything).Return(nil)

	handler.HandleBlockMessage(msg)

	mockBlockchain.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestHandleBlockMessage_BlockResponse(t *testing.T) {
	mockBlockchain := new(mocks.MockBlockchain)
	mockSender := new(mocks.MockTcpMessageSender)
	handler := NewBlockMessageHandler(mockBlockchain, mockSender, 8080)

	block := &types.Block{}
	blockHash := block.CalculateHash()
	parentBlock := &types.BlockNode{Block: &types.Block{}}

	mockBlockchain.On("GetBlock", blockHash).Return((*types.BlockNode)(nil)).Once()
	mockBlockchain.On("GetBlock", block.PreviousHash).Return(parentBlock).Once()
	mockBlockchain.On("AddBlock", parentBlock, block).Return(nil).Once()

	msg := &pb.BlockMessage{
		BlockMessageType: &pb.BlockMessage_BlockResponse{
			BlockResponse: &pb.BlockResponse{Block: block.ToProto(), SenderAddress: []byte("8081")},
		},
	}

	mockSender.On("SendMsgToAddress", 8081, mock.Anything).Return(nil)

	handler.HandleBlockMessage(msg)

	mockBlockchain.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestHandleBlockMessage_ApprovedBlock(t *testing.T) {
	mockBlockchain := new(mocks.MockBlockchain)
	mockSender := new(mocks.MockTcpMessageSender)
	handler := NewBlockMessageHandler(mockBlockchain, mockSender, 8080)

	block := &types.Block{}
	msg := &pb.BlockMessage{
		BlockMessageType: &pb.BlockMessage_ApprovedBlock{
			ApprovedBlock: &pb.ApprovedBlock{Block: block.ToProto(), SenderAddress: []byte("8081")},
		},
	}

	mockSender.On("SendMsgToAddress", 8081, mock.Anything).Return(nil)

	handler.HandleBlockMessage(msg)

	mockSender.AssertExpectations(t)
}

func TestHandleBlockMessage_BlockRequest(t *testing.T) {
	mockBlockchain := new(mocks.MockBlockchain)
	mockSender := new(mocks.MockTcpMessageSender)
	handler := NewBlockMessageHandler(mockBlockchain, mockSender, 8080)

	block := &types.Block{}
	blockNode := &types.BlockNode{Block: block}
	mockBlockchain.On("GetBlockByIndex", uint64(1)).Return(blockNode)

	msg := &pb.BlockMessage{
		BlockMessageType: &pb.BlockMessage_BlockRequest{
			BlockRequest: &pb.BlockRequest{Index: 1, SenderAddress: []byte("8081")},
		},
	}

	mockSender.On("SendMsgToAddress", 8081, mock.Anything).Return(nil)

	handler.HandleBlockMessage(msg)

	mockBlockchain.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}
