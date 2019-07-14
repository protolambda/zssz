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

type SSZBytes struct {
	limit uint64
}

func NewSSZBytes(typ reflect.Type) (*SSZBytes, error) {
	if typ.Kind() != reflect.Slice {
		return nil, fmt.Errorf("typ is not a dynamic-length bytes slice")
	}
	if typ.Elem().Kind() != reflect.Uint8 {
		return nil, fmt.Errorf("typ is not a bytes slice")
	}
	limit, err := ReadListLimit(typ)
	if err != nil {
		return nil, err
	}
	return &SSZBytes{limit: limit}, nil
}

func (v *SSZBytes) FuzzReqLen() uint64 {
	return 8
}

func (v *SSZBytes) FixedLen() uint64 {
	return 0
}

func (v *SSZBytes) MinLen() uint64 {
	return 0
}

func (v *SSZBytes) IsFixed() bool {
	return false
}

func (v *SSZBytes) Encode(eb *EncodingBuffer, p unsafe.Pointer) {
	sh := ptrutil.ReadSliceHeader(p)
	data := *(*[]byte)(unsafe.Pointer(sh))
	eb.Write(data)
}

func (v *SSZBytes) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	var length uint64
	if dr.IsFuzzMode() {
		x, err := dr.ReadUint64()
		if err != nil {
			return err
		}
		span := dr.GetBytesSpan()
		if span != 0 {
			length = x % span
		}
	} else {
		length = dr.Max() - dr.Index()
	}
	if length > v.limit {
		return fmt.Errorf("got %d bytes, expected no more than %d bytes", length, v.limit)
	}
	ptrutil.BytesAllocFn(p, length)
	data := *(*[]byte)(p)
	_, err := dr.Read(data)
	return err
}

func (v *SSZBytes) HashTreeRoot(h HashFn, p unsafe.Pointer) [32]byte {
	sh := ptrutil.ReadSliceHeader(p)
	data := *(*[]byte)(unsafe.Pointer(sh))
	dataLen := uint64(len(data))
	leafCount := (dataLen + 31) >> 5
	leafLimit := (v.limit + 31) >> 5
	leaf := func(i uint64) []byte {
		s := i << 5
		e := (i + 1) << 5
		// pad the data
		if e > dataLen {
			x := [32]byte{}
			copy(x[:], data[s:dataLen])
			return x[:]
		}
		return data[s:e]
	}
	return h.MixIn(merkle.Merkleize(h, leafCount, leafLimit, leaf), dataLen)
}
