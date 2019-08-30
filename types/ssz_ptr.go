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
	alloc   ptrutil.AllocationFn
}

func NewSSZPtr(factory SSZFactoryFn, typ reflect.Type) (*SSZPtr, error) {
	if typ.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("typ is not a pointer")
	}
	elemTyp := typ.Elem()
	elemSSZ, err := factory(elemTyp)
	if err != nil {
		return nil, err
	}
	alloc := func(p unsafe.Pointer) unsafe.Pointer {
		return ptrutil.AllocateSpace(p, elemTyp)
	}
	return &SSZPtr{elemSSZ: elemSSZ, alloc: alloc}, nil
}

func (v *SSZPtr) FuzzMinLen() uint64 {
	return v.elemSSZ.FuzzMinLen()
}

func (v *SSZPtr) FuzzMaxLen() uint64 {
	return v.elemSSZ.FuzzMaxLen()
}

func (v *SSZPtr) MinLen() uint64 {
	return v.elemSSZ.MinLen()
}

func (v *SSZPtr) MaxLen() uint64 {
	return v.elemSSZ.MaxLen()
}

func (v *SSZPtr) FixedLen() uint64 {
	return v.elemSSZ.FixedLen()
}

func (v *SSZPtr) IsFixed() bool {
	return v.elemSSZ.IsFixed()
}

func (v *SSZPtr) SizeOf(p unsafe.Pointer) uint64 {
	innerPtr := unsafe.Pointer(*(*uintptr)(p))
	return v.elemSSZ.SizeOf(innerPtr)
}

func (v *SSZPtr) Encode(eb *EncodingWriter, p unsafe.Pointer) error {
	innerPtr := unsafe.Pointer(*(*uintptr)(p))
	return v.elemSSZ.Encode(eb, innerPtr)
}

func (v *SSZPtr) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	contentsPtr := v.alloc(p)
	return v.elemSSZ.Decode(dr, contentsPtr)
}

func (v *SSZPtr) HashTreeRoot(h HashFn, p unsafe.Pointer) [32]byte {
	innerPtr := unsafe.Pointer(*(*uintptr)(p))
	return v.elemSSZ.HashTreeRoot(h, innerPtr)
}
