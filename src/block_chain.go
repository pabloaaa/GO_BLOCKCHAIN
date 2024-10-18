package src

import (
	"bytes"
	"errors"
	"sync"
	"time"
)

type BlockNode struct {
	Block  *Block
	Parent *BlockNode
	Childs []*BlockNode
}
type Blockchain struct {
	root *BlockNode
	mux  sync.Mutex
}

func NewBlockchain() *Blockchain {
	blockchain := &Blockchain{}
	blockchain.createGenesisBlock()
	return blockchain
}
func (bc *Blockchain) createGenesisBlock() {
	genesisBlock := &Block{
		Index:        0,
		Timestamp:    uint64(time.Now().Unix()),
		Transactions: make([]Transaction, 0),
		PreviousHash: []byte("0"),
		Data:         0,
	}
	bc.root = &BlockNode{
		Block:  genesisBlock,
		Parent: nil,
		Childs: make([]*BlockNode, 0),
	}
}
func (bc *Blockchain) GetRoot() *BlockNode {
	return bc.root
}
func (bc *Blockchain) AddBlock(parent *BlockNode, block *Block) error {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	if err := bc.ValidateBlock(block, parent.Block); err != nil {
		return err
	}

	blockNode := &BlockNode{
		Block:  block,
		Parent: parent,
		Childs: make([]*BlockNode, 0),
	}

	parent.Childs = append(parent.Childs, blockNode)

	// Call ApproveBlock to check and set checkpoint
	bc.ApproveBlock(blockNode)

	return nil
}

func (bc *Blockchain) ApproveBlock(blockNode *BlockNode) {
	if blockNode.Block.Index%10 == 0 {
		blockNode.Block.Checkpoint = true
	}
}
func (bc *Blockchain) ValidateBlock(block *Block, parentBlock *Block) error {
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
func (bc *Blockchain) convertToBlockNodes(blocks []*Block) []*BlockNode {
	blockNodes := make([]*BlockNode, len(blocks))
	for i, block := range blocks {
		blockNodes[i] = &BlockNode{
			Block:  block,
			Parent: nil, // You need to set the correct parent here
			Childs: make([]*BlockNode, 0),
		}
	}
	return blockNodes
}
func (bc *Blockchain) ReplaceBlocks(blocks []*Block) {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	blockNodes := bc.convertToBlockNodes(blocks)
	bc.root = blockNodes[0] // Assuming the first block is the root
}
func (bc *Blockchain) BlockExists(hash []byte) bool {
	return bc.GetBlock(hash) != nil
}
func (bc *Blockchain) traverseTree(callback func(node *BlockNode) bool) {
	var queue []*BlockNode

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
func (bc *Blockchain) GetBlock(hash []byte) *BlockNode {
	var foundNode *BlockNode
	bc.traverseTree(func(node *BlockNode) bool {
		calculatedHash := node.Block.CalculateHash()
		if bytes.Equal(calculatedHash, hash) {
			foundNode = node
			return true
		}
		return false
	})
	return foundNode
}
func (bc *Blockchain) GetLatestBlock() *Block {
	var longestPath []*BlockNode
	bc.traverseTree(func(node *BlockNode) bool {
		if len(node.Childs) > len(longestPath) {
			longestPath = node.Childs
		}
		return false
	})
	if len(longestPath) == 0 {
		return bc.root.Block
	}
	return longestPath[len(longestPath)-1].Block
}

func (bc *Blockchain) GenerateNewBlock(transaction []Transaction) *Block {
	latestBlock := bc.GetLatestBlock()
	newBlock := &Block{
		Index:        latestBlock.Index + 1,
		Timestamp:    uint64(time.Now().Unix()),
		Transactions: transaction,
		PreviousHash: latestBlock.CalculateHash(),
		Data:         0,
	}
	return newBlock
}
