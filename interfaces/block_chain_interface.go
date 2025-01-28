package interfaces

import (
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

type BlockchainInterface interface {
	GetLatestBlock() *types.Block
	GetBlock(hash []byte) *types.BlockNode
	AddBlock(parent *types.BlockNode, block *types.Block) error
	ValidateBlock(block *types.Block, parent *types.Block) error
	BlockExists(hash []byte) bool
	GenerateNewBlock(transaction []types.Transaction) *types.Block
	GetRoot() *types.BlockNode
	TraverseTree(callback func(node *types.BlockNode) bool)
	GetBlockByIndex(index uint64) *types.BlockNode
	ReplaceBlocks(blocks []*types.Block)
}
