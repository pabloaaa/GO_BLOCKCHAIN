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

func (bc *Blockchain) LongestChain() []*BlockNode {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	var longestChain []*BlockNode
	var maxLen int

	var traverse func(node *BlockNode, chain []*BlockNode)
	traverse = func(node *BlockNode, chain []*BlockNode) {
		chain = append(chain, node)

		if len(chain) > maxLen {
			maxLen = len(chain)
			longestChain = chain
		}

		for _, child := range node.Childs {
			traverse(child, chain)
		}
	}

	traverse(bc.root, []*BlockNode{})
	return longestChain
}
