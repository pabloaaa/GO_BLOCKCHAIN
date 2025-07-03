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
	root         *types.BlockNode
	reward       uint64
	mux          sync.Mutex
	blockHashMap map[string]*types.BlockNode // Map of block hashes (converted to string) to block nodes
}

// NewBlockchain creates a new Blockchain.
func NewBlockchain() *Blockchain {
	blockchain := &Blockchain{
		blockHashMap: make(map[string]*types.BlockNode),
	}
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

	// Add genesis block to the hash map
	genesisHash := genesisBlock.CalculateHash()
	bc.blockHashMap[string(genesisHash)] = bc.root
}

// GetRoot returns the root block node.
func (bc *Blockchain) GetRoot() *types.BlockNode {
	return bc.root
}

// AddBlock adds a new block to the blockchain.
func (bc *Blockchain) AddBlock(parent *types.BlockNode, block *types.Block) error {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	// Check if a block with the same hash already exists
	blockHash := string(block.CalculateHash())
	if _, exists := bc.blockHashMap[blockHash]; exists {
		Error(fmt.Sprintf("block_chain: Block with hash %x already exists", block.CalculateHash()))
		return errors.New("Block with the same hash already exists")
	}

	// Check if a block with the same index already exists
	for _, node := range bc.blockHashMap {
		if node.Block.Index == block.Index {
			Error(fmt.Sprintf("block_chain: Block with index %d already exists", block.Index))
			return errors.New("Block with the same index already exists")
		}
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

	// Add the new block to the hash map
	bc.blockHashMap[blockHash] = blockNode

	// Call ApproveBlock to check and set checkpoint
	bc.ApproveBlock(blockNode)

	Debug(fmt.Sprintf("block_chain: Block with index %d added successfully", block.Index))
	return nil
}

// ApproveBlock sets the checkpoint flag for the block to true for every block.
func (bc *Blockchain) ApproveBlock(blockNode *types.BlockNode) {
	blockNode.Block.Checkpoint = true
	Debug(fmt.Sprintf("block_chain: Checkpoint set to true for block index %d", blockNode.Block.Index))
}

// ValidateBlock validates a block against its parent block.
func (bc *Blockchain) ValidateBlock(block *types.Block, parentBlock *types.Block) error {
	if block.Index != parentBlock.Index+1 {
		return errors.New("Block index is not valid")
	}

	if !bytes.Equal(block.PreviousHash, parentBlock.CalculateHash()) {
		return errors.New("Previous hash is not valid")
	}
	hashPrefix := block.CalculateHash()[:3]
	if !bytes.Equal(hashPrefix, []byte("000")) {
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

// BlockExists checks if a block exists in the blockchain using the hash map.
func (bc *Blockchain) BlockExists(hash []byte) bool {
	bc.mux.Lock()
	defer bc.mux.Unlock()
	_, exists := bc.blockHashMap[string(hash)]
	return exists
}

// TraverseTree traverses the blockchain using the hash map and applies a callback function to each node.
func (bc *Blockchain) TraverseTree(callback func(node *types.BlockNode) bool) {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	for _, node := range bc.blockHashMap {
		if callback(node) {
			return
		}
	}
}

// GetBlock returns a block node by its hash using the hash map.
func (bc *Blockchain) GetBlock(hash []byte) *types.BlockNode {
	bc.mux.Lock()
	defer bc.mux.Unlock()
	return bc.blockHashMap[string(hash)]
}

// GetBlockByIndex returns a block node by its index using the hash map.
func (bc *Blockchain) GetBlockByIndex(index uint64) *types.BlockNode {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	for _, node := range bc.blockHashMap {
		if node.Block.Index == index {
			return node
		}
	}
	return nil
}

// GetLatestBlock returns the latest approved block or the block with the highest index if no approved block exists.
func (bc *Blockchain) GetLatestBlock() *types.Block {
	latestBlock := bc.GetLatestApprovedBlock()
	if latestBlock == nil {
		latestBlock = bc.GetBlockWithHighestIndex()
	}
	return latestBlock
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

// GenerateNewBlock generates a new block.
func (bc *Blockchain) GenerateNewBlock() *types.Block {
	latestBlock := bc.GetLatestBlock()
	newBlock := &types.Block{
		Index:        latestBlock.Index + 1,
		Timestamp:    uint64(time.Now().Unix()),
		PreviousHash: latestBlock.CalculateHash(),
		Data:         0,
	}
	return newBlock
}

// GetReward returns the current reward.
func (bc *Blockchain) GetReward() uint64 {
	return bc.reward
}

// RewardNode rewards a node with the specified amount.
func (bc *Blockchain) RewardNode(address int, amount int64) error {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	if amount < 0 && bc.reward < uint64(-amount) {
		return fmt.Errorf("insufficient reward to deduct")
	}

	bc.reward += uint64(amount)
	return nil
}

// GetBlockWithHighestIndex returns the block with the highest index.
func (bc *Blockchain) GetBlockWithHighestIndex() *types.Block {
	var highestBlock *types.Block
	bc.TraverseTree(func(node *types.BlockNode) bool {
		if highestBlock == nil || node.Block.Index > highestBlock.Index {
			highestBlock = node.Block
		}
		return false
	})
	return highestBlock
}

// Ensure Blockchain implements BlockchainInterface
var _ interfaces.BlockchainInterface = (*Blockchain)(nil)
