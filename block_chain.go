package main

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
	return nil
}
func (bc *Blockchain) ValidateBlock(block *Block, parentBlock *Block) error {
	if block.Index != parentBlock.Index+1 {
		return errors.New("Block index is not valid")
	}

	if !bytes.Equal(block.PreviousHash, parentBlock.calculateHash()) {
		return errors.New("Previous hash is not valid")
	}

	hashPrefix := block.calculateHash()[:2]
	if !bytes.Equal(hashPrefix, []byte("00")) {
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
func (bc *Blockchain) GetLatestBlock() *Block {
	var longestPath []*BlockNode
	var queue [][]*BlockNode

	queue = append(queue, []*BlockNode{bc.root})

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]

		if len(path) > len(longestPath) {
			longestPath = path
		}

		lastNodeInPath := path[len(path)-1]
		for _, child := range lastNodeInPath.Childs {
			newPath := append(path[:], child)
			queue = append(queue, newPath)
		}
	}

	return longestPath[len(longestPath)-1].Block
}
