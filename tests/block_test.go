package tests

import (
	"crypto/sha256"
	"reflect"
	"testing"

	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	. "github.com/pabloaaa/GO_BLOCKCHAIN/src"
)

func setup() *Block {
	transactions := []Transaction{
		{
			Sender:   []byte("Alice"),
			Receiver: []byte("Bob"),
			Amount:   10.0,
		},
	}
	return &Block{
		Index:        1,
		Timestamp:    123456789,
		Transactions: transactions,
		PreviousHash: []byte("previousHash"),
		Data:         0,
		Checkpoint:   false,
	}
}

func TestCalculateHash(t *testing.T) {
	block := setup()
	expectedHash := sha256.Sum256([]byte("1123456789AliceBob10previousHash0"))
	calculatedHash := block.CalculateHash()

	if !reflect.DeepEqual(calculatedHash, expectedHash[:]) {
		t.Errorf("Expected hash %x, but got %x", expectedHash, calculatedHash)
	}
}

func TestSetData(t *testing.T) {
	block := setup()
	block.SetData(42)

	if block.Data != 42 {
		t.Errorf("Expected block data to be 42, but got %d", block.Data)
	}
}

func TestBlockFromProto(t *testing.T) {
	pbBlock := &pb.Block{
		Index:        1,
		Timestamp:    123456789,
		PreviousHash: []byte("previousHash"),
		Hash:         []byte("hash"),
		Transactions: []*pb.Transaction{
			{
				Sender:   []byte("Alice"),
				Receiver: []byte("Bob"),
				Amount:   10.0,
			},
		},
		Data:       0,
		Checkpoint: true,
	}
	block := BlockFromProto(pbBlock)

	if block.Index != pbBlock.GetIndex() {
		t.Errorf("Expected %d, got %d", pbBlock.GetIndex(), block.Index)
	}
	if block.Timestamp != pbBlock.GetTimestamp() {
		t.Errorf("Expected %d, got %d", pbBlock.GetTimestamp(), block.Timestamp)
	}
	if string(block.PreviousHash) != string(pbBlock.GetPreviousHash()) {
		t.Errorf("Expected %s, got %s", pbBlock.GetPreviousHash(), block.PreviousHash)
	}
	if len(block.Transactions) != len(pbBlock.GetTransactions()) {
		t.Errorf("Expected %d, got %d", len(pbBlock.GetTransactions()), len(block.Transactions))
	}
	if block.Data != pbBlock.GetData() {
		t.Errorf("Expected %d, got %d", pbBlock.GetData(), block.Data)
	}
	if block.Checkpoint != pbBlock.GetCheckpoint() {
		t.Errorf("Expected %v, got %v", pbBlock.GetCheckpoint(), block.Checkpoint)
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
	if string(pbBlock.GetPreviousHash()) != string(block.PreviousHash) {
		t.Errorf("Expected %s, got %s", block.PreviousHash, pbBlock.GetPreviousHash())
	}
	if len(pbBlock.GetTransactions()) != len(block.Transactions) {
		t.Errorf("Expected %d, got %d", len(block.Transactions), len(pbBlock.GetTransactions()))
	}
	if pbBlock.GetData() != block.Data {
		t.Errorf("Expected %d, got %d", block.Data, pbBlock.GetData())
	}
	if pbBlock.GetCheckpoint() != block.Checkpoint {
		t.Errorf("Expected %v, got %v", block.Checkpoint, pbBlock.GetCheckpoint())
	}
}
