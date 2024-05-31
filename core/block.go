package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"time"

	"project-bee/crypto"
	"project-bee/types"
)

type Header struct {
	Version       uint32
	DataHash      types.Hash
	PrevBlockHash types.Hash
	Height        uint32
	Timestamp     int64
}

// header 头序列化成2进制[]byte
func (h *Header) Bytes() []byte {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	enc.Encode(h)

	return buf.Bytes()
}

type Block struct {
	*Header

	Transactions []*Transaction
	Validator    crypto.PublicKey
	Signature    *crypto.Signature

	// 区块头 hash 缓存
	hash types.Hash
}

func NewBlock(h *Header, txs []*Transaction) (*Block, error) {
	return &Block{
		Header:       h,
		Transactions: txs,
	}, nil
}

func NewBlockFromPrevHeader(prevHeader *Header, txs []*Transaction) (*Block, error) {
	dataHash, err := CalculateDataHash(txs)
	if err != nil {
		return nil, err
	}

	header := &Header{
		Version:       1,
		Height:        prevHeader.Height + 1, // 新高度
		DataHash:      dataHash,
		PrevBlockHash: BlockHasher{}.Hash(prevHeader),
		Timestamp:     time.Now().UnixNano(),
	}

	return NewBlock(header, txs)
}

func (b *Block) AddTransaction(tx *Transaction) {
	b.Transactions = append(b.Transactions, tx)
	hash, _ := CalculateDataHash(b.Transactions)
	b.DataHash = hash
}

func (b *Block) Sign(privKey crypto.PrivateKey) error {
	sig, err := privKey.Sign(b.Header.Bytes())
	if err != nil {
		return err
	}

	b.Validator = privKey.PublicKey()
	b.Signature = sig

	return nil
}

// 对 signature 和 tx 都要需要 verify
func (b *Block) Verify() error {
	if b.Signature == nil {
		return fmt.Errorf("block has no signature")
	}

	if !b.Signature.Verify(b.Validator, b.Header.Bytes()) {
		return fmt.Errorf("block has invalid signature")
	}

	for _, tx := range b.Transactions {
		if err := tx.Verify(); err != nil {
			return err
		}
	}

	// 验证交易
	dataHash, err := CalculateDataHash(b.Transactions)
	if err != nil {
		return err
	}
	if dataHash != b.DataHash {
		fmt.Printf("dataHash: %v", dataHash)
		fmt.Printf("b.DataHash: %v", b.DataHash)
		return fmt.Errorf("block (%s) has an invalid data hash", b.Hash(BlockHasher{}))
	}

	return nil
}

func (b *Block) Decode(dec Decoder[*Block]) error {
	return dec.Decode(b)
}

func (b *Block) Encode(enc Encoder[*Block]) error {
	return enc.Encode(b)
}

// 对区块头进行 hash, 如果缓存存在直接返回
func (b *Block) Hash(hasher Hasher[*Header]) types.Hash {
	if b.hash.IsZero() {
		b.hash = hasher.Hash(b.Header)
	}

	return b.hash
}
func CalculateDataHash(txs []*Transaction) (hash types.Hash, err error) {
	buf := &bytes.Buffer{}

	for _, tx := range txs {
		if err = tx.Encode(NewGobTxEncoder(buf)); err != nil {
			return
		}
	}
	hash = sha256.Sum256(buf.Bytes())
	return
}