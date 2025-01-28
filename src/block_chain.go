package src

import (
	"bytes"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/pabloaaa/GO_BLOCKCHAIN/interfaces"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

// Blockchain represents the blockchain.
type Blockchain struct {
	root *types.BlockNode
	mux  sync.Mutex
}

// NewBlockchain creates a new Blockchain.
func NewBlockchain() *Blockchain {
	blockchain := &Blockchain{}
	blockchain.createGenesisBlock()
	return blockchain
}

// createGenesisBlock creates the genesis block.
func (bc *Blockchain) createGenesisBlock() {
	genesisBlock := &types.Block{
		Index:        0,
		Timestamp:    uint64(time.Now().Unix()),
		Transactions: make([]types.Transaction, 0),
		PreviousHash: []byte("0"),
		Data:         0,
	}
	bc.root = &types.BlockNode{
		Block:  genesisBlock,
		Parent: nil,
		Childs: make([]*types.BlockNode, 0),
	}
}

// GetRoot returns the root block node.
func (bc *Blockchain) GetRoot() *types.BlockNode {
	return bc.root
}

// AddBlock adds a new block to the blockchain.
func (bc *Blockchain) AddBlock(parent *types.BlockNode, block *types.Block) error {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	log.Printf("Attempting to add block with index %d", block.Index)

	// Check if a block with the same index already exists
	existingBlockNode := bc.GetBlockByIndex(block.Index)
	if existingBlockNode != nil {
		log.Printf("Block with index %d already exists", block.Index)
		return errors.New("Block with the same index already exists")
	}

	if err := bc.ValidateBlock(block, parent.Block); err != nil {
		log.Printf("Block validation failed: %v", err)
		return err
	}

	blockNode := &types.BlockNode{
		Block:  block,
		Parent: parent,
		Childs: make([]*types.BlockNode, 0),
	}

	parent.Childs = append(parent.Childs, blockNode)

	// Call ApproveBlock to check and set checkpoint
	bc.ApproveBlock(blockNode)

	log.Printf("Block with index %d added successfully", block.Index)
	return nil
}

// ApproveBlock sets the checkpoint flag for the block if it meets the criteria.
func (bc *Blockchain) ApproveBlock(blockNode *types.BlockNode) {
	if blockNode.Block.Index%10 == 0 {
		blockNode.Block.Checkpoint = true
	}
	blockNode.Block.Checkpoint = false
}

// ValidateBlock validates a block against its parent block.
func (bc *Blockchain) ValidateBlock(block *types.Block, parentBlock *types.Block) error {
	if block.Index != parentBlock.Index+1 {
		return errors.New("Block index is not valid")
	}

	if !bytes.Equal(block.PreviousHash, parentBlock.CalculateHash()) {
		return errors.New("Previous hash is not valid")
	}

	hashPrefix := block.CalculateHash()[:2]
	if !bytes.Equal(hashPrefix, []byte("00")) {
		return errors.New("Block hash is not valid")
	}

	return nil
}

// convertToBlockNodes converts a slice of blocks to a slice of block nodes.
func (bc *Blockchain) convertToBlockNodes(blocks []*types.Block) []*types.BlockNode {
	blockNodes := make([]*types.BlockNode, len(blocks))
	for i, block := range blocks {
		blockNodes[i] = &types.BlockNode{
			Block:  block,
			Parent: nil, // You need to set the correct parent here
			Childs: make([]*types.BlockNode, 0),
		}
	}
	return blockNodes
}

// ReplaceBlocks replaces the current blocks with new blocks.
func (bc *Blockchain) ReplaceBlocks(blocks []*types.Block) {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	blockNodes := bc.convertToBlockNodes(blocks)
	bc.root = blockNodes[0] // Assuming the first block is the root
	log.Println("Blockchain replaced with new blocks")
}

// BlockExists checks if a block exists in the blockchain.
func (bc *Blockchain) BlockExists(hash []byte) bool {
	return bc.GetBlock(hash) != nil
}

// TraverseTree traverses the blockchain tree and applies a callback function to each node.
func (bc *Blockchain) TraverseTree(callback func(node *types.BlockNode) bool) {
	var queue []*types.BlockNode

	queue = append(queue, bc.root)

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if callback(node) {
			return
		}

		for _, child := range node.Childs {
			queue = append(queue, child)
		}
	}
}

// GetBlock returns a block node by its hash.
func (bc *Blockchain) GetBlock(hash []byte) *types.BlockNode {
	var foundNode *types.BlockNode
	bc.TraverseTree(func(node *types.BlockNode) bool {
		calculatedHash := node.Block.CalculateHash()
		if bytes.Equal(calculatedHash, hash) {
			foundNode = node
			return true
		}
		return false
	})
	return foundNode
}

// GetBlockByIndex returns a block node by its index.
func (bc *Blockchain) GetBlockByIndex(index uint64) *types.BlockNode {
	var foundNode *types.BlockNode
	bc.TraverseTree(func(node *types.BlockNode) bool {
		if node.Block.Index == index {
			foundNode = node
			return true
		}
		return false
	})
	return foundNode
}

// GetLatestBlock returns the latest approved block in the blockchain, or the latest block if no approved block exists.
func (bc *Blockchain) GetLatestBlock() *types.Block {
	var latestApprovedBlock *types.Block
	var latestBlock *types.Block
	bc.TraverseTree(func(node *types.BlockNode) bool {
		if node.Block.Checkpoint {
			latestApprovedBlock = node.Block
		}
		latestBlock = node.Block
		return false
	})
	if latestApprovedBlock == nil {
		return latestBlock
	}
	return latestApprovedBlock
}

// GetLatestBlockNode returns the latest block node in the blockchain.
func (bc *Blockchain) GetLatestBlockNode() *types.BlockNode {
	var longestPath []*types.BlockNode
	bc.TraverseTree(func(node *types.BlockNode) bool {
		if len(node.Childs) > len(longestPath) {
			longestPath = node.Childs
		}
		return false
	})
	if len(longestPath) == 0 {
		return bc.root
	}
	return longestPath[len(longestPath)-1]
}

// GenerateNewBlock generates a new block with the given transactions.
func (bc *Blockchain) GenerateNewBlock(transaction []types.Transaction) *types.Block {
	latestBlock := bc.GetLatestBlock()
	newBlock := &types.Block{
		Index:        latestBlock.Index + 1,
		Timestamp:    uint64(time.Now().Unix()),
		Transactions: transaction,
		PreviousHash: latestBlock.CalculateHash(),
		Data:         0,
	}
	return newBlock
}

// Ensure Blockchain implements BlockchainInterface
var _ interfaces.BlockchainInterface = (*Blockchain)(nil)
