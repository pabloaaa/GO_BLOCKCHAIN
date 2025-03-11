package tests

import (
	"crypto/sha256"
	"reflect"
	"testing"

	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

func setup() *types.Block {
	return &types.Block{
		Index:        1,
		Timestamp:    123456789,
		PreviousHash: []byte("previousHash"),
		Data:         0,
		Checkpoint:   false,
	}
}

func TestCalculateHash(t *testing.T) {
	block := setup()
	expectedHash := sha256.Sum256([]byte("1123456789previousHash0"))
	calculatedHash := block.CalculateHash()

	if !reflect.DeepEqual(calculatedHash, expectedHash[:]) {
		t.Errorf("Expected hash %x, but got %x", expectedHash, calculatedHash)
	}
}

func TestBlockFromProto(t *testing.T) {
	pbBlock := &pb.Block{
		Index:        1,
		Timestamp:    123456789,
		PreviousHash: []byte("previousHash"),
		Hash:         []byte("hash"),
		Data:         0,
		Checkpoint:   true,
	}
	block := types.BlockFromProto(pbBlock)

	if block.Index != pbBlock.GetIndex() {
		t.Errorf("Expected %d, got %d", pbBlock.GetIndex(), block.Index)
	}
	if block.Timestamp != pbBlock.GetTimestamp() {
		t.Errorf("Expected %d, got %d", pbBlock.GetTimestamp(), block.Timestamp)
	}
	if string(block.PreviousHash) != string(pbBlock.GetPreviousHash()) {
		t.Errorf("Expected %s, got %s", pbBlock.GetPreviousHash(), block.PreviousHash)
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
	if pbBlock.GetData() != block.Data {
		t.Errorf("Expected %d, got %d", block.Data, pbBlock.GetData())
	}
	if pbBlock.GetCheckpoint() != block.Checkpoint {
		t.Errorf("Expected %v, got %v", block.Checkpoint, pbBlock.GetCheckpoint())
	}
}
