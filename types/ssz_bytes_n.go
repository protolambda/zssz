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

type SSZBytesN struct {
	length uint32
}

func NewSSZBytesN(typ reflect.Type) (*SSZBytesN, error) {
	if typ.Kind() != reflect.Array {
		return nil, fmt.Errorf("typ is not a fixed-length bytes array")
	}
	if typ.Elem().Kind() != reflect.Uint8 {
		return nil, fmt.Errorf("typ is not a bytes array")
	}
	length := typ.Len()
	res := &SSZBytesN{length: uint32(length)}
	return res, nil
}

func (v *SSZBytesN) FuzzReqLen() uint32 {
	return v.length
}

func (v *SSZBytesN) FixedLen() uint32 {
	// 1 byte per element, just the same as the length
	return v.length
}

func (v *SSZBytesN) MinLen() uint32 {
	// 1 byte per element, just the same as the length
	return v.length
}

func (v *SSZBytesN) IsFixed() bool {
	return true
}

func (v *SSZBytesN) Encode(eb *EncodingBuffer, p unsafe.Pointer) {
	sh := ptrutil.GetSliceHeader(p, v.length)
	data := *(*[]byte)(unsafe.Pointer(sh))
	eb.Write(data)
}

func (v *SSZBytesN) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	sh := ptrutil.GetSliceHeader(p, v.length)
	data := *(*[]byte)(unsafe.Pointer(sh))
	_, err := dr.Read(data)
	return err
}

func (v *SSZBytesN) HashTreeRoot(h HashFn, p unsafe.Pointer) [32]byte {
	sh := ptrutil.GetSliceHeader(p, v.length)
	data := *(*[]byte)(unsafe.Pointer(sh))
	leafCount := (v.length + 31) >> 5
	leaf := func(i uint32) []byte {
		s := i << 5
		e := (i + 1) << 5
		// pad the data
		if e > v.length {
			x := [32]byte{}
			copy(x[:], data[s:v.length])
			return x[:]
		}
		return data[s:e]
	}
	return Merkleize(h, leafCount, leaf)
}
