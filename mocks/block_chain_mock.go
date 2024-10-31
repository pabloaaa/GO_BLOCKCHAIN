package mocks

import (
	"github.com/pabloaaa/GO_BLOCKCHAIN/interfaces"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
	"github.com/stretchr/testify/mock"
)

type MockBlockchain struct {
	mock.Mock
}

func (m *MockBlockchain) GetLatestBlock() *types.Block {
	args := m.Called()
	return args.Get(0).(*types.Block)
}

func (m *MockBlockchain) GetBlock(hash []byte) *types.BlockNode {
	args := m.Called(hash)
	return args.Get(0).(*types.BlockNode)
}

func (m *MockBlockchain) AddBlock(parent *types.BlockNode, block *types.Block) error {
	args := m.Called(parent, block)
	return args.Error(0)
}

func (m *MockBlockchain) ValidateBlock(block *types.Block, parent *types.Block) error {
	args := m.Called(block, parent)
	return args.Error(0)
}

func (m *MockBlockchain) BlockExists(hash []byte) bool {
	args := m.Called(hash)
	return args.Bool(0)
}

func (m *MockBlockchain) GenerateNewBlock(transaction []types.Transaction) *types.Block {
	args := m.Called(transaction)
	return args.Get(0).(*types.Block)
}

func (m *MockBlockchain) GetRoot() *types.BlockNode {
	args := m.Called()
	return args.Get(0).(*types.BlockNode)
}

// Ensure MockBlockchain implements BlockchainInterface
var _ interfaces.BlockchainInterface = (*MockBlockchain)(nil)
