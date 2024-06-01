package main

import (
	"testing"
)

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
	bc := NewBlockchain()
	block := NewBlock(1, bc.CurrentTimestamp(), make([]Transaction, 0), "0", 0)

	bc.AddBlock(*block)

	if len(bc.chain) != 2 {
		t.Errorf("Expected blockchain length to be 2, but got %d", len(bc.chain))
	}

	if bc.chain[1].Index != 1 {
		t.Errorf("Expected new block index to be 1, but got %d", bc.chain[1].Index)
	}
}

func TestLast(t *testing.T) {
	bc := NewBlockchain()
	block := NewBlock(1, bc.CurrentTimestamp(), make([]Transaction, 0), "0", 0)

	bc.AddBlock(*block)

	lastBlock := bc.Last()

	if lastBlock.Index != 1 {
		t.Errorf("Expected last block index to be 1, but got %d", lastBlock.Index)
	}
}

func TestLen(t *testing.T) {
	bc := NewBlockchain()

	if bc.Len() != 1 {
		t.Errorf("Expected blockchain length to be 1, but got %d", bc.Len())
	}

	block := NewBlock(1, bc.CurrentTimestamp(), make([]Transaction, 0), "0", 0)
	bc.AddBlock(*block)

	if bc.Len() != 2 {
		t.Errorf("Expected blockchain length to be 2, but got %d", bc.Len())
	}
}
