package tests

import (
	"net"
	"testing"

	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"github.com/stretchr/testify/mock"
)

type MockBlockchain struct {
	mock.Mock
}

func (m *MockBlockchain) GetBlock(hash []byte) *BlockNode {
	args := m.Called(hash)
	return args.Get(0).(*BlockNode)
}

func (m *MockBlockchain) GetLatestBlock() *BlockNode {
	args := m.Called()
	return args.Get(0).(*BlockNode)
}

func (m *MockBlockchain) BlockExists(hash []byte) bool {
	args := m.Called(hash)
	return args.Bool(0)
}

func (m *MockBlockchain) ValidateBlock(block *BlockNode, parent *Block) error {
	args := m.Called(block, parent)
	return args.Error(0)
}

func (m *MockBlockchain) AddBlock(parent *BlockNode, block *BlockNode) error {
	args := m.Called(parent, block)
	return args.Error(0)
}

// MockConn to mock dla net.Conn
type MockConn struct {
	mock.Mock
	net.Conn
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *MockConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestHandleBlockMessage_GetLatestBlockRequest(t *testing.T) {
	blockchain := new(MockBlockchain)
	handler := NewBlockMessageHandler(blockchain)

	conn := new(MockConn)
	conn.On("LocalAddr").Return(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080})

	msg := &block_chain.BlockMessage{
		BlockMessageType: &block_chain.BlockMessage_GetLatestBlockRequest{},
	}

	handler.HandleBlockMessage(msg, conn)

	// Sprawdź, czy metoda handleGetLatestBlock została wywołana
	blockchain.AssertExpectations(t)
}

func TestHandleBlockMessage_GetBlockRequest(t *testing.T) {
	blockchain := new(MockBlockchain)
	handler := NewBlockMessageHandler(blockchain)

	conn := new(MockConn)
	conn.On("LocalAddr").Return(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080})

	hash := []byte("somehash")
	msg := &block_chain.BlockMessage{
		BlockMessageType: &block_chain.BlockMessage_GetBlockRequest_{GetBlockRequest_: &block_chain.GetBlockRequest{Hash: hash}},
	}

	blockchain.On("GetBlock", hash).Return(&BlockNode{})

	handler.HandleBlockMessage(msg, conn)

	// Sprawdź, czy metoda handleGetBlockRequest została wywołana
	blockchain.AssertExpectations(t)
}

func TestHandleBlockMessage_BlockResponse(t *testing.T) {
	blockchain := new(MockBlockchain)
	handler := NewBlockMessageHandler(blockchain)

	conn := new(MockConn)
	conn.On("LocalAddr").Return(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080})

	data := []byte("blockdata")
	msg := &block_chain.BlockMessage{
		BlockMessageType: &block_chain.BlockMessage_BlockResponse{BlockResponse: &block_chain.BlockResponse{Message: data}},
	}

	blockchain.On("BlockExists", mock.Anything).Return(false)
	blockchain.On("GetBlock", mock.Anything).Return(&BlockNode{})
	blockchain.On("ValidateBlock", mock.Anything, mock.Anything).Return(nil)
	blockchain.On("AddBlock", mock.Anything, mock.Anything).Return(nil)

	handler.HandleBlockMessage(msg, conn)

	// Sprawdź, czy metoda handleBlockResponse została wywołana
	blockchain.AssertExpectations(t)
}
