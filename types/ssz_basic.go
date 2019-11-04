package types

import (
	"encoding/binary"
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	. "github.com/protolambda/zssz/pretty"
	"reflect"
	"unsafe"
)

type BasicHtrFn func(pointer unsafe.Pointer) [32]byte

type SSZBasic struct {
	Length     uint64
	Encoder    EncoderFn
	Decoder    DecoderFn
	DryChecker DryCheckFn
	HTR        BasicHtrFn
	PrettyFn   PrettyFn
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

func (v *SSZBasic) Encode(eb *EncodingWriter, p unsafe.Pointer) error {
	return v.Encoder(eb, p)
}

func (v *SSZBasic) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	return v.Decoder(dr, p)
}

func (v *SSZBasic) DryCheck(dr *DecodingReader) error {
	return v.DryChecker(dr)
}

func (v *SSZBasic) HashTreeRoot(h HashFn, pointer unsafe.Pointer) [32]byte {
	return v.HTR(pointer)
}

func (v *SSZBasic) Pretty(indent uint32, w *PrettyWriter, p unsafe.Pointer) {
	v.PrettyFn(indent, w, p)
}

var sszBool = &SSZBasic{
	Length: 1,
	Encoder: func(eb *EncodingWriter, p unsafe.Pointer) error {
		return eb.WriteByte(*(*byte)(p))
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
	DryChecker: func(dr *DecodingReader) error {
		b, err := dr.ReadByte()
		if err != nil {
			return err
		}
		if b > 1 {
			return fmt.Errorf("bool value is invalid")
		}
		return nil
	},
	HTR: func(p unsafe.Pointer) (out [32]byte) {
		out[0] = *(*byte)(p)
		return
	},
	PrettyFn: func(indent uint32, w *PrettyWriter, p unsafe.Pointer) {
		w.WriteIndent(indent)
		if *(*bool)(p) {
			w.Write("True")
		} else {
			w.Write("False")
		}
	},
}

var sszUint8 = &SSZBasic{
	Length: 1,
	Encoder: func(eb *EncodingWriter, p unsafe.Pointer) error {
		return eb.WriteByte(*(*byte)(p))
	},
	Decoder: func(dr *DecodingReader, p unsafe.Pointer) error {
		b, err := dr.ReadByte()
		if err != nil {
			return err
		}
		*(*byte)(p) = b
		return nil
	},
	DryChecker: func(dr *DecodingReader) error {
		_, err := dr.Skip(1)
		return err
	},
	HTR: func(p unsafe.Pointer) (out [32]byte) {
		out[0] = *(*byte)(p)
		return
	},
	PrettyFn: func(indent uint32, w *PrettyWriter, p unsafe.Pointer) {
		w.WriteIndent(indent)
		w.Write(fmt.Sprintf("0x%02x", *(*byte)(p)))
	},
}

var sszUint16 = &SSZBasic{
	Length: 2,
	Encoder: func(eb *EncodingWriter, p unsafe.Pointer) error {
		v := [2]byte{}
		binary.LittleEndian.PutUint16(v[:], *(*uint16)(p))
		return eb.Write(v[:])
	},
	Decoder: func(dr *DecodingReader, p unsafe.Pointer) error {
		v, err := dr.ReadUint16()
		if err != nil {
			return err
		}
		*(*uint16)(p) = v
		return nil
	},
	DryChecker: func(dr *DecodingReader) error {
		_, err := dr.Skip(2)
		return err
	},
	HTR: func(p unsafe.Pointer) (out [32]byte) {
		binary.LittleEndian.PutUint16(out[:], *(*uint16)(p))
		return
	},
	PrettyFn: func(indent uint32, w *PrettyWriter, p unsafe.Pointer) {
		w.WriteIndent(indent)
		if *(*uint16)(p) == ^uint16(0) {
			w.Write("0xFFFF")
		} else {
			w.Write(fmt.Sprintf("%d", *(*uint16)(p)))
		}
	},
}

var sszUint32 = &SSZBasic{
	Length: 4,
	Encoder: func(eb *EncodingWriter, p unsafe.Pointer) error {
		v := [4]byte{}
		binary.LittleEndian.PutUint32(v[:], *(*uint32)(p))
		return eb.Write(v[:])
	},
	Decoder: func(dr *DecodingReader, p unsafe.Pointer) error {
		v, err := dr.ReadUint32()
		if err != nil {
			return err
		}
		*(*uint32)(p) = v
		return nil
	},
	DryChecker: func(dr *DecodingReader) error {
		_, err := dr.Skip(4)
		return err
	},
	HTR: func(p unsafe.Pointer) (out [32]byte) {
		binary.LittleEndian.PutUint32(out[:], *(*uint32)(p))
		return
	},
	PrettyFn: func(indent uint32, w *PrettyWriter, p unsafe.Pointer) {
		w.WriteIndent(indent)
		if *(*uint32)(p) == ^uint32(0) {
			w.Write("0xFFFFFFFF")
		} else {
			w.Write(fmt.Sprintf("%d", *(*uint32)(p)))
		}
	},
}

var sszUint64 = &SSZBasic{
	Length: 8,
	Encoder: func(eb *EncodingWriter, p unsafe.Pointer) error {
		v := [8]byte{}
		binary.LittleEndian.PutUint64(v[:], *(*uint64)(p))
		return eb.Write(v[:])
	},
	Decoder: func(dr *DecodingReader, p unsafe.Pointer) error {
		v, err := dr.ReadUint64()
		if err != nil {
			return err
		}
		*(*uint64)(p) = v
		return nil
	},
	DryChecker: func(dr *DecodingReader) error {
		_, err := dr.Skip(8)
		return err
	},
	HTR: func(p unsafe.Pointer) (out [32]byte) {
		binary.LittleEndian.PutUint64(out[:], *(*uint64)(p))
		return
	},
	PrettyFn: func(indent uint32, w *PrettyWriter, p unsafe.Pointer) {
		w.WriteIndent(indent)
		if *(*uint64)(p) == ^uint64(0) {
			w.Write("0xFFFFFFFFFFFFFFFF")
		} else {
			w.Write(fmt.Sprintf("%d", *(*uint64)(p)))
		}
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
