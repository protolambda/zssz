package ssz

import (
	"fmt"
	"reflect"
	"unsafe"
)

type VectorLength interface {
	VectorLength() uint32
}

type SSZVector struct {
	length uint32
	elemMemSize uintptr
	elemSSZ SSZ
	isFixedLen bool
	fixedLen uint32
}

func NewSSZVector(typ reflect.Type) (*SSZVector, error) {
	if typ.Kind() != reflect.Array {
		return nil, fmt.Errorf("typ is not a fixed-length array")
	}
	length := uint32(typ.Len())
	elemTyp := typ.Elem()

	elemSSZ, err := sszFactory(elemTyp)
	if err != nil {
		return nil, err
	}
	res := &SSZVector{
		length: length,
		elemMemSize: elemTyp.Size(),
		elemSSZ: elemSSZ,
		isFixedLen: elemSSZ.IsFixed(),
		fixedLen: elemSSZ.FixedLen() * length,
	}
	return res, nil
}

func (v *SSZVector) VectorLength() uint32 {
	return v.length
}

func (v *SSZVector) FixedLen() uint32 {
	return v.fixedLen
}

func (v *SSZVector) IsFixed() bool {
	return v.isFixedLen
}

func (v *SSZVector) Encode(eb *sszEncBuf, p unsafe.Pointer) {
	EncodeSeries(v.elemSSZ, v.length, v.elemMemSize, eb, p)
}

func (v *SSZVector)  Decode(dr *SSZDecReader, p unsafe.Pointer) error {
	return DecodeSeries(v.elemSSZ, v.length, v.elemMemSize, dr, p)
}

func (v *SSZVector) Ignore() {
	// TODO skip ahead Length bytes in input
}
