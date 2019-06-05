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

type SSZList struct {
	alloc    ptrutil.SliceAllocationFn
	elemMemSize uintptr
	elemSSZ     SSZ
}

func NewSSZList(factory SSZFactoryFn, typ reflect.Type) (*SSZList, error) {
	if typ.Kind() != reflect.Slice {
		return nil, fmt.Errorf("typ is not a dynamic-length array")
	}
	elemTyp := typ.Elem()

	elemSSZ, err := factory(elemTyp)
	if err != nil {
		return nil, err
	}
	res := &SSZList{
		alloc: ptrutil.MakeSliceAllocFn(typ),
		elemMemSize: elemTyp.Size(),
		elemSSZ:     elemSSZ,
	}
	return res, nil
}

func (v *SSZList) FuzzReqLen() uint32 {
	return 4
}

func (v *SSZList) MinLen() uint32 {
	return 0
}

func (v *SSZList) FixedLen() uint32 {
	return 0
}

func (v *SSZList) IsFixed() bool {
	return false
}

func (v *SSZList) Encode(eb *EncodingBuffer, p unsafe.Pointer) {
	sh := ptrutil.ReadSliceHeader(p)
	if v.elemSSZ.IsFixed() {
		EncodeFixedSeries(v.elemSSZ.Encode, uint32(sh.Len), v.elemMemSize, eb, sh.Data)
	} else {
		EncodeVarSeries(v.elemSSZ.Encode, uint32(sh.Len), v.elemMemSize, eb, sh.Data)
	}
}

func (v *SSZList) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	if dr.IsFuzzMode() {
		x, err := dr.ReadUint32()
		if err != nil {
			return err
		}
		span := dr.GetBytesSpan()
		length := uint32(0)
		if span != 0 {
			length = (x % span) / v.elemSSZ.FuzzReqLen()
		}
		if !v.elemSSZ.IsFixed() {
			length /= 10
		}
		contentsPtr := v.alloc(p, length)
		if v.elemSSZ.IsFixed() {
			return DecodeFixedSeries(v.elemSSZ.Decode, length, v.elemMemSize, dr, contentsPtr)
		} else {
			return DecodeVarSeriesFuzzMode(v.elemSSZ, length, v.elemMemSize, dr, contentsPtr)
		}
	}
	bytesLen := dr.Max() - dr.Index()
	if v.elemSSZ.IsFixed() {
		return DecodeFixedSlice(v.elemSSZ.Decode, v.elemSSZ.FixedLen(), bytesLen, v.alloc, v.elemMemSize, dr, p)
	} else {
		// still pass the fixed length of the element, but just to check a minimum length requirement.
		return DecodeVarSlice(v.elemSSZ.Decode, v.elemSSZ.FixedLen(), bytesLen, v.alloc, v.elemMemSize, dr, p)
	}
}

func (v *SSZList) HashTreeRoot(h *Hasher, p unsafe.Pointer) [32]byte {
	elemHtr := v.elemSSZ.HashTreeRoot
	elemSize := v.elemMemSize
	sh := ptrutil.ReadSliceHeader(p)
	leaf := func(i uint32) []byte {
		r := elemHtr(h, unsafe.Pointer(uintptr(sh.Data)+(elemSize*uintptr(i))))
		return r[:]
	}
	return h.MixIn(Merkleize(h, uint32(sh.Len), leaf), uint32(sh.Len))
}
