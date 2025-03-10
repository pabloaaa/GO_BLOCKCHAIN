package mocks

import (
	"github.com/pabloaaa/GO_BLOCKCHAIN/interfaces"
	"github.com/stretchr/testify/mock"
)

// MockTcpMessageSender is a mock implementation of the TcpMessageSender.
type MockTcpMessageSender struct {
	mock.Mock
}

// SendMsgToAddress mocks the SendMsgToAddress method.
func (m *MockTcpMessageSender) SendMsgToAddress(address int, data []byte) error {
	args := m.Called(address, data)
	return args.Error(0)
}

// Ensure MockTcpMessageSender implements MessageSender interface
var _ interfaces.MessageSender = (*MockTcpMessageSender)(nil)
