package core

import (
	"fmt"
	"sync"

	"project-bee/types"

	"github.com/go-kit/log"
)

type Blockchain struct {
	logger log.Logger
	store  Storage
	// TODO: double check this!
	lock       sync.RWMutex
	headers    []*Header
	blocks     []*Block
	txStore    map[types.Hash]*Transaction
	blockStore map[types.Hash]*Block

	// accountState *AccountState

	stateLock       sync.RWMutex
	collectionState map[types.Hash]*CollectionTx
	mintState       map[types.Hash]*MintTx

	validator Validator

	// TODO: make this an interface.
	contractState *State
}

func NewBlockchain(l log.Logger, genesis *Block) (*Blockchain, error) {

	bc := &Blockchain{
		// 合约状态存储
		contractState:   NewState(),
		headers:         []*Header{},
		store:           NewMemorystore(),
		logger:          l,
		collectionState: make(map[types.Hash]*CollectionTx),
		mintState:       make(map[types.Hash]*MintTx),
		blockStore:      make(map[types.Hash]*Block),
		txStore:         make(map[types.Hash]*Transaction),
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

	bc.stateLock.Lock()
	defer bc.stateLock.Unlock()

	// 执行合约内的交易
	for _, tx := range b.Transactions {
		bc.logger.Log("msg", "executing code", "len", len(tx.Data), "hash", tx.Hash(&TxHasher{}).ToHexString())

		if len(tx.Data) > 0 {
			vm := NewVM(tx.Data, bc.contractState)
			if err := vm.Run(); err != nil {
				return err
			}

			// result := vm.stack.Pop()
			// bc.logger.Log("vm result: ", result)
			// fmt.Printf("STATE: %+v\n", vm.contractState)
		}

		hash := tx.Hash(TxHasher{})
		switch t := tx.TxInner.(type) {
		case CollectionTx:
			bc.collectionState[hash] = &t
			bc.logger.Log("msg", "create a new NFT collection", "hash", hash)
		case MintTx:
			_, ok := bc.collectionState[t.Collection]
			if !ok {
				return fmt.Errorf("collection (%s) does not exist on the blockchain", t.Collection)
			}
			bc.mintState[hash] = &t
		default:
			fmt.Printf("unsupported tx type: %v", t)
		}

	}

	return bc.addBlockWithoutValidation(b)
}

func (bc *Blockchain) GetBlockByHash(hash types.Hash) (*Block, error) {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	block, ok := bc.blockStore[hash]
	if !ok {
		return nil, fmt.Errorf("block with hash (%s) not found", hash)
	}

	return block, nil
}

func (bc *Blockchain) GetBlock(height uint32) (*Block, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("given height (%d) too high", height)
	}

	bc.lock.Lock()
	defer bc.lock.Unlock()

	return bc.blocks[height], nil
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

func (bc *Blockchain) GetTxByHash(hash types.Hash) (*Transaction, error) {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	fmt.Println("len:", len(bc.txStore))
	tx, ok := bc.txStore[hash]
	if !ok {
		return nil, fmt.Errorf("could not find tx with hash (%s)", hash)
	}

	return tx, nil
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

func (bc *Blockchain) handleTransaction(tx *Transaction) error {
	// If we have data inside execute that data on the VM.
	if len(tx.Data) > 0 {
		bc.logger.Log("msg", "executing code", "len", len(tx.Data), "hash", tx.Hash(&TxHasher{}))

		vm := NewVM(tx.Data, bc.contractState)
		if err := vm.Run(); err != nil {
			return err
		}
	}

	// If the txInner of the transaction is not nil we need to handle
	// the native NFT implemtation.
	// if tx.TxInner != nil {
	// 	if err := bc.handleNativeNFT(tx); err != nil {
	// 		return err
	// 	}
	// }

	// // Handle the native transaction here
	// if tx.Value > 0 {
	// 	if err := bc.handleNativeTransfer(tx); err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

func (bc *Blockchain) addBlockWithoutValidation(b *Block) error {

	bc.stateLock.Lock()
	for i := 0; i < len(b.Transactions); i++ {
		if err := bc.handleTransaction(b.Transactions[i]); err != nil {
			bc.logger.Log("error", err.Error())

			b.Transactions[i] = b.Transactions[len(b.Transactions)-1]
			b.Transactions = b.Transactions[:len(b.Transactions)-1]

			continue
		}
	}

	bc.stateLock.Unlock()

	bc.lock.Lock()
	bc.headers = append(bc.headers, b.Header)
	bc.blocks = append(bc.blocks, b)
	bc.lock.Unlock()

	bc.logger.Log(
		"msg", "new block",
		"hash", b.Hash(BlockHasher{}).ToHexString(),
		"height", b.Height,
		"transactions", len(b.Transactions),
	)

	return bc.store.Put(b)
}
