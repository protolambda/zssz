package types

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/util/ptrutil"
	"reflect"
	"unsafe"
)

// proxies SSZ behavior to the SSZ type of the object being pointed to.
type SSZPtr struct {
	elemSSZ SSZ
	elemSize uintptr
}

func NewSSZPtr(factory SSZFactoryFn, typ reflect.Type) (*SSZPtr, error) {
	if typ.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("typ is not a pointer")
	}
	elemTyp := typ.Elem()
	elemSize := elemTyp.Size()
	elemSSZ, err := factory(elemTyp)
	if err != nil {
		return nil, err
	}
	return &SSZPtr{elemSSZ: elemSSZ, elemSize: elemSize}, nil
}

func (v *SSZPtr) MinLen() uint32 {
	return v.elemSSZ.MinLen()
}

func (v *SSZPtr) FixedLen() uint32 {
	return v.elemSSZ.FixedLen()
}

func (v *SSZPtr) IsFixed() bool {
	return v.elemSSZ.IsFixed()
}

func (v *SSZPtr) Encode(eb *EncodingBuffer, p unsafe.Pointer) {
	innerPtr := unsafe.Pointer(*(*uintptr)(p))
	v.elemSSZ.Encode(eb, innerPtr)
}

func (v *SSZPtr) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	contentsPtr := ptrutil.AllocateSpace(p, v.elemSize)
	*(*uintptr)(p) = uintptr(contentsPtr)
	return v.elemSSZ.Decode(dr, contentsPtr)
}

func (v *SSZPtr) HashTreeRoot(h *Hasher, p unsafe.Pointer) [32]byte {
	innerPtr := unsafe.Pointer(*(*uintptr)(p))
	return v.HashTreeRoot(h, innerPtr)
}
