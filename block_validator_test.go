package main

import (
	"testing"
)

func TestValidateAndAddBlock(t *testing.T) {
	tests := []struct {
		name           string
		blockIndex     uint64
		blockTimestamp uint64
		blockHash      string
		blockData      uint64
		expectedError  string
	}{
		{"ValidBlock", 1, 123456790, "00abcdef", 0, ""},
		{"InvalidIndex", 0, 123456790, "00abcdef", 0, "Block index is not valid"},
		{"InvalidPreviousHash", 1, 123456790, "00abcdef", 0, "Previous hash is not valid"},
		{"InvalidHash", 1, 123456790, "abcdef", 0, "Block hash is not valid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bv := NewBlockValidator()
			bc := NewBlockchain()
			block := NewBlock(tt.blockIndex, tt.blockTimestamp, []Transaction{}, bc.Last().Hash, tt.blockData)
			block.Hash = tt.blockHash // Set the hash manually

			if tt.name == "InvalidPreviousHash" {
				block.PreviousHash = "invalid"
			}

			err := bv.ValidateAndAddBlock(block, bc)
			if err != nil {
				if err.Error() != tt.expectedError {
					t.Errorf("Expected '%s' error, but got %v", tt.expectedError, err)
				}
			} else if tt.expectedError != "" {
				t.Errorf("Expected '%s' error, but got no error", tt.expectedError)
			}
		})
	}
}
