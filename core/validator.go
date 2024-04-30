package core

import "fmt"

type Validator interface {
	ValidateBlock(*Block) error
}

type BlockValidator struct {
	bc *Blockchain
}

func NewBlockchainValidator(bc *Blockchain) *BlockValidator {
	return &BlockValidator{
		bc: bc,
	}
}

func (v *BlockValidator) ValidateBlock(b *Block) error {
	// height <= bc.Height()
	if v.bc.HasBlock(b.Height) {
		return fmt.Errorf("chain already contains block (%d) with hash (%s)", b.Height, b.Hash(BlockHasher{}))
	}

	// !b.Signature.Verify(b.Validator, b.HeaderData())
	if err := b.Verify(); err != nil {
		return err
	}

	return nil
}
