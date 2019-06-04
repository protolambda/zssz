package types

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/util/endianness"
	"reflect"
	"unsafe"
)

type SSZBasicVector struct {
	elemKind reflect.Kind
	elemSSZ  *SSZBasic
	length   uint32
	fixedLen uint32
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
	if elemSSZ.Length != uint32(elemTyp.Size()) {
		return nil, fmt.Errorf("basic element type has different size than SSZ type unexpectedly, ssz: %d, go: %d", elemSSZ.Length, elemTyp.Size())
	}
	length := uint32(typ.Len())

	res := &SSZBasicVector{
		elemKind: elemKind,
		elemSSZ:  elemSSZ,
		length:   length,
		fixedLen: length * elemSSZ.Length,
	}
	return res, nil
}

func (v *SSZBasicVector) FuzzReqLen() uint32 {
	// equal to fixed length
	return v.fixedLen
}

func (v *SSZBasicVector) MinLen() uint32 {
	return v.fixedLen
}

func (v *SSZBasicVector) FixedLen() uint32 {
	return v.fixedLen
}

func (v *SSZBasicVector) IsFixed() bool {
	return true
}

func (v *SSZBasicVector) Encode(eb *EncodingBuffer, p unsafe.Pointer) {
	// we can just write the data as-is in a few contexts:
	// - if we're in a little endian architecture
	// - if there is no endianness to deal with
	if endianness.IsLittleEndian || v.elemSSZ.Length == 1 {
		LittleEndianBasicSeriesEncode(eb, p, v.fixedLen)
	} else {
		EncodeFixedSeries(v.elemSSZ.Encoder, v.length, uintptr(v.elemSSZ.Length), eb, p)
	}
}

func (v *SSZBasicVector) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	if endianness.IsLittleEndian || v.elemSSZ.Length == 1 {
		return LittleEndianBasicSeriesDecode(dr, p, v.fixedLen, v.elemKind == reflect.Bool)
	} else {
		return DecodeFixedSeries(v.elemSSZ.Decoder, v.fixedLen, uintptr(v.elemSSZ.FixedLen()), dr, p)
	}
}

func (v *SSZBasicVector) HashTreeRoot(h *Hasher, p unsafe.Pointer) [32]byte {
	if endianness.IsLittleEndian || v.elemSSZ.Length == 1 {
		return LittleEndianBasicSeriesHTR(h, p, v.length, v.fixedLen, v.elemSSZ.ChunkPow)
	} else {
		return BigEndianBasicSeriesHTR(h, p, v.length, v.fixedLen, v.elemSSZ.ChunkPow)
	}
}
