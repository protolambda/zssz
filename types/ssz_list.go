package types

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/merkle"
	. "github.com/protolambda/zssz/pretty"
	"github.com/protolambda/zssz/util/ptrutil"
	"reflect"
	"unsafe"
)

type SSZList struct {
	alloc         ptrutil.SliceAllocationFn
	elemMemSize   uintptr
	elemSSZ       SSZ
	fixedElemSize uint64
	limit         uint64
	byteLimit     uint64
	maxFuzzLen    uint64
}

func NewSSZList(factory SSZFactoryFn, typ reflect.Type) (*SSZList, error) {
	if typ.Kind() != reflect.Slice {
		return nil, fmt.Errorf("typ %v is not a dynamic-length array", typ)
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
	var fixedElemSize, byteLimit uint64
	if elemSSZ.IsFixed() {
		fixedElemSize = elemSSZ.FixedLen()
		byteLimit = limit * elemSSZ.FixedLen()
	} else {
		fixedElemSize = BYTES_PER_LENGTH_OFFSET
		byteLimit = limit * elemSSZ.MaxLen()
	}
	res := &SSZList{
		alloc:         ptrutil.MakeSliceAllocFn(typ),
		elemMemSize:   elemTyp.Size(),
		elemSSZ:       elemSSZ,
		fixedElemSize: fixedElemSize,
		limit:         limit,
		byteLimit:     byteLimit,
		maxFuzzLen:    8 + (limit * elemSSZ.FuzzMaxLen()),
	}
	return res, nil
}

func (v *SSZList) FuzzMinLen() uint64 {
	return 8
}

func (v *SSZList) FuzzMaxLen() uint64 {
	return v.maxFuzzLen
}

func (v *SSZList) MinLen() uint64 {
	return 0
}

func (v *SSZList) MaxLen() uint64 {
	return v.byteLimit
}

func (v *SSZList) FixedLen() uint64 {
	return 0
}

func (v *SSZList) IsFixed() bool {
	return false
}

func (v *SSZList) SizeOf(p unsafe.Pointer) uint64 {
	sh := ptrutil.ReadSliceHeader(p)
	if v.elemSSZ.IsFixed() {
		return uint64(sh.Len) * v.fixedElemSize
	} else {
		out := uint64(sh.Len) * BYTES_PER_LENGTH_OFFSET
		memOffset := uintptr(0)
		for i := 0; i < sh.Len; i++ {
			elemPtr := unsafe.Pointer(uintptr(sh.Data) + memOffset)
			memOffset += v.elemMemSize
			out += v.elemSSZ.SizeOf(elemPtr)
		}
		return out
	}
}

func (v *SSZList) Encode(eb *EncodingWriter, p unsafe.Pointer) error {
	sh := ptrutil.ReadSliceHeader(p)
	if v.elemSSZ.IsFixed() {
		return EncodeFixedSeries(v.elemSSZ.Encode, uint64(sh.Len), v.elemMemSize, eb, sh.Data)
	} else {
		return EncodeVarSeries(v.elemSSZ.Encode, v.elemSSZ.SizeOf, uint64(sh.Len), v.elemMemSize, eb, sh.Data)
	}
}

func (v *SSZList) decodeFuzzmode(dr *DecodingReader, p unsafe.Pointer) error {
	x, err := dr.ReadUint64()
	if err != nil {
		return err
	}
	span := dr.GetBytesSpan()
	if span > v.maxFuzzLen-8 {
		span = v.maxFuzzLen - 8
	}
	length := uint64(0)
	if span != 0 {
		length = (x % span) / v.elemSSZ.FuzzMinLen()
	}
	contentsPtr := v.alloc(p, length)
	if v.elemSSZ.IsFixed() {
		return DecodeFixedSeries(v.elemSSZ.Decode, length, v.elemMemSize, dr, contentsPtr)
	} else {
		return DecodeVarSeriesFuzzMode(v.elemSSZ, length, v.elemMemSize, dr, contentsPtr)
	}
}

func (v *SSZList) decode(dr *DecodingReader, p unsafe.Pointer) error {
	if v.elemSSZ.IsFixed() {
		return DecodeFixedSlice(v.elemSSZ.Decode, v.elemSSZ.FixedLen(), dr.GetBytesSpan(), v.limit, v.alloc, v.elemMemSize, dr, p)
	} else {
		// still pass the fixed length of the element, but just to check a minimum length requirement.
		return DecodeVarSlice(v.elemSSZ.Decode, v.elemSSZ.FixedLen(), dr.GetBytesSpan(), v.limit, v.alloc, v.elemMemSize, dr, p)
	}
}

func (v *SSZList) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	if dr.IsFuzzMode() {
		return v.decodeFuzzmode(dr, p)
	} else {
		return v.decode(dr, p)
	}
}

func (v *SSZList) DryCheck(dr *DecodingReader) error {
	if v.elemSSZ.IsFixed() {
		return DryCheckFixedSlice(v.elemSSZ.DryCheck, v.elemSSZ.FixedLen(), dr.GetBytesSpan(), v.limit, dr)
	} else {
		return DryCheckVarSlice(v.elemSSZ.DryCheck, v.elemSSZ.FixedLen(), dr.GetBytesSpan(), v.limit, dr)
	}
}

func (v *SSZList) HashTreeRoot(h Hasher, p unsafe.Pointer) [32]byte {
	elemHtr := v.elemSSZ.HashTreeRoot
	elemSize := v.elemMemSize
	sh := ptrutil.ReadSliceHeader(p)
	leaf := func(i uint64) []byte {
		r := elemHtr(h, unsafe.Pointer(uintptr(sh.Data)+(elemSize*uintptr(i))))
		return r[:]
	}
	return h.MixIn(merkle.Merkleize(h, uint64(sh.Len), v.limit, leaf), uint64(sh.Len))
}

func (v *SSZList) Pretty(indent uint32, w *PrettyWriter, p unsafe.Pointer) {
	sh := ptrutil.ReadSliceHeader(p)
	length := uint64(sh.Len)
	w.WriteIndent(indent)
	w.Write("[\n")
	CallSeries(func(i uint64, p unsafe.Pointer) {
		v.elemSSZ.Pretty(indent+1, w, p)
		if i == length-1 {
			w.Write("\n")
		} else {
			w.Write(",\n")
		}
	}, length, v.elemMemSize, sh.Data)
	w.WriteIndent(indent)
	w.Write("]")
}
