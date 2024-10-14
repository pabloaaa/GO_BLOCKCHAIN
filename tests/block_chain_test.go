package tests

import (
	"bytes"
	"testing"
	"time"
)

func setupBlockchain() *Blockchain {
	bc := NewBlockchain()
	return bc
}

func TestNewBlockchain(t *testing.T) {
	bc := setupBlockchain()

	if bc.root == nil {
		t.Errorf("Expected root to be initialized, but got nil")
	}

	if bc.root.Block.Index != 0 {
		t.Errorf("Expected genesis block index to be 0, but got %d", bc.root.Block.Index)
	}
}

func generateHardcodedValidBlock(parent *Block) *Block {
	newBlock := &Block{
		Index:        parent.Index + 1,
		Timestamp:    uint64(time.Now().Unix()),
		Transactions: make([]Transaction, 0),
		PreviousHash: parent.calculateHash(),
		Data:         0,
	}
	// Hardcode the hash to match the validation criteria
	for {
		hash := newBlock.calculateHash()
		if bytes.HasPrefix(hash, []byte("000")) {
			break
		}
		newBlock.Data++
	}
	return newBlock
}

func TestAddBlock(t *testing.T) {
	bc := setupBlockchain()
	genesisBlock := bc.root.Block

	newBlock := generateHardcodedValidBlock(genesisBlock)

	err := bc.AddBlock(bc.root, newBlock)
	if err != nil {
		t.Errorf("Failed to add block: %v", err)
	}

	if len(bc.root.Childs) != 1 {
		t.Errorf("Expected 1 child block, but got %d", len(bc.root.Childs))
	}

	if bc.root.Childs[0].Block.Index != 1 {
		t.Errorf("Expected child block index to be 1, but got %d", bc.root.Childs[0].Block.Index)
	}
}

func TestValidateBlock(t *testing.T) {
	bc := setupBlockchain()
	genesisBlock := bc.root.Block

	validBlock := generateHardcodedValidBlock(genesisBlock)

	err := bc.ValidateBlock(validBlock, genesisBlock)
	if err != nil {
		t.Errorf("Expected block to be valid, but got error: %v", err)
	}

	invalidBlock := &Block{
		Index:        2,
		Timestamp:    uint64(time.Now().Unix()),
		Transactions: make([]Transaction, 0),
		PreviousHash: genesisBlock.calculateHash(),
		Data:         0,
	}

	err = bc.ValidateBlock(invalidBlock, genesisBlock)
	if err == nil {
		t.Errorf("Expected block to be invalid due to incorrect index, but got no error")
	}
}

func TestBlockExists(t *testing.T) {
	bc := setupBlockchain()
	genesisBlock := bc.root.Block

	exists := bc.BlockExists(genesisBlock.calculateHash())
	if !exists {
		t.Errorf("Expected genesis block to exist, but it does not")
	}

	nonExistentHash := []byte("nonexistenthash")
	exists = bc.BlockExists(nonExistentHash)
	if exists {
		t.Errorf("Expected block to not exist, but it does")
	}
}

func TestGetBlock(t *testing.T) {
	bc := setupBlockchain()
	genesisBlock := bc.root.Block

	foundBlock := bc.GetBlock(genesisBlock.calculateHash())
	if foundBlock == nil {
		t.Errorf("Expected to find genesis block, but got nil")
	}

	if !bytes.Equal(foundBlock.Block.calculateHash(), genesisBlock.calculateHash()) {
		t.Errorf("Expected to find genesis block, but found a different block")
	}
}

func TestGetLatestBlock(t *testing.T) {
	bc := setupBlockchain()
	genesisBlock := bc.root.Block

	newBlock := generateHardcodedValidBlock(genesisBlock)

	err := bc.AddBlock(bc.root, newBlock)
	if err != nil {
		t.Errorf("Failed to add block: %v", err)
	}

	latestBlock := bc.GetLatestBlock()
	if latestBlock.Index != 1 {
		t.Errorf("Expected latest block index to be 1, but got %d", latestBlock.Index)
	}
}
