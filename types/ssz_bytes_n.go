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

type SSZBytesN struct {
	length uint64
}

func NewSSZBytesN(typ reflect.Type) (*SSZBytesN, error) {
	if typ.Kind() != reflect.Array {
		return nil, fmt.Errorf("typ is not a fixed-length bytes array")
	}
	if typ.Elem().Kind() != reflect.Uint8 {
		return nil, fmt.Errorf("typ is not a bytes array")
	}
	length := typ.Len()
	res := &SSZBytesN{length: uint64(length)}
	return res, nil
}

func (v *SSZBytesN) FuzzMinLen() uint64 {
	return v.length
}

func (v *SSZBytesN) FuzzMaxLen() uint64 {
	return v.length
}

func (v *SSZBytesN) MinLen() uint64 {
	return v.length
}

func (v *SSZBytesN) MaxLen() uint64 {
	return v.length
}

func (v *SSZBytesN) FixedLen() uint64 {
	return v.length
}

func (v *SSZBytesN) IsFixed() bool {
	return true
}

func (v *SSZBytesN) SizeOf(p unsafe.Pointer) uint64 {
	return v.length
}

func (v *SSZBytesN) Encode(eb *EncodingWriter, p unsafe.Pointer) error {
	sh := ptrutil.GetSliceHeader(p, v.length)
	data := *(*[]byte)(unsafe.Pointer(sh))
	return eb.Write(data)
}

func (v *SSZBytesN) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	sh := ptrutil.GetSliceHeader(p, v.length)
	data := *(*[]byte)(unsafe.Pointer(sh))
	_, err := dr.Read(data)
	return err
}

func (v *SSZBytesN) Verify(dr *DecodingReader) error {
	_, err := dr.Skip(v.length)
	return err
}

func (v *SSZBytesN) HashTreeRoot(h HashFn, p unsafe.Pointer) [32]byte {
	sh := ptrutil.GetSliceHeader(p, v.length)
	data := *(*[]byte)(unsafe.Pointer(sh))
	leafCount := (v.length + 31) >> 5
	leaf := func(i uint64) []byte {
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
	return merkle.Merkleize(h, leafCount, leafCount, leaf)
}

func (v *SSZBytesN) Pretty(indent uint32, w *PrettyWriter, p unsafe.Pointer) {
	w.WriteIndent(indent)
	sh := ptrutil.GetSliceHeader(p, v.length)
	data := *(*[]byte)(unsafe.Pointer(sh))
	w.Write(fmt.Sprintf("0x%x", data))
}
