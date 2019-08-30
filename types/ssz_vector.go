package types

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/merkle"
	"reflect"
	"unsafe"
)

type SSZVector struct {
	length      uint64
	elemMemSize uintptr
	elemSSZ     SSZ
	isFixedLen  bool
	fixedLen    uint64
	minLen      uint64
	maxLen      uint64
	fuzzMinLen  uint64
	fuzzMaxLen  uint64
}

func NewSSZVector(factory SSZFactoryFn, typ reflect.Type) (*SSZVector, error) {
	if typ.Kind() != reflect.Array {
		return nil, fmt.Errorf("typ is not a fixed-length array")
	}
	length := uint64(typ.Len())
	elemTyp := typ.Elem()

	elemSSZ, err := factory(elemTyp)
	if err != nil {
		return nil, err
	}
	var fixedElemLen, minElemLen, maxElemLen uint64
	if elemSSZ.IsFixed() {
		fixedElemLen = elemSSZ.FixedLen()
		minElemLen = elemSSZ.MinLen()
		maxElemLen = elemSSZ.MaxLen()
		if fixedElemLen != minElemLen || fixedElemLen != maxElemLen {
			return nil, fmt.Errorf("fixed-size element vector has invalid element min/max length:"+
				" fixed: %d min: %d max: %d ", fixedElemLen, minElemLen, maxElemLen)
		}
	} else {
		fixedElemLen = BYTES_PER_LENGTH_OFFSET
		minElemLen = BYTES_PER_LENGTH_OFFSET + elemSSZ.MinLen()
		maxElemLen = BYTES_PER_LENGTH_OFFSET + elemSSZ.MaxLen()
	}
	res := &SSZVector{
		length:      length,
		elemMemSize: elemTyp.Size(),
		elemSSZ:     elemSSZ,
		isFixedLen:  elemSSZ.IsFixed(),
		fixedLen:    fixedElemLen * length,
		minLen:      minElemLen * length,
		maxLen:      maxElemLen * length,
		fuzzMinLen:  elemSSZ.FuzzMinLen() * length,
		fuzzMaxLen:  elemSSZ.FuzzMaxLen() * length,
	}
	return res, nil
}

func (v *SSZVector) FuzzMinLen() uint64 {
	return v.fuzzMinLen
}

func (v *SSZVector) FuzzMaxLen() uint64 {
	return v.fuzzMaxLen
}

func (v *SSZVector) MinLen() uint64 {
	return v.minLen
}

func (v *SSZVector) MaxLen() uint64 {
	return v.maxLen
}

func (v *SSZVector) FixedLen() uint64 {
	return v.fixedLen
}

func (v *SSZVector) IsFixed() bool {
	return v.isFixedLen
}

func (v *SSZVector) SizeOf(p unsafe.Pointer) uint64 {
	if v.IsFixed() {
		return v.fixedLen
	} else {
		out := v.fixedLen
		memOffset := uintptr(0)
		for i := uint64(0); i < v.length; i++ {
			elemPtr := unsafe.Pointer(uintptr(p) + memOffset)
			memOffset += v.elemMemSize
			out += v.elemSSZ.SizeOf(elemPtr)
		}
		return out
	}
}

func (v *SSZVector) Encode(eb *EncodingWriter, p unsafe.Pointer) error {
	if v.IsFixed() {
		return EncodeFixedSeries(v.elemSSZ.Encode, v.length, v.elemMemSize, eb, p)
	} else {
		return EncodeVarSeries(v.elemSSZ.Encode, v.elemSSZ.SizeOf, v.length, v.elemMemSize, eb, p)
	}
}

func (v *SSZVector) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	if v.IsFixed() {
		return DecodeFixedSeries(v.elemSSZ.Decode, v.length, v.elemMemSize, dr, p)
	} else {
		if dr.IsFuzzMode() {
			return DecodeVarSeriesFuzzMode(v.elemSSZ, v.length, v.elemMemSize, dr, p)
		} else {
			return DecodeVarSeries(v.elemSSZ.Decode, v.length, v.elemMemSize, dr, p)
		}
	}
}

func (v *SSZVector) HashTreeRoot(h HashFn, p unsafe.Pointer) [32]byte {
	elemHtr := v.elemSSZ.HashTreeRoot
	elemSize := v.elemMemSize
	leaf := func(i uint64) []byte {
		v := elemHtr(h, unsafe.Pointer(uintptr(p)+(elemSize*uintptr(i))))
		return v[:]
	}
	return merkle.Merkleize(h, v.length, v.length, leaf)
}
