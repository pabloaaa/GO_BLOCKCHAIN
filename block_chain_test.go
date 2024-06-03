package main

import (
	"testing"
)

func setupBlockchain() *Blockchain {
	bc := NewBlockchain()
	block := NewBlock(1, bc.CurrentTimestamp(), make([]Transaction, 0), "0", 0)
	bc.AddBlock(*block)
	return bc
}

func TestNewBlockchain(t *testing.T) {
	bc := NewBlockchain()

	if len(bc.chain) != 1 {
		t.Errorf("Expected blockchain length to be 1, but got %d", len(bc.chain))
	}

	if bc.chain[0].Index != 0 {
		t.Errorf("Expected genesis block index to be 0, but got %d", bc.chain[0].Index)
	}
}

func TestAddBlock(t *testing.T) {
	bc := setupBlockchain()

	block := NewBlock(2, bc.CurrentTimestamp(), make([]Transaction, 0), "0", 0)
	bc.AddBlock(*block)

	if len(bc.chain) != 3 {
		t.Errorf("Expected blockchain length to be 3, but got %d", len(bc.chain))
	}

	if bc.chain[2].Index != 2 {
		t.Errorf("Expected new block index to be 2, but got %d", bc.chain[2].Index)
	}
}

func TestLast(t *testing.T) {
	bc := setupBlockchain()

	lastBlock := bc.Last()

	if lastBlock.Index != 1 {
		t.Errorf("Expected last block index to be 1, but got %d", lastBlock.Index)
	}
}

func TestLen(t *testing.T) {
	bc := setupBlockchain()

	if bc.len() != 2 {
		t.Errorf("Expected blockchain length to be 2, but got %d", bc.len())
	}

	block := NewBlock(2, bc.CurrentTimestamp(), make([]Transaction, 0), "0", 0)
	bc.AddBlock(*block)

	if bc.len() != 3 {
		t.Errorf("Expected blockchain length to be 3, but got %d", bc.len())
	}
}

func TestGetBlocks(t *testing.T) {
	bc := setupBlockchain()

	blocks := bc.GetBlocks()

	if len(blocks) != 2 {
		t.Errorf("Expected blockchain length to be 2, but got %d", len(blocks))
	}

	if blocks[1].Index != 1 {
		t.Errorf("Expected block index to be 1, but got %d", blocks[1].Index)
	}
}
