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
	offset := uintptr(0)
	if v.elemSSZ.IsFixed() {
		for i := uint32(0); i < v.length; i++ {
			elemPtr := unsafe.Pointer(uintptr(p) + offset)
			offset += v.elemMemSize
			v.elemSSZ.Encode(eb, elemPtr)
		}
	} else {
		for i := uint32(0); i < v.length; i++ {
			elemPtr := unsafe.Pointer(uintptr(p) + offset)
			offset += v.elemMemSize
			// write an offset to the fixed data, to find the dynamic data with as a reader
			eb.WriteOffset(v.fixedLen)

			// encode the dynamic data to a temporary buffer
			temp := getPooledBuffer()
			v.elemSSZ.Encode(temp, elemPtr)
			// write it forward
			eb.WriteForward(temp.Bytes())

			releasePooledBuffer(temp)
		}
		eb.FlushForward()
	}
}

func (v *SSZVector) Decode(p unsafe.Pointer) {
	// TODO
}
func (v *SSZVector) Ignore() {
	// TODO skip ahead Length bytes in input
}
