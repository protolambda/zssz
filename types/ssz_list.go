package types

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/merkle"
	"github.com/protolambda/zssz/util/ptrutil"
	"reflect"
	"unsafe"
)

type SSZList struct {
	alloc       ptrutil.SliceAllocationFn
	elemMemSize uintptr
	elemSSZ     SSZ
	limit       uint64
}

func NewSSZList(factory SSZFactoryFn, typ reflect.Type) (*SSZList, error) {
	if typ.Kind() != reflect.Slice {
		return nil, fmt.Errorf("typ is not a dynamic-length array")
	}
	limit, err := ReadListLimit(typ)
	if err != nil {
		return nil, err
	}

	elemTyp := typ.Elem()

	elemSSZ, err := factory(elemTyp)
	if err != nil {
		return nil, err
	}
	res := &SSZList{
		alloc:       ptrutil.MakeSliceAllocFn(typ),
		elemMemSize: elemTyp.Size(),
		elemSSZ:     elemSSZ,
		limit:       limit,
	}
	return res, nil
}

func (v *SSZList) FuzzReqLen() uint64 {
	return 8
}

func (v *SSZList) MinLen() uint64 {
	return 0
}

func (v *SSZList) FixedLen() uint64 {
	return 0
}

func (v *SSZList) IsFixed() bool {
	return false
}

func (v *SSZList) Encode(eb *EncodingBuffer, p unsafe.Pointer) {
	sh := ptrutil.ReadSliceHeader(p)
	if v.elemSSZ.IsFixed() {
		EncodeFixedSeries(v.elemSSZ.Encode, uint64(sh.Len), v.elemMemSize, eb, sh.Data)
	} else {
		EncodeVarSeries(v.elemSSZ.Encode, uint64(sh.Len), v.elemMemSize, eb, sh.Data)
	}
}

func (v *SSZList) decodeFuzzmode(dr *DecodingReader, p unsafe.Pointer) error {
	x, err := dr.ReadUint64()
	if err != nil {
		return err
	}
	span := dr.GetBytesSpan()
	length := uint64(0)
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

func (v *SSZList) decode(dr *DecodingReader, p unsafe.Pointer) error {
	bytesLen := dr.Max() - dr.Index()
	if v.elemSSZ.IsFixed() {
		return DecodeFixedSlice(v.elemSSZ.Decode, v.elemSSZ.FixedLen(), bytesLen, v.limit, v.alloc, v.elemMemSize, dr, p)
	} else {
		// still pass the fixed length of the element, but just to check a minimum length requirement.
		return DecodeVarSlice(v.elemSSZ.Decode, v.elemSSZ.FixedLen(), bytesLen, v.limit, v.alloc, v.elemMemSize, dr, p)
	}
}

func (v *SSZList) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	if dr.IsFuzzMode() {
		return v.decodeFuzzmode(dr, p)
	} else {
		return v.decode(dr, p)
	}
}

func (v *SSZList) HashTreeRoot(h HashFn, p unsafe.Pointer) [32]byte {
	elemHtr := v.elemSSZ.HashTreeRoot
	elemSize := v.elemMemSize
	sh := ptrutil.ReadSliceHeader(p)
	leaf := func(i uint64) []byte {
		r := elemHtr(h, unsafe.Pointer(uintptr(sh.Data)+(elemSize*uintptr(i))))
		return r[:]
	}
	return h.MixIn(merkle.Merkleize(h, uint64(sh.Len), v.limit, leaf), uint64(sh.Len))
}
