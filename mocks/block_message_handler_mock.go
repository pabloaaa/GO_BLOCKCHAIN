package mocks

import (
	"net"

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

func (m *MockBlockMessageHandlerImpl) HandleBlockMessage(msg *proto.BlockMessage, conn net.Conn) {
	m.Called(msg, conn)
}

func (m *MockBlockMessageHandlerImpl) HandleGetLatestBlock(data []byte, address string) {
	m.Called(data, address)
}

func (m *MockBlockMessageHandlerImpl) HandleGetBlockRequest(hash []byte, address string) {
	m.Called(hash, address)
}

func (m *MockBlockMessageHandlerImpl) HandleBlockResponse(data []byte, address string) {
	m.Called(data, address)
}

func (m *MockBlockMessageHandlerImpl) SendBlock(address string, blockNode *types.BlockNode) {
	m.Called(address, blockNode)
}

func (m *MockBlockMessageHandlerImpl) SendLatestBlock(address string) {
	m.Called(address)
}

func (m *MockBlockMessageHandlerImpl) GetBlock(address string, blockHash []byte) {
	m.Called(address, blockHash)
}

func (m *MockBlockMessageHandlerImpl) GetLatestBlock(address string) {
	m.Called(address)
}

func (m *MockBlockMessageHandlerImpl) BroadcastLatestBlock(nodes [][]byte) {
	m.Called(nodes)
}
