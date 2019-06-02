package ssz

import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

type SSZBasic struct {
	Length  uint32
	Encoder EncoderFn
	Decoder DecoderFn
	HTR     HashTreeRootFn
}

func (v *SSZBasic) FixedLen() uint32 {
	return v.Length
}

func (v *SSZBasic) IsFixed() bool {
	return true
}

func (v *SSZBasic) Encode(eb *sszEncBuf, p unsafe.Pointer) {
	v.Encoder(eb, p)
}

func (v *SSZBasic) Decode(dr *SSZDecReader, p unsafe.Pointer) error {
	return v.Decoder(dr, p)
}

func (v *SSZBasic) HashTreeRoot(h *Hasher, pointer unsafe.Pointer) []byte {
	return v.HTR(h, pointer)
}

var sszBool = &SSZBasic{
	Length: 1,
	Encoder: func(eb *sszEncBuf, p unsafe.Pointer) {
		if *(*bool)(p) {
			eb.buffer.WriteByte(0x01)
		} else {
			eb.buffer.WriteByte(0x00)
		}
	},
	Decoder: func(dr *SSZDecReader, p unsafe.Pointer) error {
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
	HTR: func(h *Hasher, p unsafe.Pointer) []byte {
		scratch := h.Scratch[:1]
		if *(*bool)(p) {
			scratch[0] = 0x01
		} else {
			scratch[0] = 0x00
		}
		return h.Hash(scratch)
	},
}

var sszUint8 = &SSZBasic{
	Length: 1,
	Encoder: func(eb *sszEncBuf, p unsafe.Pointer) {
		eb.buffer.WriteByte(*(*byte)(p))
	},
	Decoder: func(dr *SSZDecReader, p unsafe.Pointer) error {
		b, err := dr.ReadByte()
		if err != nil {
			return err
		}
		*(*byte)(p) = b
		return nil
	},
	HTR: func(h *Hasher, p unsafe.Pointer) []byte {
		scratch := h.Scratch[:1]
		scratch[0] = *(*byte)(p)
		return h.Hash(scratch)
	},
}

var sszUint16 = &SSZBasic{
	Length: 2,
	Encoder: func(eb *sszEncBuf, p unsafe.Pointer) {
		binary.LittleEndian.PutUint16(eb.NextBytes(2), *(*uint16)(p))
	},
	Decoder: func(dr *SSZDecReader, p unsafe.Pointer) error {
		v, err := dr.readUint16()
		if err != nil {
			return err
		}
		*(*uint16)(p) = v
		return nil
	},
	HTR: func(h *Hasher, p unsafe.Pointer) []byte {
		scratch := h.Scratch[:2]
		binary.LittleEndian.PutUint16(scratch, *(*uint16)(p))
		return h.Hash(scratch)
	},
}

var sszUint32 = &SSZBasic{
	Length: 4,
	Encoder: func(eb *sszEncBuf, p unsafe.Pointer) {
		binary.LittleEndian.PutUint32(eb.NextBytes(4), *(*uint32)(p))
	},
	Decoder: func(dr *SSZDecReader, p unsafe.Pointer) error {
		v, err := dr.readUint32()
		if err != nil {
			return err
		}
		*(*uint32)(p) = v
		return nil
	},
	HTR: func(h *Hasher, p unsafe.Pointer) []byte {
		scratch := h.Scratch[:4]
		binary.LittleEndian.PutUint32(scratch, *(*uint32)(p))
		return h.Hash(scratch)
	},
}

var sszUint64 = &SSZBasic{
	Length: 8,
	Encoder: func(eb *sszEncBuf, p unsafe.Pointer) {
		binary.LittleEndian.PutUint64(eb.NextBytes(8), *(*uint64)(p))
	},
	Decoder: func(dr *SSZDecReader, p unsafe.Pointer) error {
		v, err := dr.readUint64()
		if err != nil {
			return err
		}
		*(*uint64)(p) = v
		return nil
	},
	HTR: func(h *Hasher, p unsafe.Pointer) []byte {
		scratch := h.Scratch[:8]
		binary.LittleEndian.PutUint64(scratch, *(*uint64)(p))
		return h.Hash(scratch)
	},
}
