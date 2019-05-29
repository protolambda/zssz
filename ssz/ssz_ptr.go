package ssz

import (
	"fmt"
	"reflect"
	"unsafe"
)

// proxies SSZ behavior to the SSZ type of the object being pointed to.
type SSZPtr struct {
	elemSSZ SSZ
}

func NewSSZPtr(typ reflect.Type) (*SSZPtr, error) {
	if typ.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("typ is not a pointer")
	}
	elemSSZ, err := sszFactory(typ.Elem())
	if err != nil {
		return nil, err
	}
	return &SSZPtr{elemSSZ: elemSSZ}, nil
}

func (v *SSZPtr) FixedLen() uint32 {
	return v.elemSSZ.FixedLen()
}

func (v *SSZPtr) IsFixed() bool {
	return v.elemSSZ.IsFixed()
}

func (v *SSZPtr) Encode(eb *sszEncBuf, p unsafe.Pointer) {
	innerPtr := unsafe.Pointer(*(*uintptr)(p))
	v.elemSSZ.Encode(eb, innerPtr)
}

func (v *SSZPtr) Decode(dr *SSZDecReader, p unsafe.Pointer) error {
	innerPtr := unsafe.Pointer(*(*uintptr)(p))
	return v.elemSSZ.Decode(dr, innerPtr)
}

func (v *SSZPtr) HashTreeRoot(hFn HashFn, pointer unsafe.Pointer) []byte {

}