package core

import (
	"fmt"
	"sync"

	"github.com/go-kit/log"
)

type Blockchain struct {
	logger    log.Logger
	store     Storage
	lock      sync.RWMutex
	headers   []*Header
	validator Validator

	contractState *State
}

func NewBlockchain(l log.Logger, genesis *Block) (*Blockchain, error) {

	bc := &Blockchain{
		// 合约状态存储
		contractState: NewState(),
		headers:       []*Header{},
		store:         NewMemorystore(),
		logger:        l,
	}
	bc.validator = NewBlockchainValidator(bc)

	err := bc.addBlockWithoutValidation(genesis)
	return bc, err
}

func (bc *Blockchain) SetValidator(v Validator) {
	bc.validator = v
}

func (bc *Blockchain) AddBlock(b *Block) error {
	if err := bc.validator.ValidateBlock(b); err != nil {
		return err
	}

	// 执行合约内的交易
	for _, tx := range b.Transactions {
		bc.logger.Log("msg", "executing code", "len", len(tx.Data), "hash", tx.Hash(&TxHasher{}).ToHexString())

		vm := NewVM(tx.Data, bc.contractState)
		if err := vm.Run(); err != nil {
			return err
		}

		// result := vm.stack.Pop()
		// bc.logger.Log("vm result: " , result)
		fmt.Printf("STATE: %+v\n", vm.contractState)
	}

	return bc.addBlockWithoutValidation(b)
}

// 拿到区块头
func (bc *Blockchain) GetHeader(height uint32) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("given height (%d) too high", height)
	}

	bc.lock.RLock()
	defer bc.lock.RUnlock()
	return bc.headers[height], nil
}

func (bc *Blockchain) HasBlock(height uint32) bool {

	return height <= bc.Height()
}

// [0,1,2,3] => 4 len
func (bc *Blockchain) Height() uint32 {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	return uint32(len(bc.headers) - 1)
}

func (bc *Blockchain) addBlockWithoutValidation(b *Block) error {
	bc.lock.Lock()
	bc.headers = append(bc.headers, b.Header)
	bc.lock.Unlock()

	bc.logger.Log(
		"msg", "new block",
		"hash", b.Hash(BlockHasher{}).ToHexString(),
		"height", b.Height,
		"transactions", len(b.Transactions),
	)

	return bc.store.Put(b)
}
