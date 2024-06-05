package main

import (
	"testing"

	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

func setup() *Block {
	transactions := []Transaction{
		{
			Sender:   "Alice",
			Receiver: "Bob",
			Amount:   10.0,
		},
	}
	return NewBlock(1, 123456789, transactions, "previousHash", 0)
}

func TestNewBlock(t *testing.T) {
	block := setup()

	if block.Index != 1 {
		t.Errorf("Expected block index to be %d, but got %d", 1, block.Index)
	}

	if block.Timestamp != 123456789 {
		t.Errorf("Expected block timestamp to be %d, but got %d", 123456789, block.Timestamp)
	}

	if block.PreviousHash != "previousHash" {
		t.Errorf("Expected block previous hash to be '%s', but got %s", "previousHash", block.PreviousHash)
	}

	if len(block.Transactions) != 1 {
		t.Errorf("Expected block to have %d transaction, but got %d", 1, len(block.Transactions))
	}

	if block.Data != 0 {
		t.Errorf("Expected block data to be %d, but got %d", 0, block.Data)
	}

	if block.Hash == "" {
		t.Errorf("Expected block hash to be calculated, but got an empty string")
	}
}

func TestCalculateHash(t *testing.T) {
	block := setup()
	originalHash := block.Hash

	// Change the state of the block
	block.Data++
	block.calculateHash()
	if block.Hash == "" || block.Hash == originalHash {
		t.Errorf("Expected block hash to be calculated, but got %s", block.Hash)
	}

	// Reset block state and test with different field
	block = setup()
	originalHash = block.Hash
	block.Index++
	block.calculateHash()
	if block.Hash == "" || block.Hash == originalHash {
		t.Errorf("Expected block hash to be calculated, but got %s", block.Hash)
	}

	// Repeat for other fields as needed...
}

func TestBlockFromProto(t *testing.T) {
	pbBlock := &pb.Block{
		Index:        1,
		Timestamp:    123456789,
		PreviousHash: "previousHash",
		Hash:         "hash",
		Transactions: []*pb.Transaction{
			{
				Sender:   "Alice",
				Receiver: "Bob",
				Amount:   10.0,
			},
		},
		Data: 0,
	}
	block := BlockFromProto(pbBlock)

	if block.Index != pbBlock.GetIndex() {
		t.Errorf("Expected %d, got %d", pbBlock.GetIndex(), block.Index)
	}
	if block.Timestamp != pbBlock.GetTimestamp() {
		t.Errorf("Expected %d, got %d", pbBlock.GetTimestamp(), block.Timestamp)
	}
	if block.PreviousHash != pbBlock.GetPreviousHash() {
		t.Errorf("Expected %s, got %s", pbBlock.GetPreviousHash(), block.PreviousHash)
	}
	if block.Hash != pbBlock.GetHash() {
		t.Errorf("Expected %s, got %s", pbBlock.GetHash(), block.Hash)
	}
	if len(block.Transactions) != len(pbBlock.GetTransactions()) {
		t.Errorf("Expected %d, got %d", len(pbBlock.GetTransactions()), len(block.Transactions))
	}
	if block.Data != pbBlock.GetData() {
		t.Errorf("Expected %d, got %d", pbBlock.GetData(), block.Data)
	}
}

func TestToProto(t *testing.T) {
	block := setup()
	pbBlock := block.ToProto()

	if pbBlock.GetIndex() != block.Index {
		t.Errorf("Expected %d, got %d", block.Index, pbBlock.GetIndex())
	}
	if pbBlock.GetTimestamp() != block.Timestamp {
		t.Errorf("Expected %d, got %d", block.Timestamp, pbBlock.GetTimestamp())
	}
	if pbBlock.GetPreviousHash() != block.PreviousHash {
		t.Errorf("Expected %s, got %s", block.PreviousHash, pbBlock.GetPreviousHash())
	}
	if pbBlock.GetHash() != block.Hash {
		t.Errorf("Expected %s, got %s", block.Hash, pbBlock.GetHash())
	}
	if len(pbBlock.GetTransactions()) != len(block.Transactions) {
		t.Errorf("Expected %d, got %d", len(block.Transactions), len(pbBlock.GetTransactions()))
	}
	if pbBlock.GetData() != block.Data {
		t.Errorf("Expected %d, got %d", block.Data, pbBlock.GetData())
	}
}
