package main

import (
	"errors"
	"strings"
)

type BlockValidator struct{}

func NewBlockValidator() *BlockValidator {
	return &BlockValidator{}
}

func (bv *BlockValidator) ValidateAndAddBlock(block *Block, blockchain *Blockchain) error {

	lastBlock := blockchain.Last()

	if block.Index != lastBlock.Index+1 {
		return errors.New("Block index is not valid")
	}

	if block.PreviousHash != lastBlock.Hash {
		return errors.New("Previous hash is not valid")
	}

	if !strings.HasPrefix(block.Hash, "00") {
		return errors.New("Block hash is not valid")
	}

	blockchain.AddBlock(*block) // Dereference block here

	return nil
}
