package main

import (
	"sync"
	"time"
)

type Blockchain struct {
	chain []Block
	mux   sync.Mutex
}

func NewBlockchain() *Blockchain {
	blockchain := &Blockchain{
		chain: make([]Block, 0),
	}
	blockchain.createGenesisBlock()
	return blockchain
}

func (bc *Blockchain) createGenesisBlock() {
	genesisBlock := NewBlock(0, uint64(time.Now().Unix()), make([]Transaction, 0), "0", 0)
	bc.chain = append(bc.chain, *genesisBlock)
}

func (bc *Blockchain) AddBlock(block Block) {
	bc.mux.Lock()
	defer bc.mux.Unlock()
	bc.chain = append(bc.chain, block)
}

func (bc *Blockchain) CurrentTimestamp() uint64 {
	return uint64(time.Now().Unix())
}

func (bc *Blockchain) Last() *Block {
	bc.mux.Lock()
	defer bc.mux.Unlock()
	if len(bc.chain) == 0 {
		return nil
	}
	return &bc.chain[len(bc.chain)-1]
}

func (bc *Blockchain) len() int {
	bc.mux.Lock()
	defer bc.mux.Unlock()
	return len(bc.chain)
}
