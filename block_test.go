package main

import (
	"testing"
)

func TestNewBlock(t *testing.T) {
	transactions := []Transaction{
		{
			Sender:   "Alice",
			Receiver: "Bob",
			Amount:   10.0,
		},
	}

	tests := []struct {
		name           string
		index          uint64
		timestamp      uint64
		transactions   []Transaction
		previousHash   string
		nonce          uint64
		expectedIndex  uint64
		expectedHash   string
		expectedPrevH  string
		expectedNonce  uint64
		expectedLength int
	}{
		{
			name:           "Test New Block",
			index:          1,
			timestamp:      123456789,
			transactions:   transactions,
			previousHash:   "previousHash",
			nonce:          0,
			expectedIndex:  1,
			expectedPrevH:  "previousHash",
			expectedNonce:  0,
			expectedLength: 1,
		},
		// Add more test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := NewBlock(tt.index, tt.timestamp, tt.transactions, tt.previousHash, tt.nonce)

			if block.Index != tt.expectedIndex {
				t.Errorf("Expected block index to be %d, but got %d", tt.expectedIndex, block.Index)
			}

			if block.Timestamp != tt.timestamp {
				t.Errorf("Expected block timestamp to be %d, but got %d", tt.timestamp, block.Timestamp)
			}

			if block.PreviousHash != tt.expectedPrevH {
				t.Errorf("Expected block previous hash to be '%s', but got %s", tt.expectedPrevH, block.PreviousHash)
			}

			if len(block.Transactions) != tt.expectedLength {
				t.Errorf("Expected block to have %d transaction, but got %d", tt.expectedLength, len(block.Transactions))
			}

			if block.Data != tt.expectedNonce {
				t.Errorf("Expected block data to be %d, but got %d", tt.expectedNonce, block.Data)
			}

			if block.Hash == "" {
				t.Errorf("Expected block hash to be calculated, but got an empty string")
			}
		})
	}
}
