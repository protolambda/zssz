package types

import (
	"encoding/binary"
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	"reflect"
	"unsafe"
)

type BasicHtrFn func(pointer unsafe.Pointer) [32]byte

type SSZBasic struct {
	Length  uint64
	Encoder EncoderFn
	Decoder DecoderFn
	HTR     BasicHtrFn
}

func (v *SSZBasic) FuzzMinLen() uint64 {
	return v.Length
}

func (v *SSZBasic) FuzzMaxLen() uint64 {
	return v.Length
}

func (v *SSZBasic) MinLen() uint64 {
	return v.Length
}

func (v *SSZBasic) MaxLen() uint64 {
	return v.Length
}

func (v *SSZBasic) FixedLen() uint64 {
	return v.Length
}

func (v *SSZBasic) IsFixed() bool {
	return true
}

func (v *SSZBasic) SizeOf(p unsafe.Pointer) uint64 {
	return v.Length
}

func (v *SSZBasic) Encode(eb *EncodingBuffer, p unsafe.Pointer) {
	v.Encoder(eb, p)
}

func (v *SSZBasic) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	return v.Decoder(dr, p)
}

func (v *SSZBasic) HashTreeRoot(h HashFn, pointer unsafe.Pointer) [32]byte {
	return v.HTR(pointer)
}

var sszBool = &SSZBasic{
	Length: 1,
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
			if dr.IsFuzzMode() {
				// just make a valid random bool
				*(*bool)(p) = b&1 != 0
				return nil
			} else {
				return fmt.Errorf("bool value is invalid")
			}
		}
	},
	HTR: func(p unsafe.Pointer) (out [32]byte) {
		out[0] = *(*byte)(p)
		return
	},
}

var sszUint8 = &SSZBasic{
	Length: 1,
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
	HTR: func(p unsafe.Pointer) (out [32]byte) {
		out[0] = *(*byte)(p)
		return
	},
}

var sszUint16 = &SSZBasic{
	Length: 2,
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
	HTR: func(p unsafe.Pointer) (out [32]byte) {
		binary.LittleEndian.PutUint16(out[:], *(*uint16)(p))
		return
	},
}

var sszUint32 = &SSZBasic{
	Length: 4,
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
	HTR: func(p unsafe.Pointer) (out [32]byte) {
		binary.LittleEndian.PutUint32(out[:], *(*uint32)(p))
		return
	},
}

var sszUint64 = &SSZBasic{
	Length: 8,
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
	HTR: func(p unsafe.Pointer) (out [32]byte) {
		binary.LittleEndian.PutUint64(out[:], *(*uint64)(p))
		return
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
