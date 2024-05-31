package core

import (
	"bytes"
	"testing"
	"time"

	"project-bee/crypto"
	"project-bee/types"

	"github.com/stretchr/testify/assert"
)

func TestSignBlock(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	b := randomBlock(t, 0, types.Hash{})

	assert.Nil(t, b.Sign(privKey))

	/*
		&crypto.Signature{
			S:112276888059855033686043184703952409124382422214469082514215436935741730210003,
			R:65803880287449092074129389894718869426009876961703930299763152285525383842104
		}
	*/
	assert.NotNil(t, b.Signature)

}

func TestVerifyBlock(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	b := randomBlock(t, 0, types.Hash{})

	assert.Nil(t, b.Sign(privKey))
	assert.Nil(t, b.Verify())

	otherPrivKey := crypto.GeneratePrivateKey()
	b.Validator = otherPrivKey.PublicKey()
	assert.NotNil(t, b.Verify())

	b.Height = 100
	assert.NotNil(t, b.Verify())
}

// 对区块解码编码
func TestDecodeEncodeBlock(t *testing.T) {
	b := randomBlock(t, 1, types.Hash{})
	buf := &bytes.Buffer{}
	assert.Nil(t, b.Encode(NewGobBlockEncoder(buf))) // 编码

	bDecode := new(Block)
	assert.Nil(t, bDecode.Decode(NewGobBlockDecoder(buf))) // 解码

	assert.Equal(t, b, bDecode) // 结果判断相同

	for i := 0; i < len(b.Transactions); i++ {
		b.Transactions[i].hash = types.Hash{}
		assert.Equal(t, b.Transactions[i], bDecode.Transactions[i])
	}

	assert.Equal(t, b.Validator, bDecode.Validator)
	assert.Equal(t, b.Signature, bDecode.Signature)
}

func randomBlock(t *testing.T, height uint32, prevBlockHash types.Hash) *Block {
	privKey := crypto.GeneratePrivateKey()
	tx := randomTxWithSignature(t)

	header := &Header{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		Height:        height,
		Timestamp:     time.Now().UnixNano(),
	}

	b, err := NewBlock(header, []*Transaction{&tx})
	assert.Nil(t, err)
	dataHash, err := CalculateDataHash(b.Transactions)
	assert.Nil(t, err)
	b.Header.DataHash = dataHash
	assert.Nil(t, b.Sign(privKey))

	return b
}
