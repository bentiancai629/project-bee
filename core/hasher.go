package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"

	"project-bee/types"
)

type Hasher[T any] interface {
	Hash(T) types.Hash
}

type BlockHasher struct{}

// 椭圆曲线签名 对 bytes 进行哈西
func (BlockHasher) Hash(b *Header) types.Hash {
	h := sha256.Sum256(b.Bytes())
	return types.Hash(h)
}

type TxHasher struct{}

// 对 txData 进行哈希
func (TxHasher) Hash(tx *Transaction) types.Hash {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, tx.Data)
	binary.Write(buf, binary.LittleEndian, tx.To)
	binary.Write(buf, binary.LittleEndian, tx.Value)
	binary.Write(buf, binary.LittleEndian, tx.From)
	binary.Write(buf, binary.LittleEndian, tx.Nonce)

	return types.Hash(sha256.Sum256(buf.Bytes()))
}
