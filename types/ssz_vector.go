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
	fuzzReqLen  uint64
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
	var fixedElemLen, minElemLen uint64
	if elemSSZ.IsFixed() {
		fixedElemLen = elemSSZ.FixedLen()
		minElemLen = elemSSZ.MinLen()
	} else {
		fixedElemLen = uint64(BYTES_PER_LENGTH_OFFSET)
		minElemLen = uint64(BYTES_PER_LENGTH_OFFSET) + elemSSZ.MinLen()
	}
	res := &SSZVector{
		length:      length,
		elemMemSize: elemTyp.Size(),
		elemSSZ:     elemSSZ,
		isFixedLen:  elemSSZ.IsFixed(),
		fixedLen:    fixedElemLen * length,
		minLen:      minElemLen * length,
		fuzzReqLen:  elemSSZ.FuzzReqLen() * length,
	}
	return res, nil
}

func (v *SSZVector) FuzzReqLen() uint64 {
	return v.fuzzReqLen
}

func (v *SSZVector) MinLen() uint64 {
	return v.minLen
}

func (v *SSZVector) FixedLen() uint64 {
	return v.fixedLen
}

func (v *SSZVector) IsFixed() bool {
	return v.isFixedLen
}

func (v *SSZVector) Encode(eb *EncodingBuffer, p unsafe.Pointer) {
	if v.elemSSZ.IsFixed() {
		EncodeFixedSeries(v.elemSSZ.Encode, v.length, v.elemMemSize, eb, p)
	} else {
		EncodeVarSeries(v.elemSSZ.Encode, v.length, v.elemMemSize, eb, p)
	}
}

func (v *SSZVector) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	if v.elemSSZ.IsFixed() {
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
