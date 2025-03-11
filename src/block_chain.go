package src

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
	// Odczytaj plik konfiguracyjny
	configData, err := ioutil.ReadFile("/Users/pawelnowakowski/go_projects/src/GO_BLOCKCHAIN/config.json")
	if err != nil {
		Error(fmt.Sprintf("Failed to read config file: %v", err))
		return
	}

	// Zdekoduj dane genesis block
	var config struct {
		GenesisBlock types.Block `json:"genesis_block"`
	}
	err = json.Unmarshal(configData, &config)
	if err != nil {
		Error(fmt.Sprintf("Failed to unmarshal config data: %v", err))
		return
	}

	// Utwórz genesis block na podstawie danych z pliku konfiguracyjnego
	genesisBlock := &config.GenesisBlock
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

	// Check if a block with the same index already exists
	existingBlockNode := bc.GetBlockByIndex(block.Index)
	if existingBlockNode != nil {
		Error(fmt.Sprintf("block_chain: Block with index %d already exists", block.Index))
		return errors.New("Block with the same index already exists")
	}

	if err := bc.ValidateBlock(block, parent.Block); err != nil {
		Error(fmt.Sprintf("block_chain: Block validation failed: %v", err))
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

	Debug(fmt.Sprintf("block_chain: Block with index %d added successfully", block.Index))
	return nil
}

// ApproveBlock sets the checkpoint flag for the block if it meets the criteria.
func (bc *Blockchain) ApproveBlock(blockNode *types.BlockNode) {
	if blockNode.Block.Index%10 == 0 {
		blockNode.Block.Checkpoint = true
		Debug(fmt.Sprintf("block_chain: Checkpoint set to true for block index %d", blockNode.Block.Index))
	} else {
		blockNode.Block.Checkpoint = false
	}
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

// GetLatestBlock returns the latest approved block or the block with the highest index if no approved block exists.
func (bc *Blockchain) GetLatestBlock() *types.Block {
	latestBlock := bc.GetLatestApprovedBlock()
	if latestBlock == nil {
		latestBlock = bc.getBlockWithHighestIndex()
	}
	return latestBlock
}

// getBlockWithHighestIndex returns the block with the highest index.
func (bc *Blockchain) getBlockWithHighestIndex() *types.Block {
	var highestBlock *types.Block
	bc.TraverseTree(func(node *types.BlockNode) bool {
		if highestBlock == nil || node.Block.Index > highestBlock.Index {
			highestBlock = node.Block
		}
		return false
	})
	return highestBlock
}

// GetLatestApprovedBlock returns the latest approved block in the blockchain.
func (bc *Blockchain) GetLatestApprovedBlock() *types.Block {
	var latestApprovedBlock *types.Block
	bc.TraverseTree(func(node *types.BlockNode) bool {
		if node.Block.Checkpoint {
			latestApprovedBlock = node.Block
		}
		return false
	})
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
