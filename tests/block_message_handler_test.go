package tests

import (
	"testing"

	"github.com/pabloaaa/GO_BLOCKCHAIN/mocks"
	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	. "github.com/pabloaaa/GO_BLOCKCHAIN/src"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
	"google.golang.org/protobuf/proto"
)

func TestHandleBlockMessage_GetLatestBlockRequest(t *testing.T) {
	mockBlockchain := new(mocks.MockBlockchain)
	testSender := NewTestSender()
	handler := NewBlockMessageHandler(mockBlockchain, testSender)

	mockBlockchain.On("GetLatestBlock").Return(&types.Block{})

	msg := &pb.BlockMessage{
		BlockMessageType: &pb.BlockMessage_GetLatestBlockRequest{
			GetLatestBlockRequest: &pb.GetLatestBlockRequest{},
		},
	}

	handler.HandleBlockMessage(msg)

	mockBlockchain.AssertExpectations(t)
	if len(testSender.GetQueue()) == 0 {
		t.Errorf("Expected data to be sent, but queue is empty")
	}
}

func TestHandleBlockMessage_GetBlockRequest(t *testing.T) {
	mockBlockchain := new(mocks.MockBlockchain)
	testSender := NewTestSender()
	handler := NewBlockMessageHandler(mockBlockchain, testSender)

	hash := []byte("somehash")
	mockBlockchain.On("GetBlock", hash).Return(&types.BlockNode{Block: &types.Block{}})

	msg := &pb.BlockMessage{
		BlockMessageType: &pb.BlockMessage_GetBlockRequest_{
			GetBlockRequest_: &pb.GetBlockRequest{Hash: hash},
		},
	}

	handler.HandleBlockMessage(msg)

	mockBlockchain.AssertExpectations(t)
	if len(testSender.GetQueue()) == 0 {
		t.Errorf("Expected data to be sent, but queue is empty")
	}
}

func TestHandleBlockMessage_BlockResponse(t *testing.T) {
	mockBlockchain := new(mocks.MockBlockchain)
	testSender := NewTestSender()
	handler := NewBlockMessageHandler(mockBlockchain, testSender)

	block := &types.Block{
		Transactions: []types.Transaction{},
	}
	data, _ := proto.Marshal(block.ToProto())
	blockHash := block.CalculateHash()
	parentBlock := &types.BlockNode{Block: &types.Block{
		Transactions: []types.Transaction{},
	}}

	mockBlockchain.On("BlockExists", blockHash).Return(true).Once()
	mockBlockchain.On("GetBlock", block.PreviousHash).Return(parentBlock).Once()
	mockBlockchain.On("ValidateBlock", block, parentBlock.Block).Return(nil).Once()
	mockBlockchain.On("AddBlock", parentBlock, block).Return(nil).Once()

	msg := &pb.BlockMessage{
		BlockMessageType: &pb.BlockMessage_BlockResponse{
			BlockResponse: &pb.BlockResponse{Message: data},
		},
	}

	handler.HandleBlockMessage(msg)

	mockBlockchain.AssertExpectations(t)
	if len(testSender.GetQueue()) == 0 {
		t.Errorf("Expected data to be sent, but queue is empty")
	}
}
