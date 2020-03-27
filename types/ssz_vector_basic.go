package types

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	. "github.com/protolambda/zssz/pretty"
	"github.com/protolambda/zssz/util/endianness"
	"reflect"
	"unsafe"
)

type SSZBasicVector struct {
	elemKind reflect.Kind
	elemSSZ  *SSZBasic
	length   uint64
	byteLen  uint64
}

func NewSSZBasicVector(typ reflect.Type) (*SSZBasicVector, error) {
	if typ.Kind() != reflect.Array {
		return nil, fmt.Errorf("typ is not a fixed-length array")
	}
	elemTyp := typ.Elem()
	elemKind := elemTyp.Kind()
	elemSSZ, err := GetBasicSSZElemType(elemKind)
	if err != nil {
		return nil, err
	}
	if elemSSZ.Length != uint64(elemTyp.Size()) {
		return nil, fmt.Errorf("basic element type has different size than SSZ type unexpectedly, ssz: %d, go: %d", elemSSZ.Length, elemTyp.Size())
	}
	length := uint64(typ.Len())

	res := &SSZBasicVector{
		elemKind: elemKind,
		elemSSZ:  elemSSZ,
		length:   length,
		byteLen:  length * elemSSZ.Length,
	}
	return res, nil
}

func (v *SSZBasicVector) FuzzMinLen() uint64 {
	return v.byteLen
}

func (v *SSZBasicVector) FuzzMaxLen() uint64 {
	return v.byteLen
}

func (v *SSZBasicVector) MinLen() uint64 {
	return v.byteLen
}

func (v *SSZBasicVector) MaxLen() uint64 {
	return v.byteLen
}

func (v *SSZBasicVector) FixedLen() uint64 {
	return v.byteLen
}

func (v *SSZBasicVector) IsFixed() bool {
	return true
}

func (v *SSZBasicVector) SizeOf(p unsafe.Pointer) uint64 {
	return v.byteLen
}

func (v *SSZBasicVector) Encode(eb *EncodingWriter, p unsafe.Pointer) error {
	// we can just write the data as-is in a few contexts:
	// - if we're in a little endian architecture
	// - if there is no endianness to deal with
	if endianness.IsLittleEndian || v.elemSSZ.Length == 1 {
		return LittleEndianBasicSeriesEncode(eb, p, v.byteLen)
	} else {
		return EncodeFixedSeries(v.elemSSZ.Encoder, v.length, uintptr(v.elemSSZ.Length), eb, p)
	}
}

func (v *SSZBasicVector) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	if endianness.IsLittleEndian || v.elemSSZ.Length == 1 {
		return LittleEndianBasicSeriesDecode(dr, p, v.byteLen, v.byteLen, v.elemKind == reflect.Bool)
	} else {
		return DecodeFixedSeries(v.elemSSZ.Decoder, v.byteLen, uintptr(v.elemSSZ.FixedLen()), dr, p)
	}
}

func (v *SSZBasicVector) DryCheck(dr *DecodingReader) error {
	return BasicSeriesDryCheck(dr, v.byteLen, v.byteLen, v.elemKind == reflect.Bool)
}

func (v *SSZBasicVector) HashTreeRoot(h Hasher, p unsafe.Pointer) [32]byte {
	if endianness.IsLittleEndian || v.elemSSZ.Length == 1 {
		return LittleEndianBasicSeriesHTR(h, p, v.byteLen, v.byteLen)
	} else {
		return BigEndianBasicSeriesHTR(h, p, v.byteLen, v.byteLen, uint8(v.elemSSZ.Length))
	}
}

func (v *SSZBasicVector) Pretty(indent uint32, w *PrettyWriter, p unsafe.Pointer) {
	w.WriteIndent(indent)
	w.Write("[\n")
	w.WriteIndent(indent + 1)
	CallSeries(func(i uint64, p unsafe.Pointer) {
		v.elemSSZ.Pretty(0, w, p)
		if i == v.length-1 {
			w.Write("\n")
		} else if i%(32/v.elemSSZ.Length) == 0 {
			w.Write(",\n")
			w.WriteIndent(indent + 1)
		} else {
			w.Write(", ")
		}
	}, v.length, uintptr(v.elemSSZ.Length), p)
	w.WriteIndent(indent)
	w.Write("]")
}
