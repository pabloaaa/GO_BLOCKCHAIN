package mocks

import (
	proto "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"github.com/stretchr/testify/mock"
)

type MockNodeMessageHandlerImpl struct {
	mock.Mock
}

func NewMockNodeMessageHandler() *MockNodeMessageHandlerImpl {
	return &MockNodeMessageHandlerImpl{}
}

func (m *MockNodeMessageHandlerImpl) HandleNodeMessage(msg *proto.NodeMessage) {
	m.Called(msg)
}

func (m *MockNodeMessageHandlerImpl) handleWelcomeRequest(senderAddress int) {
	m.Called(senderAddress)
}

func (m *MockNodeMessageHandlerImpl) handleWelcomeResponse(nodesAddresses [][]byte) {
	m.Called(nodesAddresses)
}

func (m *MockNodeMessageHandlerImpl) BroadcastAddress(nodes []int, senderAddress int) {
	m.Called(nodes, senderAddress)
}

func (m *MockNodeMessageHandlerImpl) SyncNodes(address int, senderAddress int, latestBlockHash []byte) error {
	args := m.Called(address, senderAddress, latestBlockHash)
	return args.Error(0)
}
