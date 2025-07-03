package interfaces

import (
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

type BlockchainInterface interface {
	GetLatestBlock() *types.Block
	GetLatestApprovedBlock() *types.Block
	GetBlockWithHighestIndex() *types.Block
	GetBlock(hash []byte) *types.BlockNode
	AddBlock(parent *types.BlockNode, block *types.Block) error
	ValidateBlock(block *types.Block, parent *types.Block) error
	BlockExists(hash []byte) bool
	GenerateNewBlock() *types.Block
	GetRoot() *types.BlockNode
	TraverseTree(callback func(node *types.BlockNode) bool)
	GetBlockByIndex(index uint64) *types.BlockNode
	ReplaceBlocks(blocks []*types.Block)
	GetReward() uint64
	RewardNode(address int, amount int64) error
}
