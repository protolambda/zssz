package types

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"unsafe"
	. "zssz/dec"
	. "zssz/enc"
	. "zssz/htr"
)

type SSZBasic struct {
	Length   uint32
	// 1 << ChunkPow == items of this basic type per chunk
	ChunkPow uint8
	Encoder  EncoderFn
	Decoder  DecoderFn
	HTR      HashTreeRootFn
}

func (v *SSZBasic) FixedLen() uint32 {
	return v.Length
}

func (v *SSZBasic) IsFixed() bool {
	return true
}

func (v *SSZBasic) Encode(eb *EncodingBuffer, p unsafe.Pointer) {
	v.Encoder(eb, p)
}

func (v *SSZBasic) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	return v.Decoder(dr, p)
}

func (v *SSZBasic) HashTreeRoot(h *Hasher, pointer unsafe.Pointer) [32]byte {
	return v.HTR(h, pointer)
}

var sszBool = &SSZBasic{
	Length: 1,
	ChunkPow: 5,
	Encoder: func(eb *EncodingBuffer, p unsafe.Pointer) {
		eb.WriteByte(*(*byte)(p))
	},
	Decoder: func(dr *DecodingReader, p unsafe.Pointer) error {
		b, err := dr.ReadByte()
		if err != nil {
			return err
		}
		if b == 0x00 {
			*(*bool)(p) = false
			return nil
		} else if b == 0x01 {
			*(*bool)(p) = true
			return nil
		} else {
			return fmt.Errorf("bool value is invalid")
		}
	},
	HTR: func(h *Hasher, p unsafe.Pointer) [32]byte {
		d := [1]byte{}
		d[0] = *(*byte)(p)
		return h.Hash(d[:])
	},
}

var sszUint8 = &SSZBasic{
	Length: 1,
	ChunkPow: 5,
	Encoder: func(eb *EncodingBuffer, p unsafe.Pointer) {
		eb.WriteByte(*(*byte)(p))
	},
	Decoder: func(dr *DecodingReader, p unsafe.Pointer) error {
		b, err := dr.ReadByte()
		if err != nil {
			return err
		}
		*(*byte)(p) = b
		return nil
	},
	HTR: func(h *Hasher, p unsafe.Pointer) [32]byte {
		d := [1]byte{}
		d[0] = *(*byte)(p)
		return h.Hash(d[:])
	},
}

var sszUint16 = &SSZBasic{
	Length: 2,
	ChunkPow: 4,
	Encoder: func(eb *EncodingBuffer, p unsafe.Pointer) {
		v := [2]byte{}
		binary.LittleEndian.PutUint16(v[:], *(*uint16)(p))
		eb.Write(v[:])
	},
	Decoder: func(dr *DecodingReader, p unsafe.Pointer) error {
		v, err := dr.ReadUint16()
		if err != nil {
			return err
		}
		*(*uint16)(p) = v
		return nil
	},
	HTR: func(h *Hasher, p unsafe.Pointer) [32]byte {
		d := [2]byte{}
		binary.LittleEndian.PutUint16(d[:], *(*uint16)(p))
		return h.Hash(d[:])
	},
}

var sszUint32 = &SSZBasic{
	Length: 4,
	ChunkPow: 3,
	Encoder: func(eb *EncodingBuffer, p unsafe.Pointer) {
		v := [4]byte{}
		binary.LittleEndian.PutUint32(v[:], *(*uint32)(p))
		eb.Write(v[:])
	},
	Decoder: func(dr *DecodingReader, p unsafe.Pointer) error {
		v, err := dr.ReadUint32()
		if err != nil {
			return err
		}
		*(*uint32)(p) = v
		return nil
	},
	HTR: func(h *Hasher, p unsafe.Pointer) [32]byte {
		d := [4]byte{}
		binary.LittleEndian.PutUint32(d[:], *(*uint32)(p))
		return h.Hash(d[:])
	},
}

var sszUint64 = &SSZBasic{
	Length: 8,
	ChunkPow: 2,
	Encoder: func(eb *EncodingBuffer, p unsafe.Pointer) {
		v := [8]byte{}
		binary.LittleEndian.PutUint64(v[:], *(*uint64)(p))
		eb.Write(v[:])
	},
	Decoder: func(dr *DecodingReader, p unsafe.Pointer) error {
		v, err := dr.ReadUint64()
		if err != nil {
			return err
		}
		*(*uint64)(p) = v
		return nil
	},
	HTR: func(h *Hasher, p unsafe.Pointer) [32]byte {
		d := [8]byte{}
		binary.LittleEndian.PutUint64(d[:], *(*uint64)(p))
		return h.Hash(d[:])
	},
}

func GetBasicSSZElemType(kind reflect.Kind) (*SSZBasic, error) {
	switch kind {
	case reflect.Bool:
		return sszBool, nil
	case reflect.Uint8:
		return sszUint8, nil
	case reflect.Uint16:
		return sszUint16, nil
	case reflect.Uint32:
		return sszUint32, nil
	case reflect.Uint64:
		return sszUint64, nil
	default:
		return nil, fmt.Errorf("kind %d is not a basic type", kind)
	}
}
