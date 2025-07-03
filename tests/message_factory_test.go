package tests

import (
	"testing"

	"github.com/pabloaaa/GO_BLOCKCHAIN/src"
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

func TestNewMessageFactory(t *testing.T) {
	factory := src.NewMessageFactory()
	if factory == nil {
		t.Error("NewMessageFactory returned nil")
	}
}

func TestCreateNodeMessage_WelcomeRequest(t *testing.T) {
	factory := src.NewMessageFactory()
	welcomeRequest := &block_chain.WelcomeRequest{
		SenderAddress: []byte("50001"),
	}

	mainMessage, err := factory.CreateNodeMessage(welcomeRequest)
	if err != nil {
		t.Errorf("CreateNodeMessage failed: %v", err)
	}

	if mainMessage == nil {
		t.Error("CreateNodeMessage returned nil")
	}

	nodeMessage := mainMessage.GetNodeMessage()
	if nodeMessage == nil {
		t.Error("NodeMessage is nil")
	}

	if nodeMessage.GetWelcomeRequest() == nil {
		t.Error("WelcomeRequest is nil")
	}

	if string(nodeMessage.GetWelcomeRequest().SenderAddress) != string(welcomeRequest.SenderAddress) {
		t.Errorf("Expected SenderAddress %s, got %s", string(welcomeRequest.SenderAddress), string(nodeMessage.GetWelcomeRequest().SenderAddress))
	}
}

func TestCreateNodeMessage_WelcomeResponse(t *testing.T) {
	factory := src.NewMessageFactory()
	welcomeResponse := &block_chain.WelcomeResponse{
		NodeAdresses: [][]byte{[]byte("50002")},
	}

	mainMessage, err := factory.CreateNodeMessage(welcomeResponse)
	if err != nil {
		t.Errorf("CreateNodeMessage failed: %v", err)
	}

	if mainMessage == nil {
		t.Error("CreateNodeMessage returned nil")
	}

	nodeMessage := mainMessage.GetNodeMessage()
	if nodeMessage == nil {
		t.Error("NodeMessage is nil")
	}

	if nodeMessage.GetWelcomeResponse() == nil {
		t.Error("WelcomeResponse is nil")
	}

	if len(nodeMessage.GetWelcomeResponse().NodeAdresses) != len(welcomeResponse.NodeAdresses) {
		t.Errorf("Expected %d addresses, got %d", len(welcomeResponse.NodeAdresses), len(nodeMessage.GetWelcomeResponse().NodeAdresses))
	}
}

func TestCreateBlockMessage_BlockchainSyncRequest(t *testing.T) {
	factory := src.NewMessageFactory()
	syncRequest := &block_chain.BlockchainSyncRequest{
		SenderAddress: []byte("50001"),
	}

	mainMessage, err := factory.CreateBlockMessage(syncRequest)
	if err != nil {
		t.Errorf("CreateBlockMessage failed: %v", err)
	}

	if mainMessage == nil {
		t.Error("CreateBlockMessage returned nil")
	}

	blockMessage := mainMessage.GetBlockMessage()
	if blockMessage == nil {
		t.Error("BlockMessage is nil")
	}

	if blockMessage.GetBlockchainSyncRequest() == nil {
		t.Error("BlockchainSyncRequest is nil")
	}

	if string(blockMessage.GetBlockchainSyncRequest().SenderAddress) != string(syncRequest.SenderAddress) {
		t.Errorf("Expected SenderAddress %s, got %s", string(syncRequest.SenderAddress), string(blockMessage.GetBlockchainSyncRequest().SenderAddress))
	}
}

func TestCreateBlockMessage_ApprovedBlock(t *testing.T) {
	factory := src.NewMessageFactory()
	approvedBlock := &block_chain.ApprovedBlock{
		Block: &block_chain.Block{
			Index: 1,
			Data:  100,
		},
	}

	mainMessage, err := factory.CreateBlockMessage(approvedBlock)
	if err != nil {
		t.Errorf("CreateBlockMessage failed: %v", err)
	}

	if mainMessage == nil {
		t.Error("CreateBlockMessage returned nil")
	}

	blockMessage := mainMessage.GetBlockMessage()
	if blockMessage == nil {
		t.Error("BlockMessage is nil")
	}

	if blockMessage.GetApprovedBlock() == nil {
		t.Error("ApprovedBlock is nil")
	}

	if blockMessage.GetApprovedBlock().Block.Index != approvedBlock.Block.Index {
		t.Errorf("Expected Block Index %d, got %d", approvedBlock.Block.Index, blockMessage.GetApprovedBlock().Block.Index)
	}
}

func TestCreateBlockMessage_BlockResponse(t *testing.T) {
	factory := src.NewMessageFactory()
	blockResponse := &block_chain.BlockResponse{
		Block: &block_chain.Block{Index: 1, Data: 100},
		SenderAddress: []byte("50001"),
	}

	mainMessage, err := factory.CreateBlockMessage(blockResponse)
	if err != nil {
		t.Errorf("CreateBlockMessage failed: %v", err)
	}

	if mainMessage == nil {
		t.Error("CreateBlockMessage returned nil")
	}

	blockMessage := mainMessage.GetBlockMessage()
	if blockMessage == nil {
		t.Error("BlockMessage is nil")
	}

	if blockMessage.GetBlockResponse() == nil {
		t.Error("BlockResponse is nil")
	}

	if blockMessage.GetBlockResponse().Block.Index != blockResponse.Block.Index {
		t.Errorf("Expected Block Index %d, got %d", blockResponse.Block.Index, blockMessage.GetBlockResponse().Block.Index)
	}
}

func TestCreateBlockMessage_BlockRequest(t *testing.T) {
	factory := src.NewMessageFactory()
	blockRequest := &block_chain.BlockRequest{
		SenderAddress: []byte("50001"),
	}

	mainMessage, err := factory.CreateBlockMessage(blockRequest)
	if err != nil {
		t.Errorf("CreateBlockMessage failed: %v", err)
	}

	if mainMessage == nil {
		t.Error("CreateBlockMessage returned nil")
	}

	blockMessage := mainMessage.GetBlockMessage()
	if blockMessage == nil {
		t.Error("BlockMessage is nil")
	}

	if blockMessage.GetBlockRequest() == nil {
		t.Error("BlockRequest is nil")
	}

	if string(blockMessage.GetBlockRequest().SenderAddress) != string(blockRequest.SenderAddress) {
		t.Errorf("Expected SenderAddress %s, got %s", string(blockRequest.SenderAddress), string(blockMessage.GetBlockRequest().SenderAddress))
	}
}

