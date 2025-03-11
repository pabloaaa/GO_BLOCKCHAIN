package mocks

import (
	proto "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
	"github.com/stretchr/testify/mock"
)

type MockBlockMessageHandlerImpl struct {
	mock.Mock
	blockchain *MockBlockchain
}

func NewMockBlockMessageHandler(blockchain *MockBlockchain) *MockBlockMessageHandlerImpl {
	return &MockBlockMessageHandlerImpl{blockchain: blockchain}
}

func (m *MockBlockMessageHandlerImpl) HandleBlockMessage(msg *proto.BlockMessage) {
	m.Called(msg)
}

func (m *MockBlockMessageHandlerImpl) handleBlockchainSyncRequest(blockChainSyncRequest *proto.BlockchainSyncRequest) {
	m.Called(blockChainSyncRequest)
}

func (m *MockBlockMessageHandlerImpl) handleBlockResponse(blockResponse *proto.BlockResponse) {
	m.Called(blockResponse)
}

func (m *MockBlockMessageHandlerImpl) handleApprovedBlock(approvedBlock *proto.ApprovedBlock) {
	m.Called(approvedBlock)
}

func (m *MockBlockMessageHandlerImpl) handleBlockRequest(blockRequest *proto.BlockRequest) {
	m.Called(blockRequest)
}

func (m *MockBlockMessageHandlerImpl) BroadcastApprovedBlock(block *types.Block, nodes []int) {
	m.Called(block, nodes)
}

func (m *MockBlockMessageHandlerImpl) GetBlock(hash []byte) *types.BlockNode {
	args := m.Called(hash)
	if blockNode, ok := args.Get(0).(*types.BlockNode); ok {
		return blockNode
	}
	return nil
}

func (m *MockBlockMessageHandlerImpl) AddBlock(parent *types.BlockNode, block *types.Block) error {
	args := m.Called(parent, block)
	return args.Error(0)
}
