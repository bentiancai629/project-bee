package core

import (
	"encoding/gob"
	"io"
)

//
// For now we GOB encoding is used for fast bootstrapping of the project
// in a later phase I'm considering using Protobuffers as default encoding / decoding.
//

type Encoder[T any] interface {
	Encode(T) error
}

type Decoder[T any] interface {
	Decode(T) error
}

type GobTxEncoder struct {
	W io.Writer
}

func NewGobTxEncoder(w io.Writer) *GobTxEncoder {
	return &GobTxEncoder{
		W: w,
	}
}

func (e *GobTxEncoder) Encode(tx *Transaction) error {
	return gob.NewEncoder(e.W).Encode(tx)
}

type GobTxDecoder struct {
	R io.Reader
}

func NewGobTxDecoder(r io.Reader) *GobTxDecoder {
	return &GobTxDecoder{
		R: r,
	}
}

func (e *GobTxDecoder) Decode(tx *Transaction) error {
	return gob.NewDecoder(e.R).Decode(tx)
}

type GobBlockEncoder struct {
	w io.Writer
}

func NewGobBlockEncoder(w io.Writer) *GobBlockEncoder {
	return &GobBlockEncoder{
		w: w,
	}
}

// 序列化区块
func (enc *GobBlockEncoder) Encode(b *Block) error {
	return gob.NewEncoder(enc.w).Encode(b)
}

type GobBlockDecoder struct {
	r io.Reader
}

func NewGobBlockDecoder(r io.Reader) *GobBlockDecoder {
	return &GobBlockDecoder{
		r: r,
	}
}

// 反序列化区块
func (dec *GobBlockDecoder) Decode(b *Block) error {
	return gob.NewDecoder(dec.r).Decode(b)
}