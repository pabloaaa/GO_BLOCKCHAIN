package main

import (
	"errors"
	"strings"
)

type BlockValidator struct{}

func NewBlockValidator() *BlockValidator {
	return &BlockValidator{}
}

func (bv *BlockValidator) ValidateAndAddBlock(block Block, blockchain *Blockchain) error {
	blockchain.mux.Lock()
	lastBlock := blockchain.Last()
	blockchain.mux.Unlock()

	if block.Index != lastBlock.Index+1 {
		return errors.New("Block index is not valid")
	}

	if block.PreviousHash != lastBlock.Hash {
		return errors.New("Previous hash is not valid")
	}

	if !strings.HasPrefix(block.Hash, "00") {
		return errors.New("Block hash is not valid")
	}

	blockchain.mux.Lock()
	blockchain.AddBlock(block)
	blockchain.mux.Unlock()

	return nil
}
